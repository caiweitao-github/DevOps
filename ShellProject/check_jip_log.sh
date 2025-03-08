#!/bin/bash
pid_file="/tmp/jip_pid.txt"
data_time=$(date -d "1 minutes ago" +"%H:%M")


if [[ ! -e $pid_file ]];then
    echo $(ps -ef|grep  '[/]root/jips/jip'|awk '{print $2}') > $pid_file
fi

if [[ $(ps -ef|grep  '[/]root/jips/jip'|awk '{print $2}') -ne $(cat $pid_file) ]];then
    pid=$(cat $pid_file)
    echo $(ps -ef|grep  '[/]root/jips/jip'|awk '{print $2}') > $pid_file
else
    pid=$(cat $pid_file)
fi


send_mess() {
    message="$1"
    url_token=""
    message_escaped=$(echo "$message" | sed 's/"/\\"/g')

    data=$(jq -n \
        --arg msg_type "post" \
        --arg title "[jip日志报错]" \
        --arg content "$message_escaped" \
        '{
            "msg_type": $msg_type,
            "content": {
                "post": {
                    "zh_cn": {
                        "title": $title,
                        "content": [
                            [
                                {
                                    "tag": "text",
                                    "text": $content
                                }
                            ]
                        ]
                    }
                }
            }
        }')

    headers="Content-Type: application/json"
    curl -X POST -H "$headers" -d "$data" "$url_token"
}


err_content=$(journalctl -ru jip_server|grep $data_time|grep 'jip'|grep $pid|grep -E "panic|recovered|fatal")

if [[ $err_content != "" ]];then
    send_mess "${err_content//\"/\\\"}"
fi