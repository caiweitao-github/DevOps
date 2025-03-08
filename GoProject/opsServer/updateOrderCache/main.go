package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"opsServer/service"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
)

var key = ""

var auth = ""

var Encryption = md5.Sum([]byte(key + "_" + auth))

var apiAuth = hex.EncodeToString(Encryption[:])

var masterDomain = "https://test.com/"

var OrderKey = "OrderSet"

var tmpKey = OrderKey + "_tmp"

var rdb *redis.Client

var ctx = context.Background()

var log = util.NewInitLog("/data/kdl/log/updateOrderCache.log")

var sleepTime = 30 * time.Second

func main() {
	for {
		startTime := time.Now()
		tmpData := make([]string, 0, 25000)
		d := getNode[[]service.OrderDtat]()
		for _, v := range *d.Data {
			if v.SecretID != nil {
				tmpData = append(tmpData, *v.SecretID)
			}
			tmpData = append(tmpData, v.OrderID)
		}
		rdb.SAdd(ctx, tmpKey, tmpData)
		rdb.Rename(ctx, tmpKey, OrderKey)
		log.Infof("elapsed: %v, sleep 20s.", time.Since(startTime))
		time.Sleep(sleepTime)
	}
}

func init() {
	var err error
	rdb, err = util.RedisDevOps()
	if err != nil {
		panic(err)
	}
}

func getNode[T service.Type]() service.Response[T] {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/getorder"
	client := &http.Client{}
	reqest, err := http.NewRequest(http.MethodGet, domain.String(), nil)
	reqest.Header.Add("API-AUTH", apiAuth)
	if err != nil {
		log.Fatalf("init request: %s", err)
	}
	response, err := client.Do(reqest)
	if err != nil {
		log.Fatalf("request err: %s", err)
	}
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatalf("read response body: %s", err)
	}
	var resp service.Response[T]
	err1 := json.Unmarshal(body, &resp)
	if err1 != nil {
		log.Fatalf("json unmarshal: %s", err1)
	}
	return resp
}
