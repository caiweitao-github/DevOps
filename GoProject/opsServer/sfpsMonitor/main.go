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

var (
	masterDomain = "https://test.com/"
	checkUrl     = "https://www.google.com/"
	userName = ""
	passWd   = ""
)

var log = util.NewInitLog("/data/kdl/log/sfpsmonitor.log")

var wg sync.WaitGroup

func main() {
	var sleep time.Duration
	sleepTime := 30 * time.Second
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	for {
		start := time.Now()
		aliveList, deadList := []service.DataEntry{}, []service.DataEntry{}
		alive := make(chan service.DataEntry, 2000)
		dead := make(chan service.DataEntry, 2000)
		r := getNode[[]service.SfpsData]()
		for _, server := range *r.Data {
			wg.Add(1)
			go func(sfps service.SfpsData) {
				CheckDomain(sfps, alive, dead)
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
		log.Infof("alive: %d, dead: %v\n", len(aliveList), deadList)
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
	domain.Path = "/getmonitorsfps"
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

func CheckDomain(n service.SfpsData, alive chan service.DataEntry, dead chan service.DataEntry) {
	defer wg.Done()
	var auth *url.Userinfo
	if n.UserName != nil && n.PassWord != nil {
		auth = url.UserPassword(*n.UserName, *n.PassWord)
	} else {
		auth = url.UserPassword(userName, passWd)
	}
	proxyURL := &url.URL{
		Scheme: "http",
		User:   auth,
		Host:   fmt.Sprintf("%s:%s", n.IP, n.Port),
	}
	request, _ := http.NewRequest(http.MethodHead, checkUrl, nil)
	// request.Header.Add("API-AUTH", checkToken)
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
			IP:   n.IP,
			Code: n.Code,
			Log:  "nil",
		}
	} else {
		alive <- service.DataEntry{
			IP:   n.IP,
			Code: n.Code,
			Log:  "nil",
		}
		resp.Body.Close()
	}
}

func reportData(aliveList, deadList []service.DataEntry) {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/sfpsmonitorreport"
	payload := service.NodeList{
		Category:  "SFPS",
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
