#!/bin/bash
description=$1
publisher=${3:-发布人}
version=${2:-None}
code=$(grep -E "kps[0-9]+" /root/ansible_project/kps_inventory|awk '{print $1}'|awk -F ':' '{print $1}')
mapfile -t kps_code <<< $code
num=$(grep -E "kps[0-9]+" /root/ansible_project/kps_inventory|wc -l)
text="发布人: $publisher\n版本号: $version\n发布数量: $num\n发布描述: $description\n发布内容:${kps_code[@]}"


send_msg() {
    message="$1"
    url_token=""
    data=$(cat <<EOF
{
    "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "[kps发布通知]",
                "content": [
                    [{
                        "tag": "text",
                        "text": "$message"
                    }]
                ]
            }
        }
    }
}
EOF
)

    headers="Content-Type: application/json"
    curl -X POST -H "$headers" -d "$data" "$url_token"
}

send_msg "$text"