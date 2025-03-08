#!/bin/bash
# 检测最近一分钟的webhp的django.log日志，对ERROR的日志告警到飞书Django日志告警
# */1 * * * * /bin/bash /home/httpproxy/bin/django_check.sh

log_file="/data/kdl/log/webhp/django.log"
tmp_file="/tmp/tmp.txt"
report_file="/tmp/feishu_report.log"

get_ERROR_info() {
    TIME=$(date -d "1 minute ago" +"[%Y-%m-%d %T]")
    # 查询最近一分钟的日志，以[年月日开头的行，如果有日志，就打印第一个匹配到的行号
    # num_start=$(awk -v start_time="$TIME" '$0 > start_time {print NR; exit}' "$log_file")
    num_start=$(awk '/^\[[[0-9]{4}-[0-9]{2}-[0-9]{2}/ && NR > 1 && $0 > start_time {print NR; exit}' start_time="$TIME" "$log_file")
    if [ -n "$num_start" ]; then
            awk 'NR >= start_line {print NR, $0}' start_line="$num_start" "$log_file" | grep "ERROR" > "$tmp_file"
    fi
}

ErrorInfo_lookup() {
    if [ -s "$tmp_file" ]; then
        while IFS= read -r line; do
            red_num=$(echo "$line" | cut -d " " -f2,3)
            white_show=$(echo "$line" | cut -d " " -f3-)
            printf "\033[31m%s\033[0m" "$red_num"
            echo "$white_show"
        done < "$tmp_file"

        num1=1
        while [ "$num1" -ne 0 ]; do
            read -p "要查看的报错行数(输入0退出): " num1
            tail -n +$num_start "$log_file" | head -n "$num1"
        done
    fi
}

ErrorInfo_report() {
    if [ -s "$tmp_file" ]; then
        awk '{print $2, $3; sub(/^[^:]+: /, ""); print $8,$9,$10}' "$tmp_file" >> "$report_file"
    fi
}

get_ERROR_info

if [ -z "$1" ]; then
    ErrorInfo_report
else
    ErrorInfo_lookup
fi

feishu_report() {
    curl -X POST -H "Content-Type: application/json" -d '{"msg_type":"text", "content":{"text": "'"$1"'"}}' "飞书机器人地址" >/dev/null 2>&1
}

if [ -f "$report_file" ]; then
    report_text="main4\tDjango日志报错"
    while IFS= read -r line; do
        report_text="$report_text\n$line"
    done < "$report_file"
    feishu_report "$report_text"
    rm -rf "$tmp_file"
    rm -rf "$report_file"
fi