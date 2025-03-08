#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
查询最近1天切换的隧道域名对应的tid是否还请求之前分配的tps_code，不一致则飞书告警
运行方式：cron
运行机器: main13
20 9 * * 1-6 python -u /home/httpproxy/bin/tps_host_check.py 2>&1
"""
import time
import requests
import json

import ckutil
import dbutil
from datetime import date

db_db = dbutil.get_db_db()
ck_tpsstat = ckutil.get_ck_tpsstat()

# 获取今天的日期
today = date.today()

# 获取告警信息
alerting_text = []

# 飞书告警
def feishu_notify(alerting_list):
    url = 'https://open.feishu.cn/open-apis/bot/v2/hook/a5ff1482-71e1-4a33-8bfb-94467b9b8c53'
    get_time = time.strftime("%H:%M", time.localtime())
    payload_message = {
        "msg_type": "post",
        "content": {
            "post": {
                "zh_cn": {
                    "title": "【请求错误tps_code告警】 %s" % (get_time),
                    "content": [
                        [
                            {
                                "tag": "text",
                                "text": "%s" % (alerting_list)
                            },
                        ]
                    ]
                }
            }
        }
    }
    headers = {
        'Content-Type': 'application/json'
    }
    response = requests.request("POST", url, headers=headers, data=json.dumps(payload_message))

# 执行MySQL查询
mysql_query = (
    "SELECT tps.code, tunnel.tid "
    "FROM tps_domain_change t "
    "JOIN tunnel ON tunnel.host = t.src_domain "
    "JOIN tps ON tps.id = t.dest_tps_id "
    "WHERE t.update_time >= DATE_SUB(CURDATE(), INTERVAL 1 DAY) "
    "AND tunnel.status = 1;"
)
cursor = db_db.execute(mysql_query)
mysql_results = cursor.fetchall()

# 执行ClickHouse查询
result = []
for row in mysql_results:
    tid = row[1]  # 使用整数索引
    clickhouse_query = (
        "SELECT tps_code, request_time "
        "FROM tunnel_bandwidth_history "
        "WHERE tid = '{}' "
        "AND request_time >= '{}' "
        "ORDER BY create_time DESC "
        "LIMIT 1;".format(tid, today)
    )
    clickhouse_results = ck_tpsstat.execute(clickhouse_query)
    if clickhouse_results and len(clickhouse_results) > 0:  # 检查结果是否不为空
        clickhouse_tps_code = clickhouse_results[0][0]
        request_time = clickhouse_results[0][1]  # 获取request_time
        if row[0] != clickhouse_tps_code:  # 使用整数索引
            alerting_text.append("tid: {}, MySQL: {}, CK: {}, Time: {}".format(row[1], row[0], clickhouse_tps_code, request_time))

# 打印结果
if alerting_text:
    #print(alerting_text)
    feishu_notify('\n'.join(alerting_text))

# 关闭数据库连接
cursor.close()
db_db.close()