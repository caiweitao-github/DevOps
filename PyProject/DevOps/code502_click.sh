#!/bin/bash
#执行方式crontab：*/1 * * * *
# 检测最近一分钟内是否有status code 502的报错日志信息

# 配置部分
LOG_FILE="/root/log/proxy.log"
WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/98c518dd-8faa-4ce0-b188-c939e1c34f4a"

# 获取最后一分钟的日志并检查是否存在错误
errors=$(tail -n 300000 "$LOG_FILE" | grep "`date '+%Y.%m.%d %H:%M' --date='-1 minute'`" | grep "status code 502" | wc -l)

# 如果发现错误，发送到飞书
if [ $errors -gt 0 ]; then
    # 构造JSON数据
    Host_name=$(hostname)
    json_payload=$(jq -n \
                    --arg text "$Host_name最近1分钟内status code 502报错数量为：$errors，请及时处理" \
                    '{msg_type: "text", content: {text: $text}}')

    # 使用curl发送POST请求到飞书
    curl -X POST -H 'Content-type: application/json' --data "$json_payload" $WEBHOOK_URL >/dev/null 2>&1
fi