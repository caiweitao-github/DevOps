package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	METHOD_GET  = "GET"
	METHOD_POST = "POST"
)

type accountKey struct {
	AccessKey   string
	SecurityKey string
}

var (
	file    = false
	account = make(map[string]accountKey, 12)
)

func init() {
	account["天翼云1"] = accountKey{"", ""}
	account["天翼云2"] = accountKey{"", ""}
	account["天翼云3"] = accountKey{"", ""}
	account["天翼云4"] = accountKey{"", ""}
	account["天翼云5"] = accountKey{"", ""}
	account["天翼云6"] = accountKey{"", ""}
	account["天翼云7"] = accountKey{"", ""}
	account["天翼云8"] = accountKey{"", ""}
	account["天翼云9"] = accountKey{"", ""}
	account["天翼云10"] = accountKey{"", ""}
	account["天翼云11"] = accountKey{"", ""}
	account["天翼云12"] = accountKey{"", ""}
}

func hmacSHA256(secret interface{}, data string) []byte {
	var secretBytes []byte
	switch v := secret.(type) {
	case []byte:
		secretBytes = v
	case string:
		secretBytes = []byte(v)
	default:
		panic("Unsupported secret type")
	}

	dataBytes := []byte(data)
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write(dataBytes)
	return mac.Sum(nil)
}

func base64OfHmac(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func getRequestUUID() string {
	return uuid.New().String()
}

func getSortedStr(data map[string]interface{}) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var strList []string
	for _, k := range keys {
		strList = append(strList, fmt.Sprintf("%s=%s", k, data[k]))
	}

	return strings.Join(strList, "&")
}

func buildSign(provider string, queryParams map[string]interface{}, bodyParams map[string]interface{}, eopDate, requestUUID, method string) string {
	var bodyStr string
	if !file && len(bodyParams) > 0 {
		bodyStrBytes, _ := json.Marshal(bodyParams)
		bodyStr = string(bodyStrBytes)
	}
	bodyDigest := ""
	if method == METHOD_POST {
		bodyStrBytes, _ := json.Marshal(bodyParams)
		bodyDigest = fmt.Sprintf("%x", sha256.Sum256(bodyStrBytes))
		// if file {
		// 	bodyDigest = fmt.Sprintf("%x", sha256.Sum256([]byte(string(bodyParams["file"])))
		// } else {
		// 	bodyStrBytes, _ := json.Marshal(bodyParams)
		// 	bodyDigest = fmt.Sprintf("%x", sha256.Sum256(bodyStrBytes))
		// }
	} else {
		bodyDigest = fmt.Sprintf("%x", sha256.Sum256([]byte(bodyStr)))
	}
	ak := account[provider].AccessKey
	sk := account[provider].SecurityKey
	headerStr := fmt.Sprintf("ctyun-eop-request-id:%s\neop-date:%s\n", requestUUID, eopDate)
	queryStr := EncodeQueryStr(getSortedStr(queryParams))
	signatureStr := fmt.Sprintf("%s\n%s\n%s", headerStr, queryStr, bodyDigest)
	signDate := strings.Split(eopDate, "T")[0]
	kTime := hmacSHA256(sk, eopDate)
	kAk := hmacSHA256(kTime, ak)
	kDate := hmacSHA256(kAk, signDate)
	signatureBase64 := base64OfHmac(hmacSHA256(kDate, signatureStr))
	signHeader := fmt.Sprintf("%s Headers=ctyun-eop-request-id;eop-date Signature=%s", ak, signatureBase64)
	return signHeader
}

func GetSignHeaders(provider string, queryParams map[string]interface{}, bodyParams map[string]interface{}, method, contentType string) map[string]string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	eopDate := now.Format("20060102T150405Z")
	requestUUID := getRequestUUID()

	headers := map[string]string{
		"Content-type":         contentType,
		"ctyun-eop-request-id": requestUUID,
		"Eop-Authorization":    buildSign(provider, queryParams, bodyParams, eopDate, requestUUID, method),
		"Eop-date":             eopDate,
	}
	return headers
}

func EncodeQueryStr(query string) string {
	if query == "" {
		return ""
	}

	params := strings.Split(query, "&")
	sort.Strings(params)

	encodedQuery := ""
	for i, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) == 2 {
			encodedStr := url.QueryEscape(kv[1])
			if i == 0 {
				encodedQuery += kv[0] + "=" + encodedStr
			} else {
				encodedQuery += "&" + kv[0] + "=" + encodedStr
			}
		}
	}

	return encodedQuery
}
