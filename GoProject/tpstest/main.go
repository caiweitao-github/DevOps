package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type ChanType[T bool | struct{}] chan T

var successCounter uint32

var (
	checkUrl = "https://www.qq.com/"
	userName = ""
	passWd   = ""
	httpPort = ""
	host     = ""
)

func CheckDomain(ch chan struct{}, do chan bool) {
	defer func() {
		<-ch
		do <- true
	}()
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(userName, passWd),
		Host:   fmt.Sprintf("%s:%s", host, httpPort),
	}
	request, _ := http.NewRequest(http.MethodGet, checkUrl, nil)
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
		fmt.Println(err)
	} else {
		resp.Body.Close()
		atomic.AddUint32(&successCounter, 1)
	}
}

func main() {
	count := 10000
	ch := make(ChanType[struct{}], 2000)
	done := make(ChanType[bool])
	go func() {
		defer close(ch)
		defer close(done)
		for range count {
			<-done
		}
		done <- true
	}()
	for range count {
		ch <- struct{}{}
		go CheckDomain(ch, done)
	}
	<-done
	s := atomic.LoadUint32(&successCounter)
	r := float32(s) / float32(count)
	fmt.Printf("tps56 ---> 请求: %d 次, 成功率: %.2f", count, r*100)
}
