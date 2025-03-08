package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"opsServer/service"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"util"

	"github.com/go-redis/redis/v8"
)

var key = ""

var auth = ""

var Encryption = md5.Sum([]byte(key + "_" + auth))

var apiAuth = hex.EncodeToString(Encryption[:])

var apiDomain = "https://test.com/"

var (
	logFile  = "/root/log/forbid.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

const (
	ipAccessKey       = "IpAccessMap"
	forbidKey         = "forbidMap"
	warningKey        = "warningMap"
	notifyWarningKeys = "notifyWarningSet"
	notifyforbidKeys  = "notifyforbidSet"
)

var notifyUrl = "https://open.feishu.cn/open-apis/bot/v2/hook/c6f343b3-f8e8-4e81-b27d-ae20bd31c495"

var title = "[前端API超频通知]"

var domainName = "https://dev.kdlapi.com/"

var ctx = context.Background()

var rdb *redis.Client

type Body struct {
	OrderID   string `json:"orderid"`
	UserIP    string `json:"ip"`
	Callcount string `json:"callcount"`
	OverType  string `json:"over_type"`
}

func main() {
	sleepTime := 3 * time.Second
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	for {
		t := time.Now()
		logInfo.Println("run...")
		r := getUnForbidIP[[]string]()
		if r.Data != nil {
			for _, v := range *r.Data {
				unforbidIP(v)
			}
		}
		forbidProcess()
		logInfo.Printf("elapsed: %v\n", time.Since(t))
		<-ticker.C
	}
}

func init() {
	var err error
	rdb, err = util.RedisDB()
	if err != nil {
		logError.Printf("RedisInit Failed : %s", err)
	}
}

func forbidProcess() {
	warningIP := rdb.HGetAll(ctx, warningKey).Val()
	for k, v := range warningIP {
		isWarningNotify(k, v)
	}
	forIP := rdb.HGetAll(ctx, forbidKey).Val()
	for k, v := range forIP {
		isForbid(k, v)
	}
}

func isWarningNotify(ipData, con string) {
	d := strings.Split(ipData, ":")
	isNotify := rdb.SIsMember(ctx, notifyWarningKeys, d[1]).Val()
	if !isNotify {
		if res := checkorderID(d[0]); res {
			notify(d[0], d[1], con, "1", notifyWarningKeys, d[0])
		} else {
			if r := checksecretID(d[0]); r {
				if orderID := getOrderID(d[0]); orderID != "" {
					notify(orderID, d[1], con, "1", notifyWarningKeys, d[0])
				}
			}
		}
	}
}

func isForbid(ipData, con string) {
	d := strings.Split(ipData, ":")
	r := rdb.SIsMember(ctx, notifyforbidKeys, d[1]).Val()
	if !r {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("iptables -nL|grep -w %s &>/dev/null", d[1]))
		err := cmd.Run()
		if err != nil {
			if r := forbidIP(d[1]); r {
				logInfo.Printf("%s → %s req count %s, forbid success!", d[1], d[0], con)
			}
			if res := checkorderID(d[0]); res {
				notify(d[0], d[1], con, "2", notifyforbidKeys, d[0])
				reportForbidIP(d[1], con, d[2], "-4", d[0])
			} else {
				if r := checksecretID(d[0]); r {
					if orderID := getOrderID(d[0]); orderID != "" {
						notify(orderID, d[1], con, "2", notifyforbidKeys, d[0])
						reportForbidIP(d[1], con, d[2], "-4", orderID)
					}
				}
			}
		}
	}
}

func forbidIP(ip string) (res bool) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("iptables -I INPUT -s %s -j DROP", ip))
	err := cmd.Run()
	if err != nil {
		return
	}
	res = true
	return
}

func containsValue(s []string, target string) (r bool, order []string) {
	for _, value := range s {
		o := strings.Split(value, ":")[0]
		if o == target {
			r = true
			return
		}
		order = append(order, target)
	}
	return
}

func notify(order, ip, count, overType, key, originalStr string) {
	if order == "None" {
		logInfo.Printf("%s|%s|%s|%s not orderid.", order, ip, overType, count)
		rdb.SAdd(ctx, key, ip)
		util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("%s 超频 %s次, 未查询到订单号.", ip, count)})
		return
	}

	if overType == "2" {
		r := rdb.HKeys(ctx, warningKey).Val()
		if res, o := containsValue(r, originalStr); !res {
			rdb.SAdd(ctx, key, ip)
			util.FeiShuNotify(notifyUrl, title, []string{fmt.Sprintf("查询两次orderid不一致: %s | %v", originalStr, o)})
		}
		return
	}
	domain, _ := url.Parse(domainName)
	domain.Path = "/api/reportapioverreq"
	contentType := "application/json"
	data := []Body{}
	data = append(data, Body{order, ip, count, overType})
	payload, err := json.Marshal(data)
	if err != nil {
		logError.Printf("json marshal error: %v", err)
		return
	}
	reqest, err := http.Post(domain.String(), contentType, bytes.NewBuffer(payload))
	if err != nil {
		logError.Printf("req err: %v", err)
	} else {
		reqest.Body.Close()
		logInfo.Printf("%s|%s|%s|%s notify success.", order, ip, overType, count)
		rdb.SAdd(ctx, key, ip)
	}
}

func getOrderID(secretId string) string {
	domain, _ := url.Parse(apiDomain)
	domain.Path = "/queryorderid"
	data := struct {
		SecretId string `json:"secret_id"`
	}{
		secretId,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		logError.Printf("json marshal error: %v", err)
		return ""
	}
	req, err := http.NewRequest(http.MethodPost, domain.String(), bytes.NewBuffer(payload))
	if err != nil {
		logError.Printf("new req err: %v", err)
	}
	req.Header.Add("API-AUTH", apiAuth)
	client := &http.Client{}
	reqest, err := client.Do(req)
	if err != nil {
		logError.Printf("get orderid err: %v", err)
	} else {
		body, err := io.ReadAll(reqest.Body)
		defer reqest.Body.Close()
		if err != nil {
			logError.Fatalf("read response body: %s", err)
		}
		var resp service.Response[string]
		err1 := json.Unmarshal(body, &resp)
		if err1 != nil {
			logError.Fatalf("json unmarshal: %s", err1)
		}
		return *resp.Data
	}
	return ""
}

func checkorderID(str string) (r bool) {
	orderID, _ := regexp.Compile("^9[0-9]{14}$")
	if orderID.MatchString(str) {
		r = true
	} else {
		r = false
	}
	return
}

func checksecretID(str string) (r bool) {
	secretID, _ := regexp.Compile("^o[a-z0-9]{19}$")
	if secretID.MatchString(str) {
		r = true
	} else {
		r = false
	}
	return
}

func unforbidIP(ip string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("iptables -nL|grep -w %s &>/dev/null", ip))
	err := cmd.Run()
	if err == nil {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("iptables -D INPUT -s %s -j DROP &>/dev/null", ip))
		err := cmd.Run()
		if err == nil {
			logInfo.Printf("unforbid %s success!", ip)
			rdb.HDel(ctx, ipAccessKey, ip)
			rdb.SRem(ctx, notifyWarningKeys, ip)
			rdb.SRem(ctx, notifyforbidKeys, ip)
			k1 := rdb.HKeys(ctx, forbidKey).Val()
			for _, key := range k1 {
				r := strings.Split(key, ":")[1]
				if r == ip {
					rdb.HDel(ctx, forbidKey, key)
				}
			}
			k2 := rdb.HKeys(ctx, warningKey).Val()
			for _, key := range k2 {
				r := strings.Split(key, ":")[1]
				if r == ip {
					rdb.HDel(ctx, warningKey, key)
				}
			}
		}
	} else {
		return err
	}
	return nil
}

func reportForbidIP(ip, count, forbidtime, reason, orderid string) (res bool) {
	hostname, _ := os.Hostname()
	domain, _ := url.Parse(apiDomain)
	domain.Path = "/reportforbidip"
	intTime, err := strconv.Atoi(forbidtime)
	if err != nil {
		logError.Printf("time format err: %v", err)
		return
	}
	data := service.NginxForbidData{
		Code: hostname,
		IP:   ip,
		Data: service.NginxData{
			Count:      count,
			ForbidTime: time.Unix(int64(intTime), 0).Format(time.DateTime),
			Reason:     reason,
			OrderID:    orderid,
		},
	}
	payload, err := json.Marshal(data)
	if err != nil {
		logError.Printf("json marshal error: %v", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, domain.String(), bytes.NewBuffer(payload))
	if err != nil {
		logError.Printf("new req err: %v", err)
		return
	}
	req.Header.Add("API-AUTH", apiAuth)
	client := &http.Client{}
	reqest, err := client.Do(req)
	if err != nil {
		logError.Printf("get orderid err: %v", err)
		return
	} else {
		reqest.Body.Close()
		res = true
	}
	return
}

func getUnForbidIP[T service.Type]() (res service.Response[T]) {
	ip := []string{}
	hostname, _ := os.Hostname()
	domain, _ := url.Parse(apiDomain)
	domain.Path = "/unforbidip"
	forbindList := rdb.HKeys(ctx, forbidKey).Val()
	for _, v := range forbindList {
		d := strings.Split(v, ":")
		r, err := isUnForbid(d[2])
		if err != nil {
			logError.Printf("is unforbid err: %v", err)
			continue
		}
		if r {
			err := unforbidIP(d[1])
			if err != nil {
				logError.Printf("unforbid err: %v", err)
			} else {
				ip = append(ip, d[1])
			}
		}
	}
	data := service.ReportIP{
		Code:   hostname,
		IpList: ip,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		logError.Fatalf("json marshal error: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, domain.String(), bytes.NewBuffer(payload))
	if err != nil {
		logError.Fatalf("new req err: %v", err)
	}
	req.Header.Add("API-AUTH", apiAuth)
	client := &http.Client{}
	reqest, err := client.Do(req)
	if err != nil {
		logError.Fatalf("get unforbidip err: %v", err)
	}
	defer reqest.Body.Close()
	body, err := io.ReadAll(reqest.Body)
	if err != nil {
		logError.Fatalf("read body err: %v", err)
	}
	err1 := json.Unmarshal(body, &res)
	if err1 != nil {
		logError.Fatalf("json unmarshal: %s", err1)
	}
	return
}

func isUnForbid(t string) (r bool, e error) {
	intTime, err := strconv.Atoi(t)
	if err != nil {
		e = err
		return
	}
	unixTimestamp := int64(intTime)
	forbidTime := time.Unix(unixTimestamp, 0)
	s := int(time.Since(forbidTime.Add(4 * time.Hour)).Seconds())
	if s >= 0 {
		r = true
	}
	return
}
