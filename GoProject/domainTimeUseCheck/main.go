package main

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
)

var (
	logFile  = "/data/kdl/log/devops/domainTimeUseCheck.log"
	logError = util.LogConf(logFile, "[ERROR] ")
)

var ctx = context.Background()
var tps util.CkInfo

const (
	DiffNum    = 2000
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
	TpsCode string `ch:"tps_code"`
	Domain  string `ch:"domain"`
	Tid     string `ch:"tid"`
	UserID  int32  `ch:"user_id"`
	TimeUse int64  `ch:"time_total"`
	Num     int64  `ch:"num"`
}

func main() {
	r := getData()
	diffData(r)
}

func init() {
	var err error
	tps, err = util.CkConn("", "", "")
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
	hour := int(now.Hour())
	beforeHour := int(now.Add(-1 * time.Hour).Hour())
	sqlStr := fmt.Sprintf(`SELECT tps_code,user_id,tid,domain,toInt64(ROUND(AVG(avg_time_total),0)) AS time_total,
	toInt64(SUM(count)) AS num FROM tunnel_request_user_hour WHERE day = '%s' AND hour BETWEEN '%d' AND '%d' AND tid != '' 
	AND tid != 'tpsmonitor' AND status = 0 AND avg_time_total > %d GROUP BY tps_code,user_id,tid,domain HAVING 
	num > %d`, day, beforeHour, hour, avgTimeUse, countNum)
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

func diffData(d []Data) {
	mess := make([]string, 0)
	for _, v := range d {
		num, _ := strconv.Atoi(rdb.HGet(ctx, DomainKey, v.Domain).Val())
		if v.TimeUse > int64(num)+DiffNum {
			mess = append(mess, fmt.Sprintf("Code: %s UID: %d TID: %s Domain: %s TimeUser: %d Count: %d", v.TpsCode, v.UserID, v.Tid, v.Domain, v.TimeUse, v.Num))
		}
	}
	if len(mess) > 0 {
		util.SendMess2(mess, "[请求高耗时通知]")
	}
}
