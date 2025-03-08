#!/bin/bash
# 常驻进程：
# Nginx访问日志文件路径
LOG_FILE="/data/log/nginx/webstat_access.log"

# 正常访问的IP列表
ALLOWED_IPS=("171.113" "120.85" "222.209" "171.214" "219.157" "119.103" "117.154" "221.234" "171.82" "171.83" "27.16" "171.43" "222.211" "61.52" "111.181")
ALLOWED_REGEX=$(IFS="|"; echo "${ALLOWED_IPS[*]}")

# 收集非正常访问IP
BANNED_IPS=()

# 检测非经常登录IP并触发告警
while true; do
    # 获取最近一分钟内的访问日志
    START_TIME=$(date -d '1 minute ago' +'[%FT%T+08:00]')
    END_TIME=$(date +'[%FT%T+08:00]')
    LOG_ENTRIES=$(awk -v start="$START_TIME" -v end="$END_TIME" '$4 >= start && $4 <= end {print $1}' "$LOG_FILE" | grep 'adm' | sort -u)

    # 检查每个IP地址是否为正常访问IP
    for ip in $LOG_ENTRIES; do
        # 提取IP地址的前两段
        ip_prefix=$(echo "$ip" | cut -d'.' -f1-2)
        # 判断ip前两段是否不相同
        if [[ ! "$ip_prefix" =~ $ALLOWED_REGEX ]]; then
            BANNED_IPS+=("$ip")
        fi
    done

    # 如果不相同，触发告警
    if [ ${#BANNED_IPS[@]} -gt 0 ]; then
        dt=$(date +%T)
        # 遍历 BANNED_IPS 列表中的每个 IP
        for ip in "${BANNED_IPS[@]}"; do
            region=$(curl -s "http://cip.cc/$ip" | awk -F " " 'NR == 7 {print $3}')
            ban_detail="非正常访问IP：$ip"
            # 判断 IP 所在地区是否包含 "武汉"，不包含则发送通知
            if [[ $region != *"武汉"* ]]; then
                curl -X POST -H "Content-Type: application/json" \
                -d '{"msg_type":"post","content":{"post":{"zh_cn":{"title":"【stat访问ip】 '$dt'","content":[[{"tag":"text","text":"'$ban_detail' '$region'"}]]}}}}' \
                https://open.feishu.cn/open-apis/bot/v2/hook/3f3de577-5a27-4bb1-864f-eac4792ea12b >/dev/null 2>&1
            fi
        done

        # 清空已触发告警的IP列表
        BANNED_IPS=()
    fi

    sleep 60  # 等待60秒
done