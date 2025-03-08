package main

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
	"util"
)

type ChanType[T bool | struct{}] chan T

var successCounter uint32

var checkUrl = "http://www.baidu.com"

var log = util.NewInitLog("/data/kdl/log/test/checkIP.log")

func newProxyPool() []string {
	proxyIP := make([]string, 0, 1000)
	for i := 10000; i < 13000; i++ {
		proxyIP = append(proxyIP, "ip"+":"+strconv.Itoa(i))
	}
	return proxyIP
}

func CheckDomain(ch chan struct{}, do chan bool, proxy string) {
	defer func() {
		<-ch
		do <- true
	}()
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   proxy,
	}
	request, _ := http.NewRequest(http.MethodGet, checkUrl, nil)
	request.Header.Add("Accept-Encoding", "gzip")
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
		log.Errorf("%s req fail, err: %v", proxy, err)
	} else {
		log.Infof("%s req success, ststus code: %d", proxy, resp.StatusCode)
		resp.Body.Close()
		atomic.AddUint32(&successCounter, 1)
	}
}

func main() {
	proxyIP := newProxyPool()
	ch := make(ChanType[struct{}], 500)
	done := make(ChanType[bool])
	for _, proxy := range proxyIP {
		ch <- struct{}{}
		go CheckDomain(ch, done, proxy)
	}

	for i := 0; i < len(proxyIP); i++ {
		<-done
	}

	s := atomic.LoadUint32(&successCounter)
	r := float32(s) / float32(len(proxyIP))
	log.Infof("EC-EIP ---> 请求: %d 次, 成功次数: %d 成功率: %.2f", len(proxyIP), s, r*100)
}
