#!/usr/bin/env python
# -*- encoding: utf-8 -*-

import requests
import dbutil

db_db = dbutil.get_db_db()

def send_msg(code, ip, provider, expire_time, orderid, end_time):
    data = {
	    "msg_type": "post",
	    "content": {
		    "post": {
			    "zh_cn": {
				    "title": "[KPS异常分配通知]",
				    "content": [
					    [{
                            "tag": "text",
                            "text": "机器信息: %s(%s | %s | %s) 订单: %s(%s)" %(code, ip, provider, expire_time, orderid, end_time)
						}
					    ]
				    ]
			   }
		    }
	    }
    }
    headers = {"Content-Type": "application/json"}
    url_token = ""
    requests.post(url_token, json=data, headers=headers)

def get_kps():
    sql = """select kps.code,kps.ip,kps.provider,kps.expire_time,proxy_order.orderid,proxy_order.end_time from 
    kps,proxy_order,order_kps where kps.status = 4 and kps.expire_time between DATE_FORMAT(DATE_ADD(now(), INTERVAL -7 DAY),'%Y-%m-%d 00:00:00') 
    and DATE_FORMAT(now(),'%Y-%m-%d 00:00:00') and order_kps.order_id=proxy_order.id and order_kps.kps_id=kps.id and proxy_order.end_time > now() and kps.level != 5"""
    res = db_db.execute(sql).fetchall()
    if res:
        for i in res:
            send_msg(i[0], i[1], i[2], i[3], i[4], i[5])


if __name__ == '__main__':
    get_kps()
