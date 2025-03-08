package main

import (
	"context"
	"fmt"
	"util"
)

var (
	logFile  = "/data/kdl/log/devops/apiRequestCheck.log"
	logError = util.LogConf(logFile, "[ERROR] ")
)

var ctx = context.Background()
var api util.CkInfo

const countNum = 1000

func main() {
	m1, m2 := getData()
	diffDate(m1, m2)
}

func init() {
	var err error
	api, err = util.CkConn("", "", "")
	if err != nil {
		logError.Fatalf("Conn ClickHouse Err: %s", err)
	}
}

func getData() (m1, m2 map[string]int64) {
	m1, m2 = make(map[string]int64), make(map[string]int64)
	for i := 1; i <= 2; i++ {
		sqlStr := fmt.Sprintf(`select api_name,toInt64(SUM(call_count)) as count from api_stat_hour where 
		day = toDate(now()) and hour between toHour(now() - toIntervalHour(%d)) and toHour(now() - toIntervalHour(%d))
		group by api_name having count > %d`, i, i-1, countNum)
		rows, err := api.Query(ctx, sqlStr)
		if err != nil {
			logError.Fatalf("Get Data Err: %s", err)
		}
		var m map[string]int64
		switch i {
		case 1:
			m = m1
		case 2:
			m = m2
		}
		for rows.Next() {
			var s1 string
			var n1 int64
			if err := rows.Scan(&s1, &n1); err != nil {
				logError.Fatalf("Scan Data Err: %s", err)
			}
			m[s1] = n1
		}
	}
	return m1, m2
}

func diffDate(d1, d2 map[string]int64) {
	mess := []string{}
	for k, v := range d1 {
		if v2, ok := d2[k]; ok && (v-v2)/v2 >= 1 {
			mess = append(mess, fmt.Sprintf("ApiName: %s Count: %d --> %d", k, v2, v))
		}
	}
	if len(mess) > 0 {
		util.SendMess2(mess, "[Api请求增高通知]")
	}
}
