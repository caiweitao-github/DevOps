package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
	"util"
)

var (
	checkUrl     = ""
	userName     = ""
	passWd       = ""
	masterDomain = ""
	apiToken     = ""
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
	Port   string `json:"port"`
}

func getPath() string {
	path := []string{""}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rng.Intn(len(path))
	data := path[randomIndex]
	return data
}

func getKps() Response {
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", masterDomain, nil)
	reqest.Header.Add("API-AUTH", apiToken)
	if err != nil {
		fmt.Println(err)
	}
	response, _ := client.Do(reqest)
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	var resu Response
	err = json.Unmarshal(body, &resu)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func checkDomain(node DataEntry) {
	// defer wg.Done()
	path := getPath()
	da := fmt.Sprintf("%s%s", checkUrl, path)
	request, _ := http.NewRequest("GET", da, nil)
	ua := util.GetUA()
	request.Header.Set("User-Agent", ua)
	request.Header.Add("Accept-Encoding", "gzip")
	proxyURL := &url.URL{
		Scheme: "http",
		User:   url.UserPassword(userName, passWd),
		Host:   fmt.Sprintf("%s:%s", node.IP, node.Port),
	}

	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 3,
	}
	resp, err := client.Do(request)
	if err != nil {
		// fmt.Printf("%s HTTP Error: %s", node.Code, err)
		return
	}
	defer resp.Body.Close()
}

func main() {
	for {
		resu := getKps()
		for _, i := range resu.Data {
			go checkDomain(i)
		}
	}

}
