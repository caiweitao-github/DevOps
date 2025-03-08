package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"util"
)

var (
	userName = ""
	passWd   = ""
)

type node struct {
	Data struct {
		ProxyList []string `json:"proxy_list"`
	} `json:"data"`
}

func getDomain() string {
	checkUrl := []string{"https://www.baidu.com/", "https://www.qq.com/", "https://www.bytedance.com/", "https://www.csdn.net/", "https://www.feishu.cn/", "https://www.aliyun.com/", "https://cloud.tencent.com/"}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rng.Intn(len(checkUrl))
	domain := checkUrl[randomIndex]
	return domain
}

func fetchIP() (node, error) {
	fetchUrl := ""
	resp, err := http.Get(fetchUrl)
	if err != nil {
		return node{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return node{}, err
	}
	var response node
	err1 := json.Unmarshal([]byte(body), &response)
	if err1 != nil {
		return node{}, err
	}
	return response, nil
}

func multipleFetchIPAsync(times int) (node, error) {
	var wg sync.WaitGroup
	wg.Add(times)

	var results node
	resultCh := make(chan node, times)
	errCh := make(chan error, times)

	for i := 0; i < times; i++ {
		go func() {
			result, err := fetchIP()
			if err != nil {
				errCh <- err
			} else {
				resultCh <- result
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	for result := range resultCh {
		results.Data.ProxyList = append(results.Data.ProxyList, result.Data.ProxyList...)
	}

	for err := range errCh {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}
func CheckDomain(data string) {
	node := strings.Split(data, ":")
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(userName, passWd),
		Host:   fmt.Sprintf("%s:%s", node[0], node[1]),
	}
	// checkUrl := "https://chrome-web.com/ChromeHSetup.exe"
	checkUrl := getDomain()
	request, _ := http.NewRequest(http.MethodGet, checkUrl, nil)
	// request.Header.Add("Accept-Encoding", "gzip")
	request.Header.Add("User-Agent", util.GetUA())
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
		// fmt.Printf("%s:%s err %s", node[0], node[1], err)

	} else {
		resp.Body.Close()
	}
}

func main() {
	// times := 10
	// result, err := multipleFetchIPAsync(times)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(len(result.Data.ProxyList))
	for {
		times := 10
		result, err := multipleFetchIPAsync(times)
		if err != nil {
			panic(err)
		}
		for _, n := range result.Data.ProxyList {
			for range 10 {
				go CheckDomain(n)
			}
		}
	}
}
