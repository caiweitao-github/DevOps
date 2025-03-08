package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	Check_url, username, passwd, master_domain, api_token = "", "", "", "", ""
)

type Response struct {
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
	Data []DataEntry `json:"data"`
}

type DataEntry struct {
	Status           int    `json:"status"`
	Code             string `json:"code"`
	IP               string `json:"ip"`
	LastChangeIPTime string `json:"last_changeip_time"`
	Port             string `json:"port"`
	ChangeIPPeriod   int    `json:"changeip_period"`
}

type Timeout_Kps struct {
	data []Node_Info
}

type Node_Info struct {
	Code, IP string
}

func send_mess(msgs []string) {
	apiurl := ""
	contentType := "application/json"
	var content []interface{}
	for _, m := range msgs {
		content = append(content, []interface{}{
			map[string]interface{}{
				"tag":  "text",
				"text": m,
			},
		})
	}
	sendData := map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   "[站点响应检测]",
					"content": content,
				},
			},
		},
	}
	jsonData, err := json.Marshal(sendData)
	if err != nil {
		log.Printf("json marshal failed, err:%v\n", err)
		return
	}
	result, err := http.Post(apiurl, contentType, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("post failed, err:%v\n", err)
		return
	}
	defer result.Body.Close()
}

func get_kps_info() Response {
	api_auth := api_token + "_" + ""
	md5Sum := md5.Sum([]byte(api_auth))
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", master_domain, nil)
	reqest.Header.Add("API-AUTH", hex.EncodeToString(md5Sum[:]))
	if err != nil {
		panic(err)
	}
	response, _ := client.Do(reqest)
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		panic(err)
	}
	var resu Response
	err = json.Unmarshal(body, &resu)
	if err != nil {
		panic(err)
	}
	return resu
}

func check_domain(check_url string, node DataEntry, ti *Timeout_Kps, wg *sync.WaitGroup) {
	start := time.Now()
	defer wg.Done()
	st := fmt.Sprintf("http://%s:%s@%s:%s", username, passwd, node.IP, node.Port)
	proxy_url, _ := url.Parse(st)
	request, _ := http.NewRequest("GET", check_url, nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.54 Safari/537.36")

	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy_url),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 3,
	}
	resp, err := client.Do(request)
	end := float64(time.Since(start).Seconds())
	log.Printf("%s耗时：%fs", node.Code, end)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Println(netErr)
			ti.data = append(ti.data, Node_Info{node.Code, node.IP})
			return
		}
		return
	}
	defer resp.Body.Close()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	kps_time_out := Timeout_Kps{}
	wg := sync.WaitGroup{}
	resu := get_kps_info()
	random_data := rand.Perm(len(resu.Data))
	random_num := 20
	for i := 1; i <= random_num; i++ {
		index := random_data[i]
		if resu.Data[index].Status == 1 {
			wg.Add(1)
			go check_domain(Check_url, resu.Data[index], &kps_time_out, &wg)
		}
	}
	wg.Wait()
	mess := make([]string, 0, 2)
	if len(kps_time_out.data) > 1 {
		mess = append(mess, "访问域名成功率低于95%!")
		for _, v := range kps_time_out.data {
			mess = append(mess, v.Code)
		}
	}
	if mess != nil {
		send_mess(mess)
	}
}
