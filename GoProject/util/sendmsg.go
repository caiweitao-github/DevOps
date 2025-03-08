package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func SendMessUrl(msgs []string, title string) {
	apiurl := ""
	contentType := "application/json"
	var content []interface{}
	for _, m := range msgs {
		st := strings.Split(m, " ")
		content = append(content, []interface{}{
			map[string]interface{}{
				"tag":  "text",
				"text": fmt.Sprintf("%s %s %s", st[0], st[1], st[2]),
			},
			map[string]interface{}{
				"tag":  "a",
				"text": st[3],
				"href": fmt.Sprintf("", st[3]),
			},
			map[string]interface{}{
				"tag":  "text",
				"text": fmt.Sprintf(" %s %s", st[4], st[5]),
			},
		})
	}
	sendData := map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
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

func FeiShuNotify(url, title string, msgs []string) {
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
					"title":   title,
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
	result, err := http.Post(url, contentType, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("post failed, err:%v\n", err)
		return
	}
	defer result.Body.Close()
}

func StaffFeiShuNotify(url, title string, msgs []string) {
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
	content = append(content, []interface{}{
		// map[string]interface{}{
		// 	"tag":     "at",
		// 	"user_id": "986671fb",
		// },
		// map[string]interface{}{
		// 	"tag":     "at",
		// 	"user_id": "2966d9g7",
		// },
		map[string]interface{}{
			"tag":     "at",
			"user_id": "5c852geg",
		},
	})
	sendData := map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   title,
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
	result, err := http.Post(url, contentType, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("post failed, err:%v\n", err)
		return
	}
	defer result.Body.Close()
}
