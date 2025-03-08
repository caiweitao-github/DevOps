#!/bin/bash
#执行方式crontab：*/1 * * * *
# 检测最近一分钟内是否有刷/api/getmyip接口的ip，并封禁

# 配置部分
LOG_FILE="/data/log/nginx/api_access.log"
log_rows=10000

# 获取需要封禁的IP
get_block_ip() {
    # 获取最近一分钟内的访问日志
    START_TIME=$(date -d '1 minute ago' +'[%FT%T+08:00]')
    END_TIME=$(date +'[%FT%T+08:00]')
    IP_ADDRESSES=$(tail -n "$log_rows" "$LOG_FILE" | awk -v start="$START_TIME" -v end="$END_TIME" '$4 >= start && $4 <= end' | grep "getmyip?b=22" | awk '{print $1}' | sort -u)
    echo "$IP_ADDRESSES"
}

# 封禁IP
block_ip() {
    local ip=$1
    if ! /sbin/iptables -L INPUT -v -n | grep -q "$ip"; then
        /sbin/iptables -I INPUT -s "$ip" -j DROP
        echo "Blocked IP: $ip" >> /root/log/getmyip_check.log 2>&1
    fi
}

ip_to_block=$(get_block_ip)
for ip in $ip_to_block; do
    block_ip "$ip"
done
