#!/bin/bash

logFile='/var/log/clickhouse-server/clickhouse-server.err.log'
dateTIme=`date +"%Y.%m.%d %H:"`
dateTIme=`date --date='1 minutes ago' +"%Y.%m.%d %H:%M:"`
feishuUrl=''
key='md5Key'
logLine=50000

sendMsg() {
    message="$1"
    url_token=$feishuUrl
    message_escaped=$(echo "$message" | sed 's/"/\\"/g')

    data=$(jq -n \
        --arg msg_type "post" \
        --arg title "[ClickHouse日志报错]" \
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


checkFileMd5() {
    text=`md5sum $logFile|awk '{print $1}'`
    res=`redis-cli setnx $key $text`
    if [[ $res -eq 1 ]];then
        return 1
    fi
    cacheVale=`redis-cli get $key`
    if [[ $text != $cacheVale ]];then
        redis-cli set $key $text
        return 1
    fi
    return 0
}

checkLogError() {
    txt=`tail -n$logLine $logFile | awk -v start="$(date -d "1 minutes ago" "+%Y.%m.%d %H:%M:")" -v end="$(date "+%Y.%m.%d %H:%M:")" '$1" "$2 >= start && $1" "$2 <=end && $0 ~ /<Error>/'|grep -Ev 'I/O error:|Syntax error:'`
    if [[ $txt != "" ]];then
        sendMsg "${txt//\"/\\\"}"
    fi
}

main() {
    checkFileMd5
    if [[ $? -ne 0 ]];then
        checkLogError
    fi
}

main