package util

import (
	"database/sql"
	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

type Rdb *redis.Client

type Db *sql.DB

func dbDB() (*sql.DB, error) {
	Db, err := sql.Open("mysql", ":@tcp(127.0.0.1:3306)/db?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func KdlnodeDB() (*sql.DB, error) {
	Db, err := sql.Open("mysql", ":@tcp(127.0.0.1:3306)/kdlnode?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func JlynodeDB() (*sql.DB, error) {
	Db, err := sql.Open("mysql", ":@tcp(127.0.0.1:3306)/jlynode?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func NodeOpsDB() (*sql.DB, error) {
	Db, err := sql.Open("mysql", "@tcp(127.0.0.1:3306)/nodeops?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func JlyDB() (*sql.DB, error) {
	Db, err := sql.Open("mysql", "@tcp(127.0.0.1:3306)/jiliuip?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func JipFDB() (*sql.DB, error) {
	Db, err := sql.Open("mysql", ":@tcp(127.0.0.1:3306)/kdljip?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func KdlStat() (*sql.DB, error) {
	Db, err := sql.Open("mysql", ":@tcp(127.0.0.1:3306)/kdlstat?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func KdlStaff() (*sql.DB, error) {
	Db, err := sql.Open("mysql", ":@tcp(127.0.0.1:3306)/kdlstat?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func FlNode() (*sql.DB, error) {
	Db, err := sql.Open("mysql", ":@tcp(127.0.0.1:3306)/kdlstat?charset=utf8")
	if err != nil {
		return nil, err
	}
	return Db, nil
}

func RedisInit(ip, port, db string) (*redis.Client, error) {
	con := fmt.Sprintf("redis://%s:%s/%s", ip, port, db)
	if opt, err := redis.ParseURL(con); err == nil {
		rdb := redis.NewClient(opt)
		return rdb, nil
	} else {
		return nil, fmt.Errorf("redis init failed: %s", err)
	}
}

func RedisDB() (*redis.Client, error) {
	if opt, err := redis.ParseURL("redis://127.0.0.1:6379/5"); err == nil {
		rdb := redis.NewClient(opt)
		return rdb, nil
	} else {
		return nil, fmt.Errorf("redis init failed: %s", err)
	}
}

func IPPoolRedisDB() (*redis.Client, error) {
	if opt, err := redis.ParseURL("redis://127.0.0.1:6379/10"); err == nil {
		rdb := redis.NewClient(opt)
		return rdb, nil
	} else {
		return nil, fmt.Errorf("redis init failed: %s", err)
	}
}

func KldRedisDB() (*redis.Client, error) {
	if opt, err := redis.ParseURL("redis://127.0.0.1:6379/10"); err == nil {
		rdb := redis.NewClient(opt)
		return rdb, nil
	} else {
		return nil, fmt.Errorf("redis init failed: %s", err)
	}
}

func JlyRedisDB() (*redis.Client, error) {
	if opt, err := redis.ParseURL("redis://127.0.0.1:6379/3"); err == nil {
		rdb := redis.NewClient(opt)
		return rdb, nil
	} else {
		return nil, fmt.Errorf("redis init failed: %s", err)
	}
}

func RedisCrs() (*redis.Client, error) {
	if opt, err := redis.ParseURL("redis://127.0.0.1:6379/5"); err == nil {
		rdb := redis.NewClient(opt)
		return rdb, nil
	} else {
		return nil, fmt.Errorf("redis init failed: %s", err)
	}
}

func RedisDevOps() (*redis.Client, error) {
	if opt, err := redis.ParseURL("redis://10.0.3.17:6380/5"); err == nil {
		rdb := redis.NewClient(opt)
		return rdb, nil
	} else {
		return nil, fmt.Errorf("redis init failed: %s", err)
	}
}
