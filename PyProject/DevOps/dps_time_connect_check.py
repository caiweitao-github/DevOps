# -*- coding: utf-8 -*-

"""

周期性运行，查找最近一周内，time_connect  > x 发生 y 次 的天数，大于n天。
找到这样的机器后，告警出来，让运维进行修改。

运行方式：cron.d运行
启动命令：python -u /home/httpproxy/devops/dps_time_connect_check.py >/dev/null 2>&1 &

"""
import sys
import datetime
import random
import time
import datetime
import json
import requests

import loggerutil
import dbutil
import ckutil

logger = loggerutil.get_logger("devops", "devops/dps_time_connect_check.log")
db = dbutil.get_db_db()
ck = ckutil.get_ck_nodeops()

TIME_CONNECT_THRESHOLD = 10000
TIME_CONNECT_COUNT = 1000
DAY_THRESHOLD = 5


def get_dps_dpsgroup_dict():
    """得到所有dps节点对应的分组情况"""
    dps_dpsgroup_dict = {}
    sql = """select d.code,dg.code as group_code from dps d,dps_group_relation dgr ,dpsgroup dg where d.id=dgr.dps_id and dg.id=dgr.group_id;"""
    cursor = db.execute(sql)
    rows = cursor.fetchall()
    for row in rows:
        if row[0] not in dps_dpsgroup_dict.keys():
            dps_dpsgroup_dict[row[0]] = []
            dps_dpsgroup_dict[row[0]].append(row[1])
        else:
            dps_dpsgroup_dict[row[0]].append(row[1])
    return dps_dpsgroup_dict

def get_negative_data_dict():
    data = {}
    edge_start = (datetime.datetime.now() - datetime.timedelta(days=8)).strftime("%Y-%m-%d 00:00:00")
    edge_end = (datetime.datetime.now() - datetime.timedelta(days=1)).strftime("%Y-%m-%d 23:59:59")
    sql = """select code, toDate(stat_time) as st, count(*) as ct from node_connect_history 
    where stat_time > '%s' and stat_time < '%s' and connect>%d group by code, st having ct>%d 
    """ % (edge_start, edge_end, TIME_CONNECT_THRESHOLD, TIME_CONNECT_COUNT)
    rows = ck.execute(sql)
    if rows:
        codes = set([item[0] for item in rows ])
        data = {code:0 for code in codes}
        for row in rows:
            code = row[0]
            data[code] = data[code] + 1

    return data

def process_data(data_dict):
    todo = []
    dps_dspgroup_dict = get_dps_dpsgroup_dict()
    for code, count in data_dict.items():
        if count >= DAY_THRESHOLD:
            group = ','.join(dps_dspgroup_dict.get(code, ''))
            todo.append("%s 分组:%s" % (code, group))
    return todo


# 飞书告警
def feishu_notify(text):
    url = "https://open.feishu.cn/open-apis/bot/v2/hook/b91cfff5-2881-4999-b1af-1c3edc671635"
    get_time = time.strftime("%H:%M", time.localtime())

    payload_message = {
        "msg_type": "post",
        "content": {
            "post": {
                "zh_cn": {
                    "title": "【机器更换建议】 %s" % (get_time),
                    "content": [
                        [
                            {
                                "tag": "text",
                                "text": "以下机器持续%d天连接耗时超过%d秒次数大于%d次，建议更换：\n" % (DAY_THRESHOLD, TIME_CONNECT_THRESHOLD/1000, TIME_CONNECT_COUNT)
                            },
                            {
                                "tag": "text",
                                "text": text
                            },
                            {
                                "tag": "a",
                                "text": "\n数据平台",
                                "href": "http://123.206.57.47:8701/stat/dps_node_connected_timeuse"
                            }
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
    logger.info(response.text)

def main():
    negative_data_dict = get_negative_data_dict()
    textlist = process_data(negative_data_dict)
    feishu_notify('\n'.join(textlist))

if __name__ == '__main__':
    main()