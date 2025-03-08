package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"util"
)

type Dblocation struct {
	code, ip, port     string
	changeip_period    int
	last_changeip_time string
	location           string
}

var Db *sql.DB

var (
	username = ""
	passwd   = ""
)

var (
	logFile  = "/data/kdl/log/devops/dps_location_check.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

var checkUrl = "http://myip.top"

type Response struct {
	IP       string `json:"ip"`
	Province string `json:"province"`
	City     string `json:"city"`
}

var newTime = time.Now().Add(-time.Duration(30) * time.Second)

func main() {
	mess := []string{}
	data := make(chan string, 5000)
	node := getDb()
	done := make(chan struct{})
	numWorkers := len(node)

	go func() {
		for i := 0; i < numWorkers; i++ {
			<-done
		}
		close(data)
	}()

	for _, no := range node {
		go func(no Dblocation) {
			checkDpsLocation(no, checkUrl, data)
			done <- struct{}{}
		}(no)
	}

	for info := range data {
		mess = append(mess, info)
	}
	if len(mess) > 0 {
		util.SendMess(mess)
	}
	logInfo.Printf("Totle To Be Detected Dps: %d", len(node))

}

func init() {
	var err error
	Db, err = util.ConnDb("", "", "")
	if err != nil {
		logError.Printf("ConnDb Failed : %s", err)
	}
}

func getDb() []Dblocation {
	sqlStr := "select code,ip,port,changeip_period,last_changeip_time,location from dps where dps_type not in (2,6) and status = 1"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logError.Printf("Query Failed : %s", err)
	}
	info := make([]Dblocation, 0)
	defer rows.Close()
	for rows.Next() {
		var dps Dblocation
		err := rows.Scan(&dps.code, &dps.ip, &dps.port, &dps.changeip_period, &dps.last_changeip_time, &dps.location)
		if err != nil {
			logError.Printf("Scan Failed : %s", err)
		}
		last_changeip_time, err := time.Parse("2006-01-02 15:04:05", dps.last_changeip_time)
		if err != nil {
			logError.Fatalf("time.Parse Failed : %s", err)
		}
		if last_changeip_time.Add(time.Duration(dps.changeip_period) * time.Second).Before(newTime) {
			continue
		} else {
			info = append(info, dps)
		}
	}
	return info
}

func checkDpsLocation(node Dblocation, check_url string, data chan string) {
	var lo string
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(username, passwd),
		Host:   fmt.Sprintf("%s:%s", node.ip, node.port),
	}
	// proxy_url, _ := url.Parse(proxyURL)
	request, _ := http.NewRequest("GET", check_url, nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}
	resp, err := client.Do(request)
	if err != nil {
		logError.Printf("client.Do Failed : %s", err)
	} else {
		defer resp.Body.Close()
		body, errors := io.ReadAll(resp.Body)
		if errors != nil {
			logError.Panicf("io.ReadAll Failed : %s", err)
			return
		}
		if len(body) == 0 {
			return
		}
		var resu Response
		err = json.Unmarshal(body, &resu)
		if err != nil {
			logError.Panicf("json.Unmarshal Failed : %s", err)
			return
		}
		if resu.Province == resu.City {
			lo = fmt.Sprintf("%s市", resu.Province)
		} else {
			lo = fmt.Sprintf("%s省%s市", resu.Province, resu.City)
		}
		if lo == "市" || lo == "-市" {
			return
		} else if lo != node.location {
			data <- fmt.Sprintf("%s(%s) ----> %s", node.code, node.location, lo)
		}
	}
}
