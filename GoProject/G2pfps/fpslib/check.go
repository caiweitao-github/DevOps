package fpslib

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

var (
	checkUrl = "https://www.google.com/"
	bakUrl   = "https://www.amazon.com/"
	userName = ""
	passWd   = ""
	port     = ""
)

type FpsCheck interface {
	CheckDomain(checkUrl string, alive chan DataEntry, dead chan DataEntry, retry int)
	CheckSocks(checkUrl string, alive chan DataEntry, dead chan DataEntry, retry int)
}

func (d DataEntry) CheckDomain(checkUrl string, alive chan DataEntry, dead chan DataEntry, retry int) {
	if retry > 1 {
		dead <- d
		return
	}
	var proxyName string
	if d.Location == "as" {
		proxyName = userName + "-cont-as-period-0"
	} else {
		proxyName = userName + "-cont-nasa-period-0"
	}

	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(proxyName, passWd),
		Host:   fmt.Sprintf("%s:%s", d.IP, port),
	}
	request, _ := http.NewRequest(http.MethodHead, checkUrl, nil)
	request.Header.Add("Accept-Encoding", "gzip")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}
	resp, err := client.Do(request)
	if err != nil {
		logError.Printf("%s HTTP Error: %s", d.Code, err)
		if retry == 0 {
			logInfo.Printf("%s HTTP retry: %d", d.Code, retry+1)
		}
		d.CheckDomain(bakUrl, alive, dead, retry+1)
	} else {
		resp.Body.Close()
		alive <- d
	}
}

func (d DataEntry) CheckSocks(checkUrl string, alive chan DataEntry, dead chan DataEntry, retry int) {
	if retry > 1 {
		dead <- d
		return
	}
	var proxyName string
	if d.Location == "as" {
		proxyName = userName + "-cont-as-period-0"
	} else {
		proxyName = userName + "-cont-nasa-period-0"
	}

	auth := proxy.Auth{
		User:     proxyName,
		Password: passWd,
	}
	proxyStr := fmt.Sprintf("%s:%s", d.IP, port)
	dialer, err := proxy.SOCKS5("tcp", proxyStr, &auth, proxy.Direct)
	if err != nil {
		logError.Printf("%s error: %s", d.Code, err)
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: &http.Transport{Dial: dialer.Dial},
	}
	req, _ := http.NewRequest(http.MethodHead, checkUrl, nil)
	req.Header.Add("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		logError.Printf("%s SOCKS5 Error: %s", d.Code, err)
		if retry == 0 {
			logInfo.Printf("%s SOCKS5 retry: %d", d.Code, retry+1)
		}
		d.CheckSocks(bakUrl, alive, dead, retry+1)
	} else {
		resp.Body.Close()
		alive <- d
	}
}

func ProcessFpsNodes(nodes []DataEntry) ([]DataEntry, []DataEntry) {
	m1 := make(map[DataEntry]struct{})
	m2 := make(map[DataEntry]struct{})
	var wg sync.WaitGroup
	aliveList, deadList := []DataEntry{}, []DataEntry{}
	alive := make(chan DataEntry, 50)
	dead := make(chan DataEntry, 50)

	for _, fps := range nodes {
		wg.Add(1)
		go func(fps DataEntry) {
			defer wg.Done()
			var FPS DataEntry = fps
			FPS.CheckDomain(checkUrl, alive, dead, 0)
			FPS.CheckSocks(checkUrl, alive, dead, 0)
		}(fps)
	}

	go func() {
		wg.Wait()
		close(alive)
		close(dead)
	}()

	for deadNode := range dead {
		if _, exists := m2[deadNode]; !exists {
			deadList = append(deadList, deadNode)
			m2[deadNode] = struct{}{}
		}
	}

	for aliveNode := range alive {
		_, ex1 := m2[aliveNode]
		_, ex2 := m1[aliveNode]
		if !ex1 && !ex2 {
			aliveList = append(aliveList, aliveNode)
			m1[aliveNode] = struct{}{}
		}
	}

	return aliveList, deadList
}
