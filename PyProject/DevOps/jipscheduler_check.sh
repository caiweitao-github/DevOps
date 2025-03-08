#!/bin/bash
# 检测jip_server在数据库中已经启用，但是没在scheduler注册，则飞书告警
# 脚本路径：main13：/httpproxy/bin/jipscheduler_check.sh
# 运行方式：crontab：*/5 * * * * bash /home/httpproxy/bin/jipscheduler_check.sh >> /dev/null 2>&1

# 使用环境变量
export MYSQL_PWD=jip_r@node2024

# 运行 MySQL 客户端并获取数据
DB_QUERY="SELECT code FROM jip_server WHERE status = 1"
DB_CODES=$(mysql -u kdljip_r -h10.0.7.11 -D kdljip -e "$DB_QUERY")

if [ $? -ne 0 ]; then
    log "Failed to query database."
    exit 1
fi

# 从 API 获取数据
API_RESPONSE=$(curl --request GET \
  --url http://jips.gizaworks.com/d/get_server_info \
  --header 'content-type: application/json' \
  --header 'token: G02*h0xc@e5A')

if [ $? -ne 0 ]; then
    log "Failed to fetch data from API."
    exit 1
fi

# 检查数据库中的 code 是否存在于 API 中
for code in $DB_CODES; do
  if ! echo "$API_RESPONSE" | grep "$code"; then
    # 获取当前时间
    dt=$(date '+%T')
    message="$code 在scheduler未注册，请检查"

    # 构造飞书告警消息
    FEISHU_URL="https://open.feishu.cn/open-apis/bot/v2/hook/6939f829-2081-4af8-9017-5e14f5ef4e04"
    FEISHU_MSG='{
      "msg_type": "post",
      "content": {
        "post": {
          "zh_cn": {
            "title": "【JIP告警】'$dt'",
            "content": [
              [
                {"tag": "text", "text": "'$message'"},
                {"tag": "at", "user_id": "57d87g78", "user_name": "王景高（jerrywang）"}
              ]
            ]
          }
        }
      }
    }'

    # 发送飞书告警
    curl -X POST -H "Content-Type: application/json" -d "$FEISHU_MSG" "$FEISHU_URL" >/dev/null 2>&1
  fi
done