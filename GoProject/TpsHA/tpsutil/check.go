package tpsutil

import (
	"Aliyun"
	"Tencloud"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"util"
)

var maxprocess = 30 >> 1

var title = "[TPS异常告警]"

var notifyUrl = ""

type TpsHaInfo interface {
	GetTpsDomain() error
	CheckTps(isCreate bool)
	UpdataLoginIP()
	SetTpsStatus()
}

type TpsHa struct {
	id     int
	code   string
	ip     string
	status int
}

func CheckNode() error {
	isCreate, _ := GetExceptionTps()

	sqlStr := "select id,code,login_ip,status from tps where status in ('1', '3') and login_ip != ? and code not REGEXP 'tpsb.*|tpsT.*|tpstest|tpsysy'"
	rows, err := Db.Query(sqlStr, TPSHA)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tps TpsHa
		err := rows.Scan(&tps.id, &tps.code, &tps.ip, &tps.status)
		if err != nil {
			logError.Printf("scan failed, err:%v\n", err)
			return err
		}
		var TPS TpsHaInfo = tps
		_err := TPS.GetTpsDomain()
		if _err != nil {
			logError.Printf("GetTpsDomain Failed : %s", _err)
			break
		}
		TPS.CheckTps(isCreate)
		TPS.UpdataLoginIP()
		TPS.SetTpsStatus()
	}
	return nil
}

func (t TpsHa) UpdataLoginIP() {
	tpsip := rdb.HGet(ctx, t.code, "ip").Val()
	if tpsip != t.ip {
		rdb.HSet(ctx, t.code, "ip", t.ip)
	}
}

func (t TpsHa) GetTpsDomain() error {
	domainList := make([]string, 0)
	sqlStr := "select domain from tps_domain where tps_id = ? and status != 4"
	rows, err := Db.Query(sqlStr, t.id)
	if err != nil {
		return err
	}
	defer rows.Close()
	cacheKey := fmt.Sprintf("%s_%s", t.code, keys)
	for rows.Next() {
		var domain string
		err := rows.Scan(&domain)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return err
		}
		domainList = append(domainList, domain)
		rdb.SAdd(ctx, cacheKey, domain)
	}
	cacheDomain, err := rdb.SMembers(ctx, cacheKey).Result()
	if err != nil {
		return err
	}
	dbdata := strings.Join(domainList, ",")
	for _, v := range cacheDomain {
		if !strings.Contains(dbdata, v) {
			rdb.SRem(ctx, cacheKey, v)
		}
	}
	return nil
}

func (t TpsHa) UpdataTpsDomain(ip string, TTL int64, domains ...string) {
	ch := make(chan struct{}, maxprocess)
	done := make(chan bool)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	re := regexp.MustCompile(`\.`)
	go func() {
		defer close(ch)
		defer close(done)
		for i := 0; i < len(domains); i++ {
			<-done
		}
		done <- true
	}()
	for _, domain := range domains {
		ch <- struct{}{}
		go func(domain string) {
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(rng.Intn(801)+500))
				<-ch
				done <- true
			}()
			data := re.Split(domain, 2)
			err := Aliyun.UpdateTpsDomain(data[0], data[1], "A", ip, TTL)
			if err != nil {
				logError.Printf("Update %s Domain Failed: %s", domain, err)
			} else {
				logInfo.Printf("Updata Domain %s ------> %s.", domain, ip)
			}
		}(domain)
	}
	<-done
}

func (t TpsHa) CheckTps(isCreate bool) {
	keyIsExists := rdb.Exists(ctx, t.code).Val()
	if keyIsExists == 0 {
		if t.status == 1 {
			rdb.HMSet(ctx, t.code, map[string]interface{}{"id": t.id, "ip": t.ip, "status": t.status, "ha": 0, "ha_ip": "None"})
		} else {
			rdb.HMSet(ctx, t.code, map[string]interface{}{"id": t.id, "ip": t.ip, "status": 0, "ha": 0, "ha_ip": "None"})
			util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s Down! ------> %d次", t.code, tpsStatus[0])})
		}
	} else {
		domains := rdb.SMembers(ctx, fmt.Sprintf("%s_%s", t.code, keys)).Val()
		TpsStatus, _ := strconv.Atoi(rdb.HGet(ctx, t.code, "status").Val())
		HaStatus, _ := strconv.Atoi(rdb.HGet(ctx, t.code, "ha").Val())
		if t.status != TpsStatus && t.status == 3 && TpsStatus > -2 {
			rdb.HIncrBy(ctx, t.code, "status", -1)
			if HaStatus == 0 {
				util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s Down! ------> %d次", t.code, tpsStatus[TpsStatus-1])})
			}
			CachaStatus, _ := strconv.Atoi(rdb.HGet(ctx, t.code, "status").Val())
			if CachaStatus == -2 && HaStatus == 0 && t.ip != TPSHA {
				util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 连续3次检测为异常, 切换所有域名至备用机.", t.code)})
				if !isCreate {
					t.UpdataTpsDomain(TPSHA, 1, domains...)
					util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 域名: %s 切换至备用机 %s", t.code, strings.Join(domains, "\u0020"), TPSHA)})
					rdb.HMSet(ctx, t.code, map[string]interface{}{"ha": 1, "ha_ip": TPSHA})
					if res := rdb.SetNX(ctx, TPS_HA_USE, 1, 0).Val(); !res {
						rdb.Incr(ctx, TPS_HA_USE)
					}
				} else {
					ha_use, _ := strconv.Atoi(rdb.Get(ctx, TPS_HA_USE).Val())
					if ha_use < 1 {
						t.UpdataTpsDomain(TPSHA, 1, domains...)
						util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 域名: %s 切换至备用机 %s", t.code, strings.Join(domains, "\u0020"), TPSHA)})
						rdb.HMSet(ctx, t.code, map[string]interface{}{"ha": 1, "ha_ip": TPSHA})
						if res := rdb.SetNX(ctx, TPS_HA_USE, 1, 0).Val(); !res {
							rdb.Incr(ctx, TPS_HA_USE)
						}
					} else {
						if res := rdb.SetNX(ctx, CREATE_TPS_COUNT, 1, 0).Val(); !res {
							rdb.Incr(ctx, CREATE_TPS_COUNT)
						}
						nodeNum, _ := strconv.Atoi(rdb.Get(ctx, CREATE_TPS_COUNT).Val())
						if nodeNum > 5 {
							util.FeiShuNotify(notifyUrl, title, []string{"新建机器数量已达上限, 后续异常机器将随机切换至正常的备用机."})
							bakIp := GetTpsIP()
							if bakIp != "" {
								logError.Print("Get BakIP Failed!")
							} else {
								t.UpdataTpsDomain(bakIp, 1, domains...)
								util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 域名: %s 切换至备用机 %s", t.code, strings.Join(domains, "\u0020"), bakIp)})
							}
						} else {
							id, newIP, err := Tencloud.GetNodeIp()
							if err != nil {
								logError.Printf("GetNodeIp Failed : %s", err)
								util.FeiShuNotify(notifyUrl, title, []string{"新建机器失败!"})
							} else {
								util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("新建机器成功, tpsb%d ------> %s", nodeNum, newIP)})
								code := fmt.Sprintf("tpsb%d", nodeNum)
								var Cmd util.SshUtil = util.SshInfo{Host: newIP, Port: "8022"}
								SshCfg, err := Cmd.InitSsh()
								if err != nil {
									logError.Printf("InitSsh Failed : %s", err)
								}
								cmd := []string{
									fmt.Sprintf("source /root/.bash_profile;hostnamectl set-hostname tpsb%d", nodeNum),
									fmt.Sprintf("source /root/.bash_profile;sed -ri 's/(.*)tpsb1/\\1tpsb%d/g' /root/.bash_profile", nodeNum),
									fmt.Sprintf("source /root/.bash_profile;sed -ri 's/(.*)tpsb1/\\1tpsb%d/g' /root/tps/tps.cfg;sed -ri 's/(.*)nqsb1/\\1nqsb%d/g' /root/tps_nqs/tps_nqs.cfg", nodeNum, nodeNum),
									fmt.Sprintf("source /root/.bash_profile;sed -ri 's/nqsb1/nqsb%d/g' /etc/nginx/server_nqs;nginx -s reload", nodeNum),
									"source /root/.bash_profile;ps -ef|grep -E 'tps|tps_nqs'|grep -v grep|awk '{print $2}'|xargs kill;systemctl daemon-reload;systemctl start proxy_tps;systemctl start proxy_tps_b;bash /root/tps_nqs/rc_tps_nqs"}
								Cmd.Conn(SshCfg, cmd)
								_err := UpdateDB(1, newIP, code)
								if _err != nil {
									logError.Printf("UpdateDB Failed : %s", err)
								} else {
									t.UpdataTpsDomain(newIP, 1, domains...)
									t.UpdataTpsDomain(newIP, 600, fmt.Sprintf("nqsb%d.gizaworks.com", nodeNum))
									rdb.HMSet(ctx, t.code, map[string]interface{}{"ha": 1, "ha_ip": newIP})
									rdb.HMSet(ctx, t.code+TPS_HA_INFO_KEY, map[string]interface{}{"code": code, "server_id": id, "ip": newIP})
									util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 域名: %s 切换至备用机 %s", t.code, strings.Join(domains, "\u0020"), newIP)})
								}
							}
						}
					}
				}
			}
		} else if t.status != TpsStatus && t.status == 1 {
			rdb.HIncrBy(ctx, t.code, "status", 1)
			HaStatus, _ := strconv.Atoi(rdb.HGet(ctx, t.code, "ha").Val())
			HaIp := rdb.HGet(ctx, t.code, "ha_ip").Val()
			if TpsStatus == 0 && HaStatus == 1 && t.ip != TPSHA && HaIp != "None" {
				t.UpdataTpsDomain(t.ip, 600, domains...)
				rdb.HMSet(ctx, t.code, map[string]interface{}{"ha": 0, "ha_ip": "None"})
				code := CheckCode(HaIp)
				if HaIp == TPSHA {
					rdb.Decr(ctx, TPS_HA_USE)
				}
				if !code {
					res := rdb.Expire(ctx, t.code+TPS_HA_INFO_KEY, time.Minute*30).Val()
					if !res {
						logError.Printf("Set %s Expire Failed.", t.code+TPS_HA_INFO_KEY)
					}
				}
				util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 恢复正常, 切回所有域名: %s", t.code, strings.Join(domains, "\u0020"))})
			}
		}
	}
}

func (t TpsHa) SetTpsStatus() {
	ex := rdb.Exists(ctx, t.code+TPS_HA_INFO_KEY).Val()
	HaIp := rdb.HGet(ctx, t.code, "ha_ip").Val()
	if ex != 0 && HaIp != "None" {
		bakIp := rdb.HGet(ctx, t.code+TPS_HA_INFO_KEY, "ip").Val()
		if bakIp == HaIp {
			res := rdb.Expire(ctx, t.code+TPS_HA_INFO_KEY, time.Minute*30).Val()
			if !res {
				logError.Printf("Set %s Expire Failed.", t.code+TPS_HA_INFO_KEY)
			}
		}
	}
}

func TerminateServer() {
	keys := fmt.Sprintf("*%s", TPS_HA_INFO_KEY)
	res := rdb.Keys(ctx, keys).Val()
	if len(res) > 0 {
		for _, v := range res {
			if time := rdb.TTL(ctx, v).Val(); time.Seconds() < 900 {
				data := rdb.HGetAll(ctx, v).Val()
				_err := UpdateDB(2, data["ip"], data["code"])
				if _err != nil {
					logError.Printf("UpdateDB Failed : %s", _err)
				}
				err := Tencloud.DeleteServer(data["server_id"])
				if err != nil {
					logError.Printf("DeleteServer Failed : %s", err)
					util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 删除失败.", data["code"])})
				} else {
					rdb.Del(ctx, v)
					util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 删除成功.", data["code"])})
					rdb.IncrBy(ctx, CREATE_TPS_COUNT, -1)
				}
			}
		}
	}
}
