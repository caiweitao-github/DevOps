#!/usr/bin/env python
# -*- encoding: utf-8 -*-
import time

import requests  
import dbutil


db = dbutil.get_db_db()

def send_msg(context):
    date_time = time.strftime("%H:%M:%S", time.localtime())
    data = {
            "msg_type": "post",
            "content": {
                "post": {
                    "zh_cn": {
                        "title": "[海外代理静态型续费通知]  %s" %(date_time),
                        "content": [
                            [
                                {
                                    "tag": "text",
                                    "text": context
                                },
                            ]
                        ]
                    }
                }
            }
        }
    headers = {"Content-Type": "application/json"}
    url_token = "https://open.feishu.cn/open-apis/bot/v2/hook/688dc2b3-4d21-4ccf-a49d-aeb881bff4f5"
    requests.post(url_token, json=data, headers=headers)

def get_node():
    sql = """select id from sfps where status in (1,3) and expire_time between DATE_FORMAT(NOW(), '%Y-%m-%d 00:00:00') and 
    DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 1 DAY), '%Y-%m-%d 23:59:59')"""
    return db.execute(sql).fetchall()


def get_order():
    node_id = get_node()
    mess = []
    for i in node_id:
        sql = """select sfps.code,proxy_order.orderid,proxy_order.end_time,sfps.expire_time,
        TIMESTAMPDIFF(HOUR, sfps.expire_time, proxy_order.end_time) as diff from sfps,order_sfps,
        proxy_order where proxy_order.level in ('21', '22') and proxy_order.status = 'TRADE_SUCCESS' and proxy_order.id = order_sfps.order_id and 
        order_sfps.sfps_id = sfps.id and proxy_order.end_time > NOW() and TIMESTAMPDIFF(HOUR, sfps.expire_time, proxy_order.end_time) >= 1 and sfps.id = '%s'""" %(i[0])
        rows = db.execute(sql).fetchall()
        for i in rows:
            mess.append("%s, 订单: %s, 订单到期时间: %s, 节点到期时间: %s, 差值: %s" %(i[0], i[1], i[2].strftime("%Y-%m-%d %H:%M") , i[3].strftime("%Y-%m-%d %H:%M") , i[4]))
    return mess

if __name__ == '__main__':
    mess = get_order()
    data = '\n'.join(mess)
    if len(data) > 0:
        send_msg(data)
