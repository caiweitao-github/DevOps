# !/usr/bin/python
# -*- coding: UTF-8 -*-
# author_='Joinishu';
# date: 2021/2/4 15:49
import MySQLdb
import json
import requests
import time
import dbutil
import loggerutil

logger = loggerutil.get_logger("devops", "devops/unexcepted_change_notify.log")
get_time=time.strftime("%H:%M", time.localtime())
db_nodeops = dbutil.get_db_nodeops()
textlist = []

def get_unexped_changeip():
    sql = "select dps_code,changeip_period,count(dps_code) from dps_changeip_history where is_expected=0 and is_valid=1 and change_time >=(NOW() - interval 24 hour) group by dps_code,changeip_period having count(dps_code) > 1 order by count(dps_code) desc;"
    cursor = db_nodeops.execute(sql)
    results = cursor.fetchall()
    for row in results:
        dps_code = row[0]
        period = row[1]
        count = row[2]
        is_notify(dps_code, count, period)


def notify_content(dps_code, count, period):
    text = "%s:%ss, %s次" % (dps_code, str(period), str(count))
    textlist.append(text)
    logger.info(text)


def is_notify(dps_code, count, period):
    try:
        if period == 14400 and count > 2:
            notify_content(dps_code, count, period)
        elif period == 10800 and count > 3:
            notify_content(dps_code, count, period)
        elif period == 3660 and count > 5:
            notify_content(dps_code, count, period)
        elif period == 2760 and count > 6:
            notify_content(dps_code, count, period)
        elif period == 1860 and count > 9:
            notify_content(dps_code, count, period)
        elif period == 630 and count > 27:
            notify_content(dps_code, count, period)
        elif period == 320 and count > 54:
            notify_content(dps_code, count, period)
        elif period == 200 and count > 86:
            notify_content(dps_code, count, period)
        else:
            pass
        # print(*textlist)
    except Exception as e:
        logger.exception(str(e))


# 飞书告警
def feishu_notify(textlist):
    url = "https://open.feishu.cn/open-apis/bot/v2/hook/b91cfff5-2881-4999-b1af-1c3edc671635"

    payload_message = {
        "msg_type": "post",
        "content": {
            "post": {
                "zh_cn": {
                    "title": "【异常换IP告警】 %s" % (get_time),
                    "content": [
                        [
                            {
                                "tag": "text",
                                "text": "%s" % (textlist)
                            },
                            {
                                "tag": "a",
                                "text": "\n数据平台",
                                "href": "http://123.206.57.47:8701/stat/unexpected_changeip_dps"
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


if __name__ == '__main__':
    get_unexped_changeip()
    feishu_notify('\n'.join(textlist))

