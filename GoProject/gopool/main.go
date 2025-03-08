package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"util"
)

var (
	logFile  = "/data/kdl/log/fpsmonitor.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

var (
	checkUrl     = "http://www.google.com"
	bakUrl       = "https://www.amazon.com"
	userName     = ""
	passWd       = ""
	masterDomain = ""
	apiToken     = ""
	port         = ""
	retry        = 1
)

type DataEntry struct {
	IP   string `json:"ip"`
	Code string `json:"code"`
	Port string `json:"port"`
}

type NodeList struct {
	Node []DataEntry `json:"node_list"`
}

type Response struct {
	Data NodeList `json:"data"`
}

type FpsCheck interface {
	checkDomain(checkUrl string, alive chan DataEntry, dead chan DataEntry, retry int)
}

func main() {
	alive := make(chan DataEntry, 50)
	dead := make(chan DataEntry, 50)
	aliveList, deadList := []DataEntry{}, []DataEntry{}
	node := getFpsNode()
	done := make(chan struct{})
	ch := make(chan bool)
	go func() {
		for i := 0; i < len(node.Node); i++ {
			<-done
		}
		close(alive)
		close(dead)
		ch <- true
	}()

	for _, fps := range node.Node {
		var FPS FpsCheck = fps
		go func(FPS FpsCheck) {
			FPS.checkDomain(checkUrl, alive, dead, retry)
			done <- struct{}{}
		}(FPS)
	}

	go func() {
		for v := range alive {
			aliveList = append(aliveList, v)
		}
	}()

	go func() {
		for v := range dead {
			deadList = append(deadList, v)
		}
	}()
	<-ch
	fmt.Printf("alive: %v, dead: %v", aliveList, deadList)
	report(aliveList, deadList)
}

func report(aliveList []DataEntry, deadList []DataEntry) {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/fpsmonitorreport"
	payload := struct {
		AliveList []DataEntry `json:"alive_fps_node_list"`
		DeadList  []DataEntry `json:"dead_fps_node_list"`
	}{
		AliveList: aliveList,
		DeadList:  deadList,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logError.Fatalf("json serialization error: %s", err)
	}
	req, err := http.NewRequest("POST", domain.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		logError.Fatalf("init request: %s", err)
	}

	req.Header.Add("API-AUTH", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logError.Fatalf("network error: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		logInfo.Println("Report success.")
	} else {
		logError.Println("Report failed.")
	}
}

func getFpsNode() NodeList {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/getmonitorfpslist"
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", domain.String(), nil)
	reqest.Header.Add("API-AUTH", apiToken)
	if err != nil {
		logError.Fatalf("init request: %s", err)
	}

	response, _ := client.Do(reqest)
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		logError.Fatalf("read response body: %s", err)
	}

	var resp Response
	err1 := json.Unmarshal(body, &resp)
	if err1 != nil {
		logError.Fatalf("json unmarshal: %s", err1)
	}
	return resp.Data
}

func (d DataEntry) checkDomain(checkUrl string, alive chan DataEntry, dead chan DataEntry, retry int) {
	if retry >= 2 {
		dead <- d
		return
	}
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(userName, passWd),
		Host:   fmt.Sprintf("%s:%s", d.IP, port),
	}
	request, _ := http.NewRequest("GET", checkUrl, nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36")
	request.Header.Add("Accept-Encoding", "gzip")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 15,
	}
	resp, err := client.Do(request)
	if err != nil {
		logError.Printf("%s error: %s", d.Code, err)
		logInfo.Printf("%s retry: %d", d.Code, retry)
		d.checkDomain(bakUrl, alive, dead, retry)
		retry++
	} else {
		alive <- d
		resp.Body.Close()
	}
}
