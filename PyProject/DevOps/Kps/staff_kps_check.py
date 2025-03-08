#!/usr/bin/env python
# -*- encoding: utf-8 -*-

import requests
import dbutil

db_db = dbutil.get_db_db()

def send_msg(user, name, code, ip):
    data = {
	    "msg_type": "post",
	    "content": {
		    "post": {
			    "zh_cn": {
				    "title": "[KPS异常分配通知]",
				    "content": [
					    [{
                            "tag": "text",
                            "text": "员工账号占用独享型kps时间过长: %s(%s) ---> %s(%s)" %(user, name, code, ip)
						}
					    ]
				    ]
			   }
		    }
	    }
    }
    headers = {"Content-Type": "application/json"}
    url_token = "https://open.feishu.cn/open-apis/bot/v2/hook/688dc2b3-4d21-4ccf-a49d-aeb881bff4f5"
    requests.post(url_token, json=data, headers=headers)


def get_data():
    sql = """select au.username,au.first_name,k.code,k.ip from proxy_order po, auth_user au,order_kps ok,kps k  
    where po.user_id = au.id and ok.user_id = au.id   and po.id=ok.order_id and ok.kps_id = k.id and au.is_staff =1 
    and k.status not in (2,4) and k.level =5 and po.end_time > now() and DATEDIFF(CURDATE(), ok.create_time) > 2"""
    res = db_db.execute(sql).fetchall()
    if res:
        for i in res:
            send_msg(i[0], i[1], i[2], i[3])

if __name__ == '__main__':
    get_data()