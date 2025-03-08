package main

import (
	"context"
	"fmt"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
)

var (
	logFile  = "/data/kdl/log/devops/getDomainTimeUse.log"
	logError = util.LogConf(logFile, "[ERROR] ")
)

var ctx = context.Background()
var tps util.CkInfo

const (
	avgTimeUse = 5000
	countNum   = 500
	DomainKey  = "domain_time_use"
)

var (
	rdb  *redis.Client
	ip   = "localhost"
	port = "6379"
	db   = "5"
)

type Data struct {
	Domain  string `ch:"domain"`
	TimeUse int64  `ch:"time_total"`
	Num     int64  `ch:"num"`
}

func main() {
	r := getData()
	updataCache(r)
}

func init() {
	var err error
	tps, err = util.CkConn("kdl", "yvk8fcfb", "tpsstat")
	if err != nil {
		logError.Fatalf("Conn ClickHouse Err: %s", err)
	}
	rdb, err = util.RedisInit(ip, port, db)
	if err != nil {
		logError.Printf("RedisInit Failed : %s", err)
	}
}

func getData() (dataList []Data) {
	now := time.Now()
	day := now.Format(time.DateOnly)
	yesterday := now.Add(-24 * time.Hour).Format(time.DateOnly)
	sqlStr := fmt.Sprintf(`SELECT domain, toInt64(ROUND(AVG(avg_time_total),0)) AS time_total,
	toInt64(SUM(count)) AS num FROM tunnel_request_user_day WHERE day BETWEEN '%s' AND '%s' AND tid != '' 
	AND tid != 'tpsmonitor' AND status = 0 AND avg_time_total > %d GROUP BY domain HAVING 
	num > %d`, yesterday, day, avgTimeUse, countNum)
	rows, err := tps.Query(ctx, sqlStr)
	if err != nil {
		logError.Fatalf("Get Data Err: %s", err)
	}
	for rows.Next() {
		var res Data
		if err := rows.ScanStruct(&res); err != nil {
			logError.Fatalf("Scan Data Err: %s", err)
		}
		dataList = append(dataList, res)
	}
	return
}

func updataCache(d []Data) {
	rdb.Del(ctx, DomainKey)
	for _, v := range d {
		rdb.HSet(ctx, DomainKey, v.Domain, v.TimeUse)
	}
}
