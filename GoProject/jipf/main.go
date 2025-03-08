package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
	"util"
)

var Db *sql.DB

var key = ""

var auth = ""

var Encryption = md5.Sum([]byte(key + "_" + auth))

var apiAuth = hex.EncodeToString(Encryption[:])

var masterDomain = "https://test.com/"

var (
	logFile  = "/data/kdl/log/get_jipf_data.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

type JipF struct {
	Country    string `json:"country_code"`
	ProviderID string `json:"provider_id"`
}

type JipFCount struct {
	Country string `json:"country_code"`
	Count   int    `json:"count"`
}

func init() {
	var err error
	Db, err = util.JipFDB()
	if err != nil {
		logError.Printf("ConnDb Failed : %s", err)
	}
}
func main() {
	defer func(pre time.Time) {
		elapsed := time.Since(pre)
		logInfo.Printf("elapsed: %v", elapsed)
	}(time.Now())

	logInfo.Println("run..")
	data, err := getJipFData()
	if err != nil {
		logError.Printf("getJipFData data err: %v", err)
	}
	data1, err := getJipFCountData()
	if err != nil {
		logError.Printf("getJipFCountData data err: %v", err)
	}
	reportDate(data, data1)
}

func getJipFData() (data []JipF, err error) {
	dataList := make([]JipF, 0, 150)
	sqlStr := "select country_code,provider_id from jip_node where status = 1"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var d JipF
		err := rows.Scan(&d.Country, &d.ProviderID)
		if err != nil {
			return nil, err
		}
		dataList = append(dataList, d)
	}
	return dataList, nil
}

func getJipFCountData() (data []JipFCount, err error) {
	dataList := make([]JipFCount, 0, 50)
	sqlStr := "select country_code,count(*) as count from jip_node where status=1 group by country_code"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var d JipFCount
		err := rows.Scan(&d.Country, &d.Count)
		if err != nil {
			return nil, err
		}
		dataList = append(dataList, d)
	}
	return dataList, nil
}

func reportDate(data []JipF, data1 []JipFCount) {
	domain, _ := url.Parse(masterDomain)
	domain.Path = "/reportjipfdata"
	payload := struct {
		JipFList      []JipF      `json:"jipf_list"`
		JipFCountList []JipFCount `json:"jipf_count_list"`
	}{
		JipFList:      data,
		JipFCountList: data1,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logError.Fatalf("json serialization error: %s", err)
	}
	req, err := http.NewRequest(http.MethodPost, domain.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		logError.Fatalf("init request: %s", err)
	}

	req.Header.Add("API-AUTH", apiAuth)

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
