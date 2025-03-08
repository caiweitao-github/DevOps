package main

import (
	"Aliyun"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"util"
)

var (
	logFile  = "/data/kdl/log/devops/jlyDomainCheck.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
	prefix   = "tn"
	mu       sync.Mutex
	Db       *sql.DB
)

const (
	tpsHA      = 31
	domaiNum   = 3
	domain     = ""
	domainLine = "default"
)

type tps struct {
	id            int
	code          string
	ip            string
	num           int
	avg_bandwidth int
	bandwidth     int
}

type tpsOrder struct {
	domain string
	num    int
}

type tpsDomainInfo struct {
	domain        string
	avg_bandwidth int
	bandwidth     int
}

type tpsBandwidth struct {
	id            int
	code          string
	avg_bandwidth int
	bandwidth     int
}

func main() {
	start := time.Now()
	logInfo.Println("Run jlyDomainCheck ...")
	ChangeTpsDomainStatus()
	CheckTpsBandwidth()
	CheckTps()
	exceptionTps := SyncDomainRemark()
	if len(exceptionTps) > 0 {
		util.SendMess2(exceptionTps, "[域名解析异常]")
	}
	end := time.Now()
	logInfo.Printf("Time Use: %v", end.Sub(start))
}

func init() {
	var err error
	Db, err = util.JlyDB()
	if err != nil {
		panic(err)
	}
}

func CheckTpsDomainStatus() ([]tpsOrder, error) {
	sqlStr := "select domain,allocated_order_count from tps_domain where status = 1 and allocated_order_count >= 1"
	domainList := make([]tpsOrder, 0)
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tps tpsOrder
		err := rows.Scan(&tps.domain, &tps.num)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		domainList = append(domainList, tps)
	}
	return domainList, err
}

func ChangeTpsDomainStatus() {
	domainList := make([]tpsDomainInfo, 0)
	tpsDomainList, err := CheckTpsDomainStatus()
	if err != nil {
		logError.Printf("get tps doamin list failed: %s", err)
		return
	}
	if len(tpsDomainList) > 0 {
		for _, v := range tpsDomainList {
			_, err := Db.Exec("update tps_domain set status = 2 where domain = ?", v.domain)
			if err != nil {
				logError.Printf("update tps_domain failed: %v", err)
				return
			} else {
				logInfo.Printf("Change Domain %s Status ------> Pause.", v.domain)
			}
		}
	}
	sqlStr := `select tps_domain.domain,format(tps.avg_bandwidth/1024/128,0) as avg_bandwidth,tps.bandwidth from
	tps_domain,tps where tps.status = 1 and tps_domain.tps_id=tps.id and tps_domain.use_type = 1 and tps_domain.status = 2 and tps_domain.allocated_order_count = 0`
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logError.Printf("query tps doamin failed: %s", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var domainInfo tpsDomainInfo
		err := rows.Scan(&domainInfo.domain, &domainInfo.avg_bandwidth, &domainInfo.bandwidth)
		if err != nil {
			logError.Printf("query tps_domain failed: %v", err)
			return
		}
		domainList = append(domainList, domainInfo)
	}

	if len(domainList) > 0 {
		for _, v := range domainList {
			if v.avg_bandwidth < int(float64(v.bandwidth)*0.8) {
				_, err := Db.Exec("update tps_domain set status = 1 where domain = ?", v.domain)
				if err != nil {
					logError.Printf("update tps_domain failed: %v", err)
					return
				} else {
					logInfo.Printf("Change Domain %s Status ------> Normal.", v.domain)
				}
			}
		}
	}
}

func CheckTpsBandwidth() {
	TpsList := make([]tpsBandwidth, 0)
	sqlStr1 := "select id,code,format(tps.avg_bandwidth/1024/128,0) as avg_bandwidth,tps.bandwidth from tps where status = 1 and id != ?"
	sqlStr2 := "update tps_domain set status = 2 where tps_id = ?"
	rows, err := Db.Query(sqlStr1, tpsHA)
	if err != nil {
		logError.Printf("query failed: %s", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tps tpsBandwidth
		err := rows.Scan(&tps.id, &tps.code, &tps.avg_bandwidth, &tps.bandwidth)
		if err != nil {
			logError.Printf("query tps_domain failed: %v", err)
			return
		}
		TpsList = append(TpsList, tps)
	}

	if len(TpsList) > 0 {
		for _, v := range TpsList {
			if v.avg_bandwidth > int(float64(v.bandwidth)*0.8) {
				_, err := Db.Exec(sqlStr2, v.id)
				if err != nil {
					logError.Printf("update tps_domain failed: %v", err)
					return
				}
				logInfo.Printf("%s当前平均带宽超总带宽80%%, 机器下的所有域名已全部设为暂停分配!", v.code)
			}
		}
	}
}

func CheckTps() {
	mess := []string{}
	TpsList := make([]tps, 0)
	sqlStr := `select ANY_VALUE(tps.id) as id,ANY_VALUE(tps.code) as code,tps.login_ip,count(ANY_VALUE(tps_domain.tps_id)) as num,
    format(ANY_VALUE(tps.avg_bandwidth)/1024/128,0) as avg_bandwidth,ANY_VALUE(tps.bandwidth) as bandwidth from tps 
    left join tps_domain on tps.id = tps_domain.tps_id and tps_domain.status = 1 where tps.code not REGEXP 'tpsb.*|tpstest|tpsysy|tps56' 
	and tps_domain.domain like '%jiliuip.com' and tps.status = 1 group by tps.login_ip`
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logError.Printf("query failed: %s", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tps tps
		err := rows.Scan(&tps.id, &tps.code, &tps.ip, &tps.num, &tps.avg_bandwidth, &tps.bandwidth)
		if err != nil {
			logError.Printf("query tps_domain failed: %v", err)
			return
		}
		TpsList = append(TpsList, tps)
	}
	for _, v := range TpsList {
		if v.num < domaiNum {
			if v.avg_bandwidth < int(float64(v.bandwidth)*0.8) {
				str, err := CreateDomain(v.id, v.ip, v.code)
				if err != nil {
					logError.Printf("Create Domain Err: %s", err)
				} else {
					mess = append(mess, str...)
				}
			} else {
				logInfo.Printf("%s机器域名数量低于预设值, 但是机器带宽过高!", v.code)
			}
		}
	}
	if len(mess) > 0 {
		util.SendMess2(mess, "[新建Jly域名解析通知]")
	}
}

func getMaxDomain() (maxDomainNum int, e error) {
	var s string
	sqlStr := "select domain2 from tps_domain order by id desc limit 1"
	err := Db.QueryRow(sqlStr).Scan(&s)
	if err != nil {
		e = err
		return
	}

	if strings.HasPrefix(s, "tn") {
		numPart := strings.TrimPrefix(s, "tn")
		parts := strings.SplitN(numPart, ".", 2)
		num, err := strconv.Atoi(parts[0])
		if err != nil {
			e = err
			return
		}
		maxDomainNum = num + 1
	}
	return
}

func CreateDomain(id int, ip string, code string) ([]string, error) {
	str := []string{}
	sqlStr := `insert into tps_domain(domain, domain2, tps_id, status, use_type, create_time) values(?, ?, ?, 1, 1, CURRENT_TIMESTAMP())`
	i := 1
	for i <= 5 {
		suffix, err3 := getMaxDomain()
		if err3 != nil {
			panic(err3)
		}
		domainList := []string{fmt.Sprintf("%s%d", prefix, suffix), fmt.Sprintf("%s%d", prefix, suffix+1)}
		_, err := Aliyun.QueryDomainRecord(domainList[0], domain)
		_, err2 := Aliyun.QueryDomainRecord(domainList[1], domain)
		if err == nil || err2 == nil {
			logError.Printf("%s or %s Domain Already Exist, Skip!", domainList[0], domainList[1])
			break
		} else {
			for _, v := range domainList {
				err := Aliyun.AddDomainRecord(v, domain, "A", ip, domainLine, 600)
				if err != nil {
					logError.Printf("Create %s Domain Failed, %s", v, err)
					return nil, err
				} else {
					str = append(str, fmt.Sprintf("%s.%s ------> %s(%s)", v, domain, code, ip))
					logInfo.Printf("Create %s.%s ------> %s(%s) Success.", v, domain, code, ip)
				}
			}
			i++
			_, err := Db.Exec(sqlStr, domainList[0]+"."+domain, domainList[1]+"."+domain, id)
			if err != nil {
				logError.Printf("Create Domain Success, But Sync DB Failed, %s", err)
			}
		}
	}
	return str, nil
}

func GetTpsCode(ip string) string {
	sqlStr := "select code from tps where login_ip = ?"
	var code string
	err := Db.QueryRow(sqlStr, ip).Scan(&code)
	if err != nil {
		logError.Printf("Get Tps Code Failed: %s", err)
	}
	return code
}

func SyncDomainRemark() []string {
	pool := util.New(15)
	exception := []string{}
	re := regexp.MustCompile(`\.`)
	sqlStr1 := "select id,code,login_ip from tps where status = 1 and id != ?"
	sqlStr2 := "select domain from tps_domain where tps_id = ? and status in (1, 2)"
	res, _err := Db.Query(sqlStr1, tpsHA)
	if _err != nil {
		logError.Printf("query failed: %s", _err)
		return nil
	}
	defer res.Close()
	for res.Next() {
		var id int
		var code string
		var ip string
		err := res.Scan(&id, &code, &ip)
		if err != nil {
			logError.Fatalf("query tps_info failed: %v", err)
		}
		rows, err := Db.Query(sqlStr2, id)
		if err != nil {
			logError.Fatalf("query failed: %s", err)
		}
		defer rows.Close()
		pool.Wg.Add(1)
		pool.NewTask(func() {
			defer pool.Wg.Done()
			for rows.Next() {
				var domainName string
				err := rows.Scan(&domainName)
				if err != nil {
					logError.Fatalf("query tps_domain failed: %v", err)
				}
				parts := re.Split(domainName, 2)
				res, err := Aliyun.QueryDomainRecord(parts[0], parts[1])
				if err != nil {
					logError.Printf("%s, %s", domainName, err)
				} else {
					if res[0].Remark != code {
						err := Aliyun.UpdateDomainRemark(res[0].RecordId, code)
						if err != nil {
							logError.Printf("Update %s Remark Failed.", domainName)
						} else {
							logInfo.Printf("Update Remark: %s ------> %s", domainName, code)
						}
					}
				}
				if res[0].Value != ip {
					tps_code := GetTpsCode(res[0].Value)
					str := fmt.Sprintf("%s ------> Ali: %s(%s) DB: %s(%s)", domainName, res[0].Value, tps_code, ip, code)
					mu.Lock()
					exception = append(exception, str)
					mu.Unlock()
				}
			}
		})
	}
	pool.Wg.Wait()
	return exception
}
