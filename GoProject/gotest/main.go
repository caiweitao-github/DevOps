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

var Db *sql.DB

var (
	username = ""
	passwd   = ""
)

var logInfo, logError = util.InitLog("/root/dps_test/dpstest.log", "[INFO] ", "[ERROR] ")

var checkUrl = "http://myip.top"

var fetchUrl = ""

type Response struct {
	IP       string `json:"ip"`
	Province string `json:"province"`
	City     string `json:"city"`
	Isp      string `json:"isp"`
}

type Data struct {
	Count     int      `json:"count"`
	ProxyList []string `json:"proxy_list"`
}

type Resp struct {
	Code int  `json:"code"`
	Data Data `json:"data"`
}

func main() {
	ticker := time.NewTicker(time.Minute)
	done := make(chan struct{})
	defer ticker.Stop()
	for {
		data := make(chan string, 150)
		start := time.Now()
		node := getDps()
		numWorkers := len(node.Data.ProxyList)
		logInfo.Printf("fetch dps node: %d", numWorkers)
		go func() {
			for i := 0; i < numWorkers; i++ {
				<-done
			}
			close(data)
		}()

		for _, no := range node.Data.ProxyList {
			go func(no string) {
				checkDpsLocation(no, checkUrl, data)
				done <- struct{}{}
			}(no)
		}

		for info := range data {
			go insertDb(info)
		}
		end := time.Since(start)
		logInfo.Printf("Time Use: %v\n", end)
		<-ticker.C
	}

}

func init() {
	var err error
	Db, err = util.ConnDb("root", "", "tunnel_test")
	if err != nil {
		logError.Printf("ConnDb Failed : %s", err)
	}
}

func getDps() (resu Resp) {
	request, _ := http.NewRequest("GET", fetchUrl, nil)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := client.Do(request)
	if err != nil {
		logError.Printf("client.Do Failed : %s", err)
	} else {
		defer resp.Body.Close()
		body, errors := io.ReadAll(resp.Body)
		if errors != nil {
			logError.Printf("io.ReadAll Failed : %s", err)
		}
		err = json.Unmarshal(body, &resu)
		if err != nil {
			logError.Printf("json.Unmarshal Failed : %s", err)
		}
	}
	return resu
}

func insertDb(sqlStr string) {
	_, err := Db.Exec(sqlStr)
	if err != nil {
		logError.Printf("Exec Failed : %s", err)
	}
}

func checkDpsLocation(node string, checkUrl string, data chan string) {
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(username, passwd),
		Host:   node,
	}
	request, _ := http.NewRequest("GET", checkUrl, nil)
	request.Header.Set("User-Agent", util.GetUA())
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
			logError.Printf("io.ReadAll Failed : %s", err)
		}
		var resu Response
		err = json.Unmarshal(body, &resu)
		if err != nil {
			logError.Printf("json.Unmarshal Failed : %s", err)
		} else {
			data <- fmt.Sprintf("INSERT INTO dpstest_20231106 (proxy_ip, ip, province, city, isp)  VALUES ('%s', '%s', '%s', '%s', '%s')", node, resu.IP, resu.Province, resu.City, resu.Isp)
		}
	}
}
