#!/usr/bin/env python
# -*- encoding: utf-8 -*-
"""
中转服务带宽检测脚本，持续超带宽会电话通知运维人员
超带宽处理方式：
    1. 限制用户实控带宽
    2. 调整IP池 [优先 dps/jde ]
    3. 暂停订单
"""

from datetime import datetime
import requests
import staffnotify
import dbutil

from redisutil import redisdb

node_key = "_bandwidth_data"
notify_key = "is_notify"

db_db = dbutil.get_db_db()
kdlnode_db = dbutil.get_db_kdlnode()


def send_message(message):
    data = {
        "msg_type": "post",
        "content": {
            "post": {
                "zh_cn": {
                    "title": "[中转服务带宽监控]",
                    "content": [
                        [{
                            "tag": "text",
                            "text": message
                        }]
                    ]
                }
            }
        }
    }
    headers = {"Content-Type": "application/json"}
    url_token = "https://open.feishu.cn/open-apis/bot/v2/hook/e9178eab-c86a-485b-a761-5223a3fb2f24"
    requests.post(url_token, json=data, headers=headers)


def check_node_bandwidth(category):
    params = {
        'jde': ('jde_server', db_db),
        'yle': ('yle_server', kdlnode_db),
        'tdps': ('transfer_server', kdlnode_db)
    }
    key = category + node_key
    if category in params:
        tb, db = params[category]
    node_count = """select count(*) from %s where status = 1""" %(tb)
    sql = """select code,format(realtime_bandwidth/1024/128,0),bandwidth from %s where status  = 1""" %(tb)
    node_data = db.execute(sql).fetchall()
    notify_count = int(db.execute(node_count).fetchone()[0]) / 2
    count = 0
    for code, realtime_bandwidth, bandwidth in node_data:
        if int(realtime_bandwidth) > int(bandwidth * 0.9):
            redisdb.hincrby(key, code, amount=1)
        else:
            redisdb.hset(key, code, 0)
        cache_num = redisdb.hget(key, code)
        if int(cache_num) >= 3:
            count += 1
            send_message("%s 实时带宽: %s, 机器总带宽: %s" %(code, realtime_bandwidth, bandwidth))
    return count > int(notify_count)

def is_working_time():
    working_time = [i for i in range(0,9)] + [23]
    now = datetime.now().hour
    return now in working_time


if __name__ == '__main__':
    is_notify = is_working_time()
    transfer_server_category = ['tdps', 'yle', 'jde']
    for i in transfer_server_category:
        r = check_node_bandwidth(i)
        if r:
            if is_notify and redisdb.set(notify_key, '1', ex=60 * 60 * 4, nx=True):
                staffnotify.notify_voice("多台中转服务带宽过高")