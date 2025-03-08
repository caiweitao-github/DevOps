package fpsutil

import (
	"context"
	"database/sql"
	"util"

	"github.com/go-redis/redis/v8"
)

var FPSHA string

var (
	ip   = "localhost"
	port = "6379"
	db   = "5"
)

var fpsStatus = map[int]int{0: 1 << 0, -1: 1 << 1, -2: 1<<1 | 1}

var ctx = context.Background()

var rdb *redis.Client
var Db *sql.DB

var logFile = "/data/kdl/log/devops/FPS_HA.log"

var logInfo, logError = util.InitLog(logFile, "[INFO] ", "[ERROR] ")

const FpsBakCode = ""

var keys = ""

var Domain = ""

type FpsHaInfo interface {
	GetFpsDomain() error
	CheckFps()
	UpdataLoginIP()
}

type fpsHa struct {
	id            int
	status        int
	code          string
	ip            string
	locationCcode string
}

func init() {
	var err error
	rdb, err = util.RedisInit(ip, port, db)
	if err != nil {
		logError.Fatalf("RedisInit Failed : %s", err)
	}
	Db, err = util.dbDB()
	if err != nil {
		logError.Fatalf("ConnDb Failed : %s", err)
	}
	err = getFpsHA()
	if err != nil {
		logError.Fatalf("getFpsHA Failed : %s", err)
	}
}

func getFpsHA() (err error) {
	var ip string
	sqlStr := "select login_ip from fps where code = ?"
	err = Db.QueryRow(sqlStr, FpsBakCode).Scan(&ip)
	if err != nil {
		return
	} else {
		FPSHA = ip
		return
	}
}

func CheckNode() error {
	sqlStr := "select id,status,code,login_ip,location_code from fps where status in ('1', '3') and login_ip != ?"
	rows, err := Db.Query(sqlStr, FPSHA)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var fps fpsHa
		err := rows.Scan(&fps.id, &fps.status, &fps.code, &fps.ip, &fps.locationCcode)
		if err != nil {
			logError.Printf("scan failed, err:%v\n", err)
			return err
		}
		var FPS FpsHaInfo = fps
		_err := FPS.GetFpsDomain()
		if _err != nil {
			logError.Printf("GetTpsDomain Failed : %s", _err)
			break
		}
		FPS.CheckFps()
		FPS.UpdataLoginIP()
	}
	return nil
}
