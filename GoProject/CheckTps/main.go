package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
	"util"
)

var (
	authKey     = ""
	domainName  = ""
	checkDomain = "https://www.baidu.com"
	proxyUser   = ""
	proxyPass   = ""
	checkNum    = 200
	wg          sync.WaitGroup
	notifyUrl   = ""
	port        = "15818"
	logFile     = "/data/kdl/log/checkTpsReq.log"
	logInfo     = util.LogConf(logFile, "[INFO] ")
	logError    = util.LogConf(logFile, "[ERROR] ")
)

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	NodeList []Node `json:"node_list"`
}

type Node struct {
	IP   string `json:"ip"`
	Code string `json:"code"`
}

func main() {
	defer func(pre time.Time) {
		elapsed := time.Since(pre)
		logInfo.Printf("process elapsed: %v", elapsed)
	}(time.Now())
	res := getTps()
	var w sync.WaitGroup
	for _, v := range res.NodeList {
		w.Add(1)
		go v.checkReq(&w)
	}
	w.Wait()
}

func (n Node) checkReq(w *sync.WaitGroup) {
	defer w.Done()
	proChan := make(chan struct{}, 3)
	var (
		errNum          uint32
		timeUse         uint64
		successNum      uint32
		errNumMutex     sync.Mutex
		timeUseMutex    sync.Mutex
		successNumMutex sync.Mutex
	)

	for range checkNum {
		proChan <- struct{}{}
		wg.Add(1)
		go n.reqDomain(&errNum, &successNum, &timeUse, &errNumMutex, &timeUseMutex, &successNumMutex, proChan)
	}
	wg.Wait()
	close(proChan)
	atomicErrNum := atomic.LoadUint32(&errNum)
	atomicSuccessNum := atomic.LoadUint32(&successNum)
	atomicTimeUse := atomic.LoadUint64(&timeUse)

	if atomicErrNum >= 50 {
		util.FeiShuNotify(notifyUrl, "[隧道请求耗时检测]", []string{fmt.Sprintf("%s 请求失败次数过多, 失败次数 -> %d", n.Code, atomicErrNum)})
	}

	if x := float64(atomicTimeUse) / float64(atomicSuccessNum); x > 3. {
		util.FeiShuNotify(notifyUrl, "[隧道请求耗时检测]", []string{fmt.Sprintf("%s 200次请求平均耗时为: %f!", n.Code, x)})
	}
}

func (n Node) reqDomain(e *uint32, s *uint32, t *uint64, eMutex *sync.Mutex, sMutex *sync.Mutex, tMutex *sync.Mutex, pro chan struct{}) {
	defer func() {
		wg.Done()
		<-pro
	}()
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(proxyUser, proxyPass),
		Host:   fmt.Sprintf("%s:%s", n.IP, port),
	}
	request, _ := http.NewRequest(http.MethodGet, checkDomain, nil)
	request.Header.Add("Accept-Encoding", "gzip")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 8,
	}
	start := time.Now()
	resp, err := client.Do(request)
	if err != nil {
		eMutex.Lock()
		*e += 1
		eMutex.Unlock()
	} else {
		resp.Body.Close()
		tMutex.Lock()
		*t += uint64(time.Since(start).Seconds())
		tMutex.Unlock()
		sMutex.Lock()
		*s += 1
		sMutex.Unlock()
	}
}

func getTps() Data {
	domain, _ := url.Parse(domainName)
	domain.Path = "/getmonitortpslist"
	client := &http.Client{}
	reqest, err := http.NewRequest(http.MethodGet, domain.String(), nil)
	if err != nil {
		logError.Fatalf("init request: %s", err)
	}
	reqest.Header.Add("API-AUTH", authKey)
	resp, err := client.Do(reqest)
	if err != nil {
		logError.Fatalf("request err: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logError.Fatalf("read body err: %v", err)
	}
	defer resp.Body.Close()
	var r Response
	err1 := json.Unmarshal(body, &r)
	if err1 != nil {
		logError.Fatalf("json unmarshal: %s", err1)
	}
	return r.Data
}
