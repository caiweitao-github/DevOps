package fpslib

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"util"
)

var (
	logFile  = "/data/kdl/log/G2pfpsmonitor.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

var (
	masterDomain = ""
	apiToken     = ""
)

type DataEntry struct {
	IP       string `json:"ip"`
	Code     string `json:"code"`
	Port     string `json:"port"`
	Location string `json:"location"`
}

type NodeList struct {
	Node []DataEntry `json:"node_list"`
}

type Response struct {
	Data NodeList `json:"data"`
}

func GetFpsNode() NodeList {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/getmonitorfpslist"
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", domain.String(), nil)
	reqest.Header.Add("API-AUTH", apiToken)
	if err != nil {
		logError.Fatalf("init request: %s", err)
	}

	response, err := client.Do(reqest)
	if err != nil {
		logError.Fatalf("client.Do: %s", err)
	}
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

func Fpsreport(aliveList []DataEntry, deadList []DataEntry) {
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
		logError.Printf("network error: %s", err)
	} else {
		if resp.StatusCode == 200 {
			logInfo.Println("Report success.")
		} else {
			logError.Println("Report failed.")
		}
	}
	defer resp.Body.Close()
}
