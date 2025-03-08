package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"util"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var dps util.CkInfo

var tps util.CkInfo

var rdb *redis.Client

var db *sql.DB

var (
	logFile  = "/data/kdl/log/devops/userReqCheckDay.log"
	logError = util.LogConf(logFile, "[ERROR] ")
)

const (
	DpsKey = "DpsHoruReqDay"
	TpsKey = "TpsHoruReqDay"
	Count  = 50 * 60 * 60 * 24
)

type UserProfile struct {
	userId   int
	userName string
}

func main() {
	getUser()
}

func init() {
	var err error
	rdb, err = util.RedisDB()
	if err != nil {
		panic(err)
	}
	db, err = util.dbDB()
	if err != nil {
		panic(err)
	}
	dps, err = util.CkConn("kdl", "yvk8fcfb", "nodeops")
	if err != nil {
		panic(err)
	}
	tps, err = util.CkConn("kdl", "yvk8fcfb", "tpsstat")
	if err != nil {
		panic(err)
	}
}

func getUser() {
	mess := []string{}
	sqlSrt := `select user_profile.user_id,auth_user.username from user_profile,auth_user where 
	user_profile.importance = 2 and user_profile.user_id = auth_user.id and auth_user.is_staff = 0`
	rows, err := db.Query(sqlSrt)
	if err != nil {
		logError.Fatalf("Query err : %s", err)
	}
	for rows.Next() {
		var d UserProfile
		err := rows.Scan(&d.userId, &d.userName)
		if err != nil {
			logError.Fatalf("Scan err : %s", err)
		}
		d.checkReq(mess, tps, "tunnel_request_user_hour", TpsKey)
		d.checkReq(mess, dps, "stat_request_user_hour", DpsKey)
	}
	if len(mess) > 0 {
		util.SendMess2(mess, "[大客户请求增加通知]")
	}
}

func (u UserProfile) checkReq(m []string, db util.CkInfo, table string, key string) {
	req := fmt.Sprintf(`select toInt64(sum(count)) from %s 
	where day = toDate(now() - toIntervalDay(1)) and user_id = %d`, table, u.userId)
	rows, err := db.Query(ctx, req)
	if err != nil {
		logError.Fatalf("Query User Data For %s Fail : %v", table, err)
	}
	for rows.Next() {
		var num int64
		err := rows.Scan(&num)
		if err != nil {
			logError.Fatalf("Scan %s Data err : %v", table, err)
		}
		if num > Count {
			cacheData, _ := strconv.Atoi(rdb.HGet(ctx, key, u.userName).Val())
			if num > int64(cacheData) {
				rdb.HSet(ctx, key, u.userName, num)
			}
			if r := ((float64(num) - float64(cacheData)) / float64(num)) * 100; r > 30 {
				m = append(m, fmt.Sprintf("用户: %s, 请求次数: %d, 历史最大请求次数: %d, 增长率: %.1f", u.userName, num, cacheData, r))
			}
		}
	}
}
