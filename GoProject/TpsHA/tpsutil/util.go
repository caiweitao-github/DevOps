package tpsutil

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
)

var TPSHA string

var (
	ip   = "localhost"
	port = "6379"
	db   = "5"
)

var tpsStatus = map[int]int{0: 1 << 0, -1: 1 << 1, -2: 1<<1 | 1}

var ctx = context.Background()

var rdb *redis.Client
var Db *sql.DB

var (
	logFile  = "/data/kdl/log/devops/TPS_HA.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

const (
	CREATE_TPS_COUNT    = "create_tps_count"
	TPS_HA_INFO_KEY     = "_ha_info"
	EXCEPTION_TPS_COUNT = "exception_tps"
	TPS_HA_USE          = "tps_ha_is_use"
)

var keys = "domains"


func init() {
	var err error
	rdb, err = util.RedisInit(ip, port, db)
	if err != nil {
		logError.Printf("RedisInit Failed : %s", err)
	}
	Db, err = util.ConnDb("", "", "")
	if err != nil {
		logError.Printf("ConnDb Failed : %s", err)
	}
	err = getTpsHA()
	if err != nil {
		logError.Printf("getTpsHA Failed : %s", err)
	}
}

func getTpsHA() (err error) {
	sqlStr := "select login_ip from tps where code = 'tps56'"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ip string
		err = rows.Scan(&ip)
		if err != nil {
			return
		}
		TPSHA = ip
	}
	return nil
}

func GetTpsIP() string {
	tpsCode := []string{"tpsb1", "tpsb2", "tpsb3", "tpsb4", "tpsb5"}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rng.Intn(len(tpsCode))
	code := tpsCode[randomIndex]
	ip := rdb.HGet(ctx, code+TPS_HA_INFO_KEY, "ip").Val()
	return ip
}

func CheckCode(ip string) bool {
	sqlStr := "select code from tps where login_ip = ? and code not REGEXP 'tpsb.*|tpsT.*|tpstest|tpsysy'"
	var code string
	err := Db.QueryRow(sqlStr, ip).Scan(&code)
	if err != nil {
		logError.Printf("scan failed: %s", err)
		return false
	}
	if code == "" {
		return false
	} else {
		return true
	}
}

func UpdateDB(status int, ip, code string) error {
	sqlStr := "update tps set status = ?, login_ip = ? where code = ?"
	_, err := Db.Exec(sqlStr, status, ip, code)
	if err != nil {
		return err
	}
	return nil
}

func GetExceptionTps() (bool, error) {
	sqlStr := "select count(*) from tps where status = 3"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var count int
		err := rows.Scan(&count)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return false, err
		}
		if count > 2 {
			if res, _ := rdb.SetNX(ctx, EXCEPTION_TPS_COUNT, 1, 0).Result(); !res {
				rdb.Incr(ctx, EXCEPTION_TPS_COUNT)
			}
		} else {
			rdb.Set(ctx, EXCEPTION_TPS_COUNT, 0, 0)
		}
	}
	num, err := rdb.Get(ctx, EXCEPTION_TPS_COUNT).Result()
	if err != nil {
		return false, err
	}
	flag, _ := strconv.Atoi(num)
	if flag >= 3 {
		return true, nil
	}
	return false, nil
}
