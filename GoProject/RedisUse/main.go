package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func main() {
	log_dir := getEnv("$LOG_DIR", "/root/log")
	log_name := "web_captcha.log"
	db := redis_init()
	forbid_key := "forbid_ip_set"
	white_ip_key := "cc_white_ip_set"
	ticker := time.Tick(time.Millisecond * 500)
	for range ticker {
		key_is_exists := get_key(white_ip_key)
		if key_is_exists {
			ip_set, _ := db.SInter(ctx, white_ip_key, forbid_key).Result()
			for _, k := range ip_set {
				cmd := fmt.Sprintf("iptables -nL|grep %s && iptables -D INPUT -s %s -j DROP", k, k)
				cmd_res := exec.Command("/bin/bash", "-c", cmd).Run()
				if cmd_res != nil {
					log.Printf("%s unforbid fail!", k)
				} else {
					log.Printf("%s unforbid success!", k)
					db.SRem(ctx, forbid_key, k).Result()
				}
			}
		} else {
			log.Print("white ip not found!")
		}
		log_file := strings.Join([]string{log_dir, log_name}, "/")
		clean := read_file(log_file)
		if clean {
			if _, err := db.Exists(ctx, forbid_key, white_ip_key).Result(); err == nil {
				db.Del(ctx, white_ip_key, forbid_key).Result()
				exec.Command("/bin/bash", "-c", "iptables -F").Run()
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

func redis_init() (rdb *redis.Client) {
	ip := getEnv("REDIS_AN_HOST", "localhost")
	port := getEnv("REDIS_AN_PORT", "6379")
	db := getEnv("REDIS_AN_DB", "5")
	// passwd := getEnv("REDIS_AN_PASSWORD", "j58degc69d")
	con := fmt.Sprintf("redis://%s:%s/%s", ip, port, db)
	if opt, err := redis.ParseURL(con); err == nil {
		rdb = redis.NewClient(opt)
	} else {
		panic(err)
	}
	return
}

func get_key(key string) (ex bool) {
	db := redis_init()
	if res, err := db.Exists(ctx, key).Result(); err == nil {
		if res > 0 {
			ex = true
		}
	} else {
		ex = false
	}
	return
}

func read_file(filename string) (res bool) {

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	var lastLine string
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lastLine = string(line)
	}
	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	st := strings.Split(lastLine, ",")[0]
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", st, loc)
	del_time := int(now.Sub(t).Minutes())
	if del_time > 20 {
		res = true
	} else {
		res = false
	}
	return
}
