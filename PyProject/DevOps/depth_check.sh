#!/bin/bash
# 执行计划：crontab 每小时执行一次 0 * * * * bash /home/httpproxy/bin/depth_check.sh
# 检测每个requests的depth报错数量是否大于1w，大于则告警到飞书

NSQ_STATS_URL="http://127.0.0.1:4151/stats"
THRESHOLD=10000
FEISHU_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/3f3de577-5a27-4bb1-864f-eac4792ea12b"

# 获取NSQ统计信息，并提取包含"requests"的行
stats=$(curl -s "$NSQ_STATS_URL" | grep -E "\[.*channel.*\]")

# 检查是否成功获取到统计信息
if [ -z "$stats" ]; then
    echo "Failed to fetch NSQ stats."
    exit 1
fi

# 提取每个请求的名称和depth值，并进行比较
while read -r line; do
    request=$(echo "$line" | awk '{print $1}' | tr -d '[]')
    depth=$(echo "$line" | awk '{print $4}')
    if [ "$depth" -gt "$THRESHOLD" ]; then
        # 获取当前时间
        Time=$(date +"%Y-%m-%d %H:%M:%S")
        # 发送告警到飞书
        message="$request depth exceeded threshold: $depth"
        curl -X POST -H "Content-Type: application/json" \
        -d "{\"msg_type\":\"post\",\"content\":{\"post\":{\"zh_cn\":{\"title\":\"【NSQ告警】 $Time\",\"content\":[[{\"tag\":\"text\",\"text\":\"$message\"},{\"tag\": \"at\",\"user_id\": \"57d87g78\",\"user_name\": \"王景高（jerrywang）\"}]]}}}}}" \
        "$FEISHU_WEBHOOK_URL" &> /dev/null
    fi
done <<< "$stats"