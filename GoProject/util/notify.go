package util

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var apiUrl = ""

var appID = ""

var appKey = ""

var phoneNumber = [...]string{""}

func notifyVoice(phone string) {
	payload := struct {
		AppID   string `json:"appid"`
		Mobile  string `json:"to"`
		Message string `json:"content"`
		AppKey  string `json:"signature"`
	}{
		AppID:   appID,
		Mobile:  phone,
		Message: "系统重要通知，工程师你好，系统发生严重告警，中转服务告警，请立即处理。",
		AppKey:  appKey,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	contentType := "application/json"
	resp, err := http.Post(apiUrl, contentType, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func Notify() {
	for _, i := range phoneNumber {
		notifyVoice(i)
	}
}
