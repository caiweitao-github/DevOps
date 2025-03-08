package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"opsServer/service"
	"sync"
	"time"
	"util"
)

var key = ""

var auth = ""

var Encryption = md5.Sum([]byte(key + "_" + auth))

var apiAuth = hex.EncodeToString(Encryption[:])

const LineNum = 1

var (
	masterDomain = "https://test.com/"
	checkUrl     = "http://dushu.baidu.com"
	userName     = ""
	passWd       = ""
)

var log = util.NewInitLog("/data/kdl/log/transfer_server_monitor.log")

var wg sync.WaitGroup

func main() {
	var sleep time.Duration
	sleepTime := 60 * time.Second
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	for {
		start := time.Now()
		aliveList, deadList := []service.DataEntry{}, []service.DataEntry{}
		alive := make(chan service.DataEntry, 50)
		dead := make(chan service.DataEntry, 50)
		r := getNode[[]string]()
		for _, server := range *r.Data {
			wg.Add(1)
			go func(server string) {
				checkTransferServer(server, alive, dead)
			}(server)
		}

		go func() {
			wg.Wait()
			close(alive)
			close(dead)
		}()

		for deadNode := range dead {
			deadList = append(deadList, deadNode)
		}

		for aliveNode := range alive {
			aliveList = append(aliveList, aliveNode)
		}

		log.Infof("alive: %v, dead: %v\n", aliveList, deadList)
		reportData(aliveList, deadList)
		end := time.Since(start)
		if end <= time.Second*0 || end >= time.Second*30 {
			sleep = 3 * time.Second
		} else {
			sleep = sleepTime - end
		}
		log.Infof("Time Use: %v, sleep: %v\n", end, sleep)
		<-ticker.C
	}
}

func getNode[T service.Type]() service.Response[T] {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/gettransfer"
	client := &http.Client{}
	reqest, err := http.NewRequest(http.MethodGet, domain.String(), nil)
	reqest.Header.Add("API-AUTH", apiAuth)
	if err != nil {
		log.Fatalf("init request: %s", err)
	}
	response, _ := client.Do(reqest)
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

func getLine[T service.Type](d service.Parameter) service.Response[T] {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/getline"
	payloadBytes, err := json.Marshal(d)
	if err != nil {
		log.Fatalf("json serialization error: %s", err)
	}
	client := &http.Client{}
	reqest, err := http.NewRequest(http.MethodPost, domain.String(), bytes.NewBuffer(payloadBytes))
	reqest.Header.Add("API-AUTH", apiAuth)
	if err != nil {
		log.Fatalf("init request: %s", err)
	}
	response, _ := client.Do(reqest)
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

func checkDomain(n service.Result, alive chan service.DataEntry, dead chan service.DataEntry) {
	defer wg.Done()
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(userName, passWd),
		Host:   fmt.Sprintf("%s:%s", n.NodeIP, n.NodePort),
	}
	request, _ := http.NewRequest("GET", checkUrl, nil)
	request.Header.Add("Accept-Encoding", "gzip")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 8,
	}
	resp, err := client.Do(request)
	if err != nil {
		dead <- service.DataEntry{
			IP:   n.NodeIP,
			Code: n.ServerCode,
			Log:  err.Error(),
		}
	} else {
		alive <- service.DataEntry{
			IP:   n.NodeIP,
			Code: n.ServerCode,
			Log:  "nil",
		}
		resp.Body.Close()
	}
}

func checkTransferServer(ip string, alive chan service.DataEntry, dead chan service.DataEntry) {
	d := service.Parameter{
		ServerIP: ip,
		Num:      LineNum,
	}
	r := getLine[[]service.Result](d)
	if r.Data != nil {
		for _, v := range *r.Data {
			go checkDomain(v, alive, dead)
		}
	} else {
		log.Infof("%s transfer server line is empty.\n", ip)
		wg.Done()
	}
}

func reportData(aliveList, deadList []service.DataEntry) {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/report"
	payload := service.NodeList{
		Category:  "TFS",
		AliveList: aliveList,
		DeadList:  deadList,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("json serialization error: %s", err)
	}
	req, err := http.NewRequest(http.MethodPost, domain.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatalf("init request: %s", err)
	}

	req.Header.Add("API-AUTH", apiAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("network error: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		log.Info("Report success.")
	} else {
		log.Warn("Report failed.")
	}
}
