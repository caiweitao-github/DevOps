package main

import (
	"Aliyun"
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"time"
	"util"
)

var Db *sql.DB

var (
	logFile  = "/data/kdl/log/devops/tpsProDomainCheck.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

const (
	domaiNum   = 5
	domain     = "kdltpspro.com"
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

func main() {
	version := flag.Bool("version", false, "show version.")
	flag.Parse()
	if *version {
		fmt.Println("Author: weitaocai")
		fmt.Println("Version: 1.0.0")
		fmt.Println("Builder: goreleaser")
		fmt.Println("Date: 2024-04-15")
		return
	}
	defer func(pre time.Time) {
		elapsed := time.Since(pre)
		logInfo.Printf("elapsed: %v", elapsed)
	}(time.Now())
	logInfo.Println("Run TpsProDomainCheck ...")
	CheckTps()
}

func init() {
	var err error
	Db, err = util.dbDB()
	if err != nil {
		panic(err)
	}
}

func CheckTps() {
	mess := []string{}
	TpsList := make([]tps, 0, 30)
	sqlStr := `select ANY_VALUE(tps.id) as id,ANY_VALUE(tps.code) as code,tps.login_ip,count(ANY_VALUE(tps_domain.tps_id)) as num,
    format(ANY_VALUE(tps.avg_bandwidth)/1024/128,0) as avg_bandwidth,ANY_VALUE(tps.bandwidth) as bandwidth from tps 
    left join tps_domain on tps.id = tps_domain.tps_id and tps_domain.status = 1  where tps.code not REGEXP 'tpsb.*|tpstest|tpsysy|tps56' 
	and tps_domain.domain like '%kdltpspro.com' and tps.status = 1 group by tps.login_ip`
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
				if err == nil {
					mess = append(mess, str...)
				}
			} else {
				logInfo.Printf("%s机器域名数量低于预设值, 但是机器带宽过高!", v.code)
			}
		}
	}
	if len(mess) > 0 {
		util.SendMess2(mess, "[新建TpsPro域名解析通知]")
	}
}

func CreateDomain(id int, ip string, code string) ([]string, error) {
	str := []string{}
	sqlStr := `insert into tps_domain(domain, domain2, tps_id, status, use_type, create_time) values(?, ?, ?, 1, 1, CURRENT_TIMESTAMP())`
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := 1
LOOP:
	for i <= 5 {
		prefix := 'a' + rng.Intn(26)
		suffix := rng.Intn(901) + 100
		domainList := []string{fmt.Sprintf("%c%d", prefix, suffix), fmt.Sprintf("%c%d", prefix, suffix+1)}
		_, err := Aliyun.QueryDomainRecord(domainList[0], domain)
		_, err2 := Aliyun.QueryDomainRecord(domainList[1], domain)
		if err == nil || err2 == nil {
			logError.Printf("%s or %s Domain Already Exist, Skip!", domainList[0], domainList[1])
			goto LOOP
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
