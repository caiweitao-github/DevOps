package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"util"
)

var (
	logFile  = "/data/kdl/log/devops/requests_check.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

var ctx = context.Background()
var node util.CkInfo
var tps util.CkInfo
var DB *sql.DB

type Ti struct {
	n  string
	h1 string
	h2 string
}

var newtime *Ti

func init() {
	var err error
	node, err = util.CkConn("", "", "")
	if err != nil {
		logError.Fatalf("Conn ClickHouse Err: %s", err)
	}
	tps, err = util.CkConn("", "", "")
	if err != nil {
		logError.Fatalf("Conn ClickHouse Err: %s", err)
	}
	DB, err = util.ConnDb("", "", "")
	if err != nil {
		logError.Fatalf("Conn DB Err: %s", err)
	}
	newtime = NewTime(time.Now())
}
func main() {
	logInfo.Println("run ...")
	GetData()
	GetTpsData()
	TpsReqLimt()
}

func NewTime(now time.Time) *Ti {
	return &Ti{
		n:  fmt.Sprintf("%s:%s", now.Format("2006-01-02 15"), "00:00"),
		h1: fmt.Sprintf("%s:%s", now.Add(-1*time.Hour).Format("2006-01-02 15"), "00:00"),
		h2: fmt.Sprintf("%s:%s", now.Add(-2*time.Hour).Format("2006-01-02 15"), "00:00"),
	}
}

func GetData() {
	var count int64
	var count1 int64
	sqlStr := fmt.Sprintf(`SELECT toInt64(count1),toInt64(count2) FROM (SELECT count(*) AS count1 FROM node_request_history WHERE request_time BETWEEN '%s' AND '%s') AS 
	t1 CROSS JOIN (SELECT count(*) AS count2 FROM node_request_history WHERE request_time BETWEEN '%s' AND '%s') AS t2`, newtime.h1, newtime.n, newtime.h2, newtime.h1)
	rows, err := node.Query(ctx, sqlStr)
	if err != nil {
		logError.Fatalf("GetData Err: %s", err)
	}
	for rows.Next() {
		if err := rows.Scan(
			&count,
			&count1,
		); err != nil {
			logError.Fatalf("Scan GetData Err: %s", err)
		}
	}
	defer rows.Close()
	if ((float64(count)-float64(count1))/float64(count))*100 > 30 {
		messStr := GetDetailedData()
		if len(messStr) > 0 {
			messStr = append([]string{fmt.Sprintf("%s -- %s请求数: %d, %s -- %s请求数: %d", newtime.h1, newtime.n, count, newtime.h2, newtime.h1, count1)}, messStr...)
			util.SendMess2(messStr, "[dps总请求超过最近一小时30%]")
		}
	}
}

func GetTpsData() {
	var count int64
	var count1 int64
	sqlStr := fmt.Sprintf(`SELECT toInt64(count1),toInt64(count2) FROM (SELECT count(*) AS count1 FROM tunnel_request_history WHERE request_time BETWEEN '%s' AND '%s') AS 
	t1 CROSS JOIN (SELECT count(*) AS count2 FROM tunnel_request_history WHERE request_time BETWEEN '%s' AND '%s') AS t2`, newtime.h1, newtime.n, newtime.h2, newtime.h1)
	rows, err := tps.Query(ctx, sqlStr)
	if err != nil {
		logError.Fatalf("GetData Err: %s", err)
	}
	for rows.Next() {
		if err := rows.Scan(
			&count,
			&count1,
		); err != nil {
			logError.Fatalf("Scan GetData Err: %s", err)
		}
	}
	defer rows.Close()
	if ((float64(count)-float64(count1))/float64(count))*100 > 30 {
		messStr := GetTpsDetailedData()
		if len(messStr) > 0 {
			messStr = append([]string{fmt.Sprintf("%s -- %s请求数: %d, %s -- %s请求数: %d", newtime.h1, newtime.n, count, newtime.h2, newtime.h1, count1)}, messStr...)
			util.SendMess2(messStr, "[tps总请求超过最近一小时30%]")
		}
	}
}

func GetTpsDetailedData() []string {
	mess := []string{}
	var (
		tid   string
		count int64
	)
	sqlStr := fmt.Sprintf(`select tid,toInt64(count(*)) as count from tunnel_request_history where request_time between 
	'%s' and '%s' and status = 0 group by tid order by count desc limit 5`, newtime.h1, newtime.n)
	rows, err := tps.Query(ctx, sqlStr)
	if err != nil {
		logError.Fatalf("GetDetailedData Err: %s", err)
	}
	for rows.Next() {
		if err := rows.Scan(
			&tid,
			&count,
		); err != nil {
			logError.Fatalf("Scan GetDetailedData Err: %s", err)
		}
		mess = append(mess, fmt.Sprintf("Tid: %s, 请求次数: %d", tid, count))
	}
	defer rows.Close()
	return mess
}

func GetDetailedData() []string {
	mess := []string{}
	var (
		user_id int64
		orderid string
		count   int64
	)
	sqlStr := fmt.Sprintf(`select toInt64(user_id) as user_id,orderid,toInt64(count(*)) as count from node_request_history 
	where request_time between '%s' and '%s' and status = 0 and user_id != 0 group by user_id,orderid order by count desc limit 5`, newtime.h1, newtime.n)
	rows, err := node.Query(ctx, sqlStr)
	if err != nil {
		logError.Fatalf("GetDetailedData Err: %s", err)
	}
	for rows.Next() {
		if err := rows.Scan(
			&user_id,
			&orderid,
			&count,
		); err != nil {
			logError.Fatalf("Scan GetDetailedData Err: %s", err)
		}
		mess = append(mess, fmt.Sprintf("用户ID: %d, 订单号: %s, 请求次数: %d", user_id, orderid, count))
	}
	defer rows.Close()
	return mess
}

func getUserName(uid int32) (userName string) {
	sqlStr := "select username from auth_user where id = ?"
	if err := DB.QueryRow(sqlStr, uid).Scan(&userName); err != nil {
		logError.Fatalf("GetUserErr: %s", err)
	}
	return
}

func getOrderID(tid string) (orderid string) {
	sqlStr := "select proxy_order.orderid from proxy_order,tunnel where tunnel.order_id = proxy_order.id and tunnel.tid = ?"
	if err := DB.QueryRow(sqlStr, tid).Scan(&orderid); err != nil {
		logError.Fatalf("GetOrderIDErr: %s", err)
	}
	return
}

func TpsReqLimt() {
	mess := []string{}
	sqlStr := fmt.Sprintf(`select user_id,tid,toInt64(count(*)) as count from tunnel_request_history where request_time between 
	'%s' and '%s' and status = 8 group by tid,user_id HAVING count > 200000 order by count desc`, newtime.h1, newtime.n)
	rows, err := tps.Query(ctx, sqlStr)
	if err != nil {
		logError.Fatalf("GetDetailedData Err: %s", err)
	}
	var user_id int32
	var tid string
	var count int64
	for rows.Next() {
		if err := rows.Scan(
			&user_id,
			&tid,
			&count,
		); err != nil {
			logError.Fatalf("Scan GetDetailedData Err: %s", err)
		}
		userName := getUserName(user_id)
		orderID := getOrderID(tid)
		mess = append(mess, fmt.Sprintf("用户名: %s 订单号: %s 超频次数: %d", userName, orderID, count))
	}
	defer rows.Close()
	if len(mess) > 0 {
		util.SendMess3(mess, "[tps超频通知]")
	}
}
