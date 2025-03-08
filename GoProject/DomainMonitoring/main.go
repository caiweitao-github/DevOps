package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"time"
	"util"
)

var (
	checkUrl = "https://www.db.com"
	// checkUrl     = "https://www.adasdas.com"
	userName     = ""
	passWd       = ""
	masterDomain = ""
	apiToken     = ""
	port         = ""
)

var (
	logFile  = "/data/kdl/log/devops/domain_monitor.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

type Response struct {
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
	Data []DataEntry `json:"data"`
}

type DataEntry struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
	IP     string `json:"ip"`
}

type Timeout struct {
	data []Node
}

type Node struct {
	Code, IP string
}

type Mess struct {
	Info []struct {
		Code        string
		IP          string
		StatusCode  int
		DNSTime     time.Duration
		ConnectTime time.Duration
		TLSTime     time.Duration
		Total       time.Duration
	}
}

func getKps() Response {
	// api_auth := apiToken + "_" + "kdl"
	// md5Sum := md5.Sum([]byte(api_auth))
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", masterDomain, nil)
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
	var resu Response
	err = json.Unmarshal(body, &resu)
	if err != nil {
		logError.Fatalf("json unmarshal: %s", err)
	}
	return resu
}

func checkDomain(checkUrl string, node DataEntry, mess *Mess, ti *Timeout, wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		dnsStart, dnsDone         time.Time
		connectStart, connectDone time.Time
		tlsStart, tlsDone         time.Time
	)
	request, _ := http.NewRequest("GET", checkUrl, nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36")
	request.Header.Add("Accept-Encoding", "gzip")
	trace := &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
			// getIP()
		},
		ConnectDone: func(network, addr string, err error) {
			connectDone = time.Now()
		},

		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone:  func(cs tls.ConnectionState, err error) { tlsDone = time.Now() },

		DNSStart: func(dsi httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(ddi httptrace.DNSDoneInfo) { dnsDone = time.Now() },
	}
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(userName, passWd),
		Host:   fmt.Sprintf("%s:%s", node.IP, port),
	}

	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}
	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

	start := time.Now()
	resp, err := client.Do(request)
	elapsed := time.Since(start)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			ti.data = append(ti.data, Node{node.Code, node.IP})
			logError.Fatalf("%s HTTP Timeout: %s", node.Code, err)
		} else {
			logError.Fatalf("%s HTTP Error: %s", node.Code, err)
		}
	} else {
		defer resp.Body.Close()
		fmt.Println()
		data := struct {
			Code        string
			IP          string
			StatusCode  int
			DNSTime     time.Duration
			ConnectTime time.Duration
			TLSTime     time.Duration
			Total       time.Duration
		}{
			Code:        node.Code,
			IP:          node.IP,
			StatusCode:  resp.StatusCode,
			DNSTime:     dnsDone.Sub(dnsStart),
			ConnectTime: connectDone.Sub(connectStart),
			TLSTime:     tlsDone.Sub(tlsStart),
			Total:       elapsed,
		}
		mess.Info = append(mess.Info, data)
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	timeOut := Timeout{}
	mess := Mess{}
	wg := sync.WaitGroup{}
	resu := getKps()
	random_data := rand.Perm(len(resu.Data))
	random_num := 1
	for i := 1; i <= random_num; i++ {
		index := random_data[i]
		if resu.Data[index].Status == 1 {
			wg.Add(1)
			go checkDomain(checkUrl, resu.Data[index], &mess, &timeOut, &wg)
		}
	}
	wg.Wait()
	fmt.Println(mess.Info)
	// if len(timeOut.data) > 1 {
	// 	util.SendMess([]string{""})
	// }
}
