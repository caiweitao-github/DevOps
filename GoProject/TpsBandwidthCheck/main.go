package main

import (
	"Aliyun"
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
)

type TPS interface {
	GetTunnelIfo(tid string) (tunnelData, error)
	SetExpire(code string) (e error)
	GetKey(code string) (resu string, e error)
	Ssh()
	CheckBandwidth()
}

type TpsInfo struct {
	id                                           int
	code, ip                                     string
	realtime_bandwidth, avg_bandwidth, bandwidth int
}

type tunnelData struct {
	bandwidth_max, host, host2 string
}

var rdb *redis.Client
var Db *sql.DB
var ctx = context.Background()
var url = ""
var title = "[TPS域名切换通知]"
var tpsBakCode = "tps79"

var (
	logFile  = "/data/kdl/log/devops/tpsDomainAutoSwitch.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

func main() {
	defer func(pre time.Time) {
		elapsed := time.Since(pre)
		logInfo.Printf("elapsed: %v", elapsed)
	}(time.Now())
	logInfo.Println("run..")
	GetTps()
}

func init() {
	var err error
	Db, err = util.dbDB()
	if err != nil {
		panic(err)
	}
	rdb, err = util.RedisDB()
	if err != nil {
		panic(err)
	}
}

func GetTps() {
	sqlStr := "select id,code,login_ip,format(realtime_bandwidth/1024/128,0) as realtime_bandwidth,format(avg_bandwidth/1024/128,0) as avg_bandwidth,bandwidth from tps where status = 1 and code not in ('tps56', 'tps78', 'tps79', 'tps42')"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logError.Panicf("get tps err: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tps TpsInfo
		err := rows.Scan(&tps.id, &tps.code, &tps.ip, &tps.realtime_bandwidth, &tps.avg_bandwidth, &tps.bandwidth)
		if err != nil {
			logError.Panicf("scan failed, err:%v\n", err)
		}
		var TPS TPS = tps
		TPS.CheckBandwidth()
	}
}

func GetJdeIP() (string, error) {
	ipList := []string{}
	sqlStr := "select login_ip from jde_server where status = '1'"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var ip string
		err := rows.Scan(&ip)
		if err != nil {
			return "", err
		}
		ipList = append(ipList, ip)
	}
	return strings.Join(ipList, "\\|"), nil
}

func (t TpsInfo) SetKey(code string) (e error) {
	if r, err := rdb.SetNX(ctx, code, 1, time.Second*420).Result(); !r && err == nil {
		if _, err := rdb.Incr(ctx, code).Result(); err != nil {
			e = err
		}
	}
	return
}

func (t TpsInfo) GetKey(code string) (resu string, er error) {
	if res, err := rdb.Get(ctx, code).Result(); err == redis.Nil {
		resu, er = "", err
	} else {
		resu, er = res, nil
	}
	return
}

func (t TpsInfo) SetExpire(code string) (e error) {
	if _, err := rdb.Expire(ctx, code, time.Second*1800).Result(); err == nil {
		e = nil
	} else {
		e = err
	}
	return
}

func (t TpsInfo) Ssh() {
	tidinfo := []string{}
	str := []string{}
	str = append(str, fmt.Sprintf("%s带宽过载, 实时带宽: %dM, 平均带宽: %dM, 机器带宽: %dM\n已自动处理:", t.code, t.realtime_bandwidth, t.avg_bandwidth, t.bandwidth))
	var Cmd util.SshUtil = util.SshInfo{Host: t.ip, Port: "8022"}
	SshCfg, err := Cmd.InitSsh()
	if err != nil {
		logError.Fatal(err)
	}
	jdeIP, err := GetJdeIP()
	if err != nil {
		logError.Fatal(err)
	}
	cmd := []string{fmt.Sprintf("iftop -t -s 3 -n -F 'IPs' 2>/dev/null | sed -n '4,24p'|awk '{print $1}'|grep -Eo '([0-9]{1,3}\\.){3}[0-9]{1,3}'|grep -v '%s'|head -n 1", jdeIP)}
	res, _err := Cmd.Conn(SshCfg, cmd)
	if _err != nil {
		logError.Println(_err)
	}
	addresses := strings.Split(res, "\n")
	for _, address := range addresses {
		if address == "" {
			continue
		}
		cmd2 := []string{fmt.Sprintf("tail -n50000 /data/log/tps/proxy.log|grep '%s'|grep 'INFO'|awk '{print $5}'|grep -v '-'|awk -F '|' '{print $4}'|sort -rn|uniq|head -n 1", address)}
		tidlist, err := Cmd.Conn(SshCfg, cmd2)
		if err != nil {
			logError.Printf("Get Tid Err: %v", err)
		} else {
			tid := strings.Split(tidlist, "\n")
			for _, v := range tid {
				if v == "" {
					continue
				}
				td, err := t.GetTunnelIfo(v)
				if err != nil {
					logError.Printf("Failed to get info from DB: %v", err)
				} else {
					id, bakIp, err := getBakIP(tpsBakCode)
					if err != nil {
						logError.Fatalf("get bakip err: %v", err)
						continue
					} else {
						updateDomain(td.host, bakIp)
						updateDomain(td.host2, bakIp)
						insertData(td.host, td.host2, t.id)
						updateDB(id, td.host)
					}
					tidinfo = append(tidinfo, fmt.Sprintf("Tid: %s, 订单带宽: %sM, 切换详情: %s --> %s(%s)", v, td.bandwidth_max, td.host, bakIp, tpsBakCode))
				}
			}
			str = append(str, fmt.Sprintf("IP: %s\n\t%s", address, strings.Join(tidinfo, "\n\t")))
			tidinfo = []string{}
		}
	}
	util.FeiShuNotify(url, title, str)
}

func (t TpsInfo) GetTunnelIfo(tid string) (tunnelData, error) {
	sqlStr := "select bandwidth_max,host,host2 from tunnel where tid = ?"
	var tunnel tunnelData
	rows := Db.QueryRow(sqlStr, tid).Scan(&tunnel.bandwidth_max, &tunnel.host, &tunnel.host2)
	if rows != nil {
		return tunnel, fmt.Errorf("query failed: %v", rows)
	} else {
		return tunnel, nil
	}
}

func (t TpsInfo) CheckBandwidth() {
	if t.avg_bandwidth > int(float64(t.bandwidth)*0.8) {
		err := t.SetKey(t.code)
		if err != nil {
			logError.Panicln(err)
		}
		if num, err := t.GetKey(t.code); err != nil {
			logError.Panicln(err)
		} else {
			if threshold, err := strconv.Atoi(num); err != nil {
				logError.Print(err)
			} else {
				if threshold == 3 {
					if err := t.SetExpire(t.code); err != nil {
						logError.Panicln(err)
					} else {
						t.Ssh()
					}
				}
			}
		}
	}
}

func getBakIP(code string) (int, string, error) {
	var id int
	var ip string
	sqlStr := `select id,login_ip from tps where code = ?`
	rows := Db.QueryRow(sqlStr, code).Scan(&id, &ip)
	if rows != nil {
		return 0, "", fmt.Errorf("query failed: %v", rows)
	} else {
		return id, ip, nil
	}

}

func updateDomain(domain, ip string) {
	re := regexp.MustCompile(`\.`)
	data := re.Split(domain, 2)
	err := Aliyun.UpdateTpsDomain(data[0], data[1], "A", ip, 600)
	if err != nil {
		logError.Printf("Update %s Domain Failed: %s", domain, err)
	} else {
		logInfo.Printf("Updata Domain %s ------> %s.", domain, ip)
	}
}

func updateDB(id int, domain string) {
	sqlStr := "update tps_domain set tps_id = ? where domain = ? or domain2 = ?"
	Db.Exec(sqlStr, id, domain, domain)
}

func insertData(srcDomain, srcDomain2 string, srcId int) {
	sqlStr := `insert into tps_domain_change(src_domain, src_tps_id, dest_tps_id, status, update_time, create_time, memo)
	values (?, ?, 60, 1, now(), now(), '[added by TpsBandwidthCheck]'), (?, ?, 60, 1, now(), now(), '[added by TpsBandwidthCheck]')`
	_, err := Db.Exec(sqlStr, srcDomain, srcId, srcDomain2, srcId)
	if err != nil {
		logError.Printf("insert data err %v", err)
		return
	}
}
