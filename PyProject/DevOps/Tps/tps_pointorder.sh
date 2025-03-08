#!/bin/bash
# db="mysql -Ddb -N -e"
db="mysql -udb_r -pdb_r -h10.0.3.17 -Ddb -N -e"
get_tps=/tmp/get_tps.txt

# 构造 JSON 数据
json_data='{
    "msg_type": "interactive",
    "card": {
        "config": {
            "wide_screen_mode": true
        },
        "header": {
            "template": "red",
            "title": {
                "content": "用户指定tps域名分配异常",
                "tag": "plain_text"
            }
        },
        "i18n_elements": {
            "zh_cn": ['

# 记录是否有异常的标志
has_abnormal=0

while read line 
do
    order=$(echo $line | awk '{print $1}')
    domain=$(echo $line | awk '{print $2}')
    id=$(echo $line | awk '{print $3}')
    tps_code=$(echo $line | awk '{print $4}')
    judgment=$($db "select count(*) from tps_domain where domain = '$domain' and tps_id = $id;")
    if [ $judgment -eq 0 ]; then
        now_local=$($db "SELECT t.code FROM tps t JOIN ( SELECT tps_id FROM tps_domain WHERE domain = '$domain') td ON t.id = td.tps_id;")
        json_data="${json_data}{\"tag\":\"div\",\"text\":{\"content\":\"订单: $order, 域名: $domain, 目前位置: $now_local, 请尽快切回:$tps_code\n\",\"tag\":\"lark_md\"}},"
        has_abnormal=1
    fi
done < $get_tps

# 如果有异常，最后添加@所有人
if [ $has_abnormal -eq 1 ]; then
    json_data="${json_data}{\"tag\":\"div\",\"text\":{\"content\":\"<at id='all'></at>\",\"tag\":\"lark_md\"}},"
fi

# 去除最后的逗号
json_data=${json_data%,}
json_data="${json_data}]}}}"

# 检查 JSON 数据是否为空
if [[ $json_data != *"\"zh_cn\": []"* ]]; then
    webhook_url="https://open.feishu.cn/open-apis/bot/v2/hook/a5ff1482-71e1-4a33-8bfb-94467b9b8c53"
    # 使用 curl 发送到飞书机器人，并获取响应状态码和响应内容
    response=$(curl -s -w "\n%{http_code}" -H "Content-Type: application/json" -X POST -d "$json_data" "$webhook_url")
fi

