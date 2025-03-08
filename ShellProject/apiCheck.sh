#!/bin/bash
declare -A data
declare -A text

declare -A Other=(
    [feishuUrl]=''
)

logDir=/data/log/nginx/
logFile=(api_access.log fetch_access.log)
baseTime=$(date --date='2 hour ago' +"%Y-%m-%dT%H")
beforeHour=$(date --date='1 hour ago' +"%Y-%m-%dT%H")
regexStr="[0-5][0-9]:[0-5][0-9]"
tmpFile=/root/bin/tmp.log
send_msg() {
    name=$(hostname)
    message="$1"
    url_token=${Other[feishuUrl]}
    data=$(cat <<EOF
{
    "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "[API请求次数过高通知]     $name",
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

main() {
for i in ${logFile[@]};do
    if [[ -f $logDir$i ]];then
        awk  "/$beforeHour:$regexStr/" $logDir$i|awk -F '/api/' '{print $2}'|awk -F '?' '{print $1}'|awk -F '[ /]' '{print $1}'|sort -rn |uniq -c|sort -rn|sed 's/^[[:space:]]*//' > $tmpFile
        while read line;do
            num=$(echo $line|awk '{print $1}')
            api=$(echo $line|awk '{print $2}')
            if [[ $api != "" ]];then
                data[$api]=$num
            fi
        done < $tmpFile
        awk  "/$baseTime:$regexStr/" $logDir$i|awk -F '/api/' '{print $2}'|awk -F '?' '{print $1}'|awk -F '[ /]' '{print $1}'|sort -rn |uniq -c|sort -rn|sed 's/^[[:space:]]*//' > $tmpFile
        while read line;do
            num=$(echo $line|awk '{print $1}')
            api=$(echo $line|awk '{print $2}')
            if [[ $api != "" ]];then
                diffData=${data[$api]}
                if [[ $diffData != "" ]];then
                    res=$(awk -v n1="$diffData" -v n2="$num" 'BEGIN{printf ("%.0f\n",((n1-n2)/n2)*100)}')
                    if [ $res -gt 50 -a $diffData -ge 500 ];then
                        text[$api]="$api：$num\t→\t$diffData，增长率：$res%"
                    fi
                elif [[ $num -gt 1000 ]];then
                    text[$api]="$api：0\t→\t$num，增长率：$num%"
                fi
            fi
        done < $tmpFile
        report_text="以下API请求增长超过阈值"
        if [[ ${#text[@]} -ne 0 ]];then
            for txt in ${text[@]}
            do
                report_text="$report_text\n$txt"
            done
            send_msg $report_text
        fi
        unset text
    fi
done
}

main
