package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"util"

	"github.com/go-redis/redis/v8"
)

var (
	ctx = context.Background()
	dps util.CkInfo
	// tps         util.CkInfo
	rdb         *redis.Client
	db          *sql.DB
	notifyUrl   = "https://open.feishu.cn/open-apis/bot/v2/hook/c6f343b3-f8e8-4e81-b27d-ae20bd31c495"
	notifyTitle = "[大客户最近一小时请求增加通知]"
	logFile     = "/data/kdl/log/devops/userReqCheckHour.log"
	logError    = util.LogConf(logFile, "[ERROR] ")
)

const (
	DpsKey = "DpsHoruReqHour"
	TpsKey = "TpsHoruReqHour"
	Count  = 150 * 60 * 60
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
	// tps, err = util.CkConn("kdl", "yvk8fcfb", "tpsstat")
	// if err != nil {
	// 	panic(err)
	// }
}

func getUser() {
	mess := make([]string, 0, 50)
	sqlSrt := `select auth_user.id,auth_user.username from user_profile,auth_user where user_profile.importance in (1, 2) and user_profile.user_id = auth_user.id and auth_user.is_staff = 0`
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
		// mess = append(mess, d.checkReq(tps, "tunnel_request_user_hour", TpsKey)...)
		mess = append(mess, d.checkReq(dps, "stat_request_user_hour", DpsKey)...)
	}
	// customizedUser := UserProfile{195986, "15913143909"}
	// mess = append(mess, customizedUser.checkReq(tps, "tunnel_request_user_hour", TpsKey)...)
	// mess = append(mess, customizedUser.checkReq(dps, "stat_request_user_hour", DpsKey)...)
	if len(mess) > 0 {
		util.StaffFeiShuNotify(notifyUrl, notifyTitle, mess)
	}
}

func (u UserProfile) checkReq(db util.CkInfo, table string, key string) []string {
	sqlStr := fmt.Sprintf("select toInt64(sum(count)),ROUND(sum(avg_request_size)/1024/1024/60/60, 2) as request,ROUND(sum(avg_response_size)/1024/1024/60/60, 2) as response from %s where day = toDate(now()) and hour between toHour(now() - toIntervalHour(2)) and toHour(now() - toIntervalHour(1)) and user_id = %d", table, u.userId)
	var num int64
	var request float64
	var response float64
	err := db.QueryRow(ctx, sqlStr).Scan(&num, &request, &response)
	if err != nil {
		logError.Fatalf("Query User Data For %s Fail : %v", table, err)
	}
	messageList := make([]string, 0, 50)
	if num > Count {
		k := rdb.HExists(ctx, key, u.userName).Val()
		if !k {
			rdb.HSet(ctx, key, u.userName, num)
			messageList = append(messageList, fmt.Sprintf("用户: %s, 请求次数: %d", u.userName, num))
		} else {
			cacheData, err := strconv.Atoi(rdb.HGet(ctx, key, u.userName).Val())
			if err != nil {
				logError.Fatalf("conversion data err %v", err)
			}
			// r := ((float64(num) - float64(cacheData)) / float64(num)) * 100
			previous := ((float64(num) - float64(cacheData)) / float64(cacheData)) * 100
			if previous > float64(30) {
				messageList = append(messageList, fmt.Sprintf("用户: %s, 请求次数: %d, 历史数据: %d, 每秒带宽(仅参考): 出 %1.fM 入 %1.f%% 环比增长率: %1.fM", u.userName, num, cacheData, request, response, previous))
			}
		}
	}
	rdb.HSet(ctx, key, u.userName, num)
	return messageList
}
