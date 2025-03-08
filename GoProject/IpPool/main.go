package main

import (
	"context"
	"database/sql"
	"strconv"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

var log = util.NewInitLog("/data/kdl/log/devops/IpPool.log")

var ctx = context.Background()

var rdb *redis.Client

var Db *sql.DB

type IPPool interface {
	writeInCache(timeStamp, nowTIme int64)
	strToTimeStamp(nowTIme int64)
}

type NodeData struct {
	Ip             string
	NextChangeTime int64
	Source         string
}

var PoolKey = ":Pool"

var POOL map[int]string = make(map[int]string, 64)

var LongPeriod = []int{7200, 10800, 14400, 21600, 43200, 86400}

var sleepTime = 3 * time.Second

func main() {
	for {
		startTime := time.Now()
		log.Info("run..")
		now := time.Now().Unix()
		clearExpireData(now)
		node, err := getNode()
		if err != nil {
			log.Errorf("get node err: %v", err)
		}
		for _, v := range node {
			var d IPPool = v
			d.strToTimeStamp(now)
		}
		log.Infof("elapsed: %v, sleep 3s.", time.Since(startTime))
		time.Sleep(sleepTime)
	}
}

func init() {
	var err error
	Db, err = util.KdlnodeDB()
	if err != nil {
		panic(err)
	}
	rdb, err = util.KldRedisDB()
	if err != nil {
		panic(err)
	}
	newObj()
}

func newObj() {
	for i := 60; i <= 3600; i += 60 {
		POOL[i] = strconv.Itoa(i) + PoolKey
	}

	for _, p := range LongPeriod {
		POOL[p] = strconv.Itoa(p) + PoolKey
	}
}

func getNode() (res []NodeData, e error) {
	sqlStr := "select public_ip,UNIX_TIMESTAMP(next_change_time) AS next_change_time,source from node where status=1 and role>=8 AND role<=24 and public_ip<>'' and exclude_for_transfer=0 and enable=1"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		e = err
		return
	}
	for rows.Next() {
		var d NodeData
		err := rows.Scan(&d.Ip, &d.NextChangeTime, &d.Source)
		if err != nil {
			e = err
			return
		}
		res = append(res, d)
	}
	return
}

func (n NodeData) strToTimeStamp(nowTIme int64) {
	timeStamp := n.NextChangeTime - nowTIme
	n.writeInCache(timeStamp, nowTIme)
}

func (n NodeData) writeInCache(timeStamp, nowTIme int64) {
	today := "-" + time.Now().Format(time.DateOnly)
	pipe := rdb.Pipeline()
	defer pipe.Close()
	processInterval := func(interval int, ttl float64) {
		if timeStamp >= int64(interval) {
			if r := rdb.ZScore(ctx, POOL[interval], n.Ip).Val(); r == 0 {
				pipe.Incr(ctx, POOL[interval]+"-"+n.Source+today)
				pipe.ZAdd(ctx, POOL[interval], &redis.Z{
					Score:  float64(nowTIme) + ttl,
					Member: n.Ip,
				})
			}
		}
	}

	for interval := 60; interval <= 3600; interval += 60 {
		var ttl float64
		if timeStamp <= 600 {
			ttl = 21600
		} else {
			ttl = 43200
		}
		processInterval(interval, ttl)
	}

	if timeStamp > 3600 {
		for _, v := range LongPeriod {
			if timeStamp > int64(v) {
				processInterval(v, 86400)
			}
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Errorf("write redis fail: %v", err)
	}
}

func clearExpireData(nowTIme int64) {
	for _, v := range POOL {
		rdb.ZRemRangeByScore(ctx, v, "0", strconv.FormatInt(nowTIme, 10))
	}
}
