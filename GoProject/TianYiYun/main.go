package TianYiYun

import (
	"TianYiYun/auth"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"util"
)

var baseDomain = "https://ctecs-global.ctapi.ctyun.cn"

type Region struct {
	RegionParent     string `json:"regionParent"`
	RegionID         string `json:"regionID"`
	RegionName       string `json:"regionName"`
	OpenapiAvailable bool   `json:"openapiAvailable"`
}

type ReturnObj struct {
	RegionList []Region `json:"regionList"`
}

type Response struct {
	ReturnObj ReturnObj `json:"returnObj"`
}

type NodeRegion struct {
	ExpiredTime string `json:"expiredTime"`
	InstanceID  string `json:"instanceID"`
	DisplayName string `json:"displayName"`
	FloatingIP  string `json:"floatingIP"`
}

type NodeReturnObj struct {
	RegionList []NodeRegion `json:"results"`
}

type NodeResponse struct {
	ReturnObj NodeReturnObj `json:"returnObj"`
}

type KpsData struct {
	location string
	ip       string
	provider string
}

var (
	Db       *sql.DB
	logFile  = "/data/kdl/log/opsServer/TYCloud.log"
	logError = util.LogConf(logFile, "[ERROR] ")
)

func init() {
	var err error
	Db, err = util.dbDB()
	if err != nil {
		logError.Printf("init db err: %v", err)
	}
}

func tianYApi(provider, urlStr, query, method string, params map[string]interface{}, headerParams map[string]string, contentType string) ([]byte, error) {
	if contentType == "application/x-www-form-urlencoded" {
		paramsEncoded := url.Values{}
		for k, v := range params {
			switch v.(type) {
			case string:
				paramsEncoded.Add(k, v.(string))
			default:
				logError.Printf("Can't handle type %v value %s\n", v, k)
			}
		}
		query = paramsEncoded.Encode()
	}

	queryParams := make(map[string]interface{})
	if method == "GET" {
		for k, v := range params {
			queryParams[k] = v
		}
	} else {
		if query != "" {
			for _, q := range strings.Split(query, "&") {
				kv := strings.Split(q, "=")
				if len(kv) == 2 {
					queryParams[kv[0]] = kv[1]
				}
			}
		}
	}

	headers := auth.GetSignHeaders(provider, queryParams, params, method, contentType)
	for k, v := range headerParams {
		headers[k] = v
	}
	reqURL := urlStr + "?" + auth.EncodeQueryStr(query)
	client := &http.Client{}

	var req *http.Request
	var err error

	if method == http.MethodGet {
		req, err = http.NewRequest(http.MethodGet, reqURL, nil)
	} else {
		reqBody, _ := json.Marshal(params)
		req, err = http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", contentType)
	}

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/110.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return bodyBytes, nil

}

func (k KpsData) getNodeID() ([]string, error) {
	regionIdList := make([]string, 0, 5)
	urlStr := baseDomain + "/v4/region/list-regions"
	res, err := tianYApi(k.provider, urlStr, "", http.MethodGet, nil, nil, "application/json;charset=UTF-8")
	if err != nil {
		return regionIdList, err
	}
	var response Response
	err = json.Unmarshal([]byte(res), &response)
	if err != nil {
		return regionIdList, err
	}
	for _, d := range response.ReturnObj.RegionList {
		if strings.HasPrefix(d.RegionName, k.location) && d.OpenapiAvailable {
			regionIdList = append(regionIdList, d.RegionID)
		}
	}
	return regionIdList, nil
}

func (k KpsData) getNode(regid []string) (string, string, error) {
	urlStr := baseDomain + "/v4/ecs/list-instances"
	for _, id := range regid {
		params := map[string]interface{}{
			"regionID": id,
		}
		res, err := tianYApi(k.provider, urlStr, "", http.MethodPost, params, nil, "application/json;charset=UTF-8")
		if err != nil {
			return "", "", err
		}
		var response NodeResponse
		err = json.Unmarshal([]byte(res), &response)
		if err != nil {
			return "", "", err
		}
		for _, i := range response.ReturnObj.RegionList {
			if i.FloatingIP == k.ip {
				return i.InstanceID, id, nil
			}
		}
	}
	return "", "", fmt.Errorf("在支持的API接口中未找到%s", k.ip)
}

func getKpsLocation(kpsCode string) (kps KpsData) {
	var loca string
	sqlStr := "select ip,location,provider from kps where code = ?"
	Db.QueryRow(sqlStr, kpsCode).Scan(&kps.ip, &loca, &kps.provider)
	kps.location = strings.Split(loca, "市")[0]
	return
}

func (k KpsData) reNew(nodeID, regionID string) error {
	urlStr := baseDomain + "/v4/ecs/resubscribe-instance"
	now := time.Now().Unix()
	strNow := strconv.FormatInt(now, 10)
	params := map[string]interface{}{
		"cycleType":   "MONTH",
		"instanceID":  nodeID,
		"regionID":    regionID,
		"clientToken": strNow + "-" + "kpsRenew",
		"cycleCount":  1,
	}
	_, err := tianYApi(k.provider, urlStr, "", http.MethodPost, params, nil, "application/json;charset=UTF-8")
	if err != nil {
		return err
	}
	k.updataKpsExpTime()
	return nil
}

func (k KpsData) updataKpsExpTime() {
	sqlStr := "update kps set expire_time = DATE_ADD(expire_time, INTERVAL 1 MONTH) where ip = ?"
	Db.Exec(sqlStr, k.ip)
}

func RenewHandle(kpsCode string) (string, error) {
	kpsNode := getKpsLocation(kpsCode)
	regList, err := kpsNode.getNodeID()
	if err != nil {
		logError.Printf("get node id err %v", err)
		return "", err
	}
	if len(regList) > 0 {
		kpsID, regID, err := kpsNode.getNode(regList)
		if err != nil {
			return "", err
		} else {
			err := kpsNode.reNew(kpsID, regID)
			if err != nil {
				return "", err
			} else {
				return fmt.Sprintf("%s|%s|%s|%s 续费成功.", kpsCode, kpsNode.ip, kpsNode.location, kpsNode.provider), nil
			}
		}
	} else {
		return "", errors.New("没有找到对应地区或该地区不支持暂时不支持API")
	}
}
