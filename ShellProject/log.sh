#!/bin/bash
LogFile=${1:-"/data/kdl/log/kdllogstash/logstash.log"}
now=$(date +"%FT%H")

send_msg() {
    message="$1"
    url_token=""
    data=$(cat <<EOF
{
    "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "[kdllogstash日志报错通知]",
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

err_num=$(cat $LogFile |grep "$now:"|grep 'i/o timeout; invalid transaction'|wc -l)



if [[ $err_num -gt 0 ]];then
    ps -ef|grep 'kdllogstash'|awk '{print $2}'|xargs kill -9
    send_msg "检测到kdllogstash日志报错，报错数量$err_num，已自动重启进程."
fi