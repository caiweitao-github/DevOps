package fpsutil

import (
	"Aliyun"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"util"
)

var maxprocess = 30 >> 1

func (f fpsHa) UpdataLoginIP() {
	tpsip := rdb.HGet(ctx, f.code, "ip").Val()
	if tpsip != f.ip {
		rdb.HSet(ctx, f.code, "ip", f.ip)
	}
}

func (f fpsHa) CheckFps() {
	keyIsExists := rdb.Exists(ctx, f.code).Val()
	if keyIsExists == 0 {
		if f.status == 1 {
			rdb.HMSet(ctx, f.code, map[string]interface{}{"id": f.id, "ip": f.ip, "status": f.status, "ha": 0, "ha_ip": "None"})
		} else {
			rdb.HMSet(ctx, f.code, map[string]interface{}{"id": f.id, "ip": f.ip, "status": 0, "ha": 0, "ha_ip": "None"})
			util.SendMess([]string{fmt.Sprintf("%s Down! ------> %d次", f.code, fpsStatus[0])})
		}
	} else {
		domains := rdb.SMembers(ctx, fmt.Sprintf("%s_%s", f.code, keys)).Val()
		FpsStatus, _ := strconv.Atoi(rdb.HGet(ctx, f.code, "status").Val())
		HaStatus, _ := strconv.Atoi(rdb.HGet(ctx, f.code, "ha").Val())
		if f.status != FpsStatus && f.status == 3 && FpsStatus > -2 {
			rdb.HIncrBy(ctx, f.code, "status", -1)
			if HaStatus == 0 {
				util.SendMess([]string{fmt.Sprintf("%s Down! ------> %d次", f.code, fpsStatus[FpsStatus-1])})
			}
			CachaStatus, _ := strconv.Atoi(rdb.HGet(ctx, f.code, "status").Val())
			if CachaStatus == -2 && HaStatus == 0 && f.ip != FPSHA {
				util.SendMess([]string{fmt.Sprintf("%s 连续3次检测为异常, 切换所有域名至备用机.", f.code)})
				f.UpdataDomain(FPSHA, 600, domains...)
				util.SendMess([]string{fmt.Sprintf("%s 域名: %s 切换至备用机 %s", f.code, strings.Join(domains, "\u0020"), FPSHA)})
				rdb.HMSet(ctx, f.code, map[string]interface{}{"ha": 1, "ha_ip": FPSHA})
			}
		} else if f.status != FpsStatus && f.status == 1 {
			rdb.HIncrBy(ctx, f.code, "status", 1)
			HaStatus, _ := strconv.Atoi(rdb.HGet(ctx, f.code, "ha").Val())
			HaIp := rdb.HGet(ctx, f.code, "ha_ip").Val()
			if FpsStatus == 0 && HaStatus == 1 && f.ip != FPSHA && HaIp != "None" {
				f.UpdataDomain(f.ip, 600, domains...)
				rdb.HMSet(ctx, f.code, map[string]interface{}{"ha": 0, "ha_ip": "None"})
				util.SendMess([]string{fmt.Sprintf("%s 恢复正常, 切回所有域名: %s", f.code, strings.Join(domains, "\u0020"))})
			}
		}
	}
}

func (f fpsHa) UpdataDomain(ip string, TTL int64, domains ...string) {
	ch := make(chan struct{}, maxprocess)
	done := make(chan bool)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	re := regexp.MustCompile(`\.kdlfps.com`)
	go func() {
		defer close(ch)
		defer close(done)
		for i := 0; i < len(domains); i++ {
			<-done
		}
		done <- true
	}()
	for _, domain := range domains {
		ch <- struct{}{}
		go func(domain string) {
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(rng.Intn(801)+500))
				<-ch
				done <- true
			}()
			d := re.Split(domain, 2)[0]
			err := Aliyun.UpdateFpsDomain(d, Domain, "A", ip, TTL)
			if err != nil {
				logError.Printf("Update %s Domain Failed: %s", d, err)
			} else {
				logInfo.Printf("Updata Domain %s.%s ------> %s.", d, Domain, ip)
			}
		}(domain)
	}
	<-done
}

func (f fpsHa) GetFpsDomain() error {
	domainList := make([]string, 0)
	var sqlStr string
	var localtionId string
	var label int8

	switch f.locationCcode {
	case "us":
		sqlStr = "select domain from fps_domain where fps_us_id = ? and status != 4"
		localtionId = "us"
		label = 1
	case "as":
		sqlStr = "select domain from fps_domain where fps_as_id = ? and status != 4"
		localtionId = "as"
	}
	rows, err := Db.Query(sqlStr, f.id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var domain string
		err := rows.Scan(&domain)
		if err != nil {
			logError.Printf("scan failed, err:%v\n", err)
			return err
		}
		subDomain := fmt.Sprintf("%s.%s", localtionId, domain)
		domainList = append(domainList, subDomain)
		rdb.SAdd(ctx, fmt.Sprintf("%s_%s", f.code, keys), subDomain)
		if label == 1 {
			domainList = append(domainList, domain)
			rdb.SAdd(ctx, fmt.Sprintf("%s_%s", f.code, keys), domain)
		}
	}
	cacheKey := fmt.Sprintf("%s_%s", f.code, keys)
	cacheDomain, err := rdb.SMembers(ctx, cacheKey).Result()
	if err != nil {
		return err
	}
	dbdata := strings.Join(domainList, ",")
	for _, v := range cacheDomain {
		if !strings.Contains(dbdata, v) {
			rdb.SRem(ctx, cacheKey, v)
		}
	}
	return nil
}
