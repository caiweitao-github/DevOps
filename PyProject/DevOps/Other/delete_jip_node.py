# -*- encoding: utf-8 -*-

import requests
import dbutil

kdljip = dbutil.get_db_kdljip()


def send_message(message):
    data = {
        "msg_type": "post",
        "content": {
            "post": {
                "zh_cn": {
                    "title": "[删除JIP节点通知]",
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
    url_token = ""
    requests.post(url_token, json=data, headers=headers)

def delete_data():
    sql = "delete from jip_node where status = 3 and update_time < DATE_SUB(now(),INTERVAL 15 DAY)"
    row = kdljip.execute(sql)
    return row.rowcount


if __name__ == '__main__':
    delete_num = delete_data()
    mess = "删除15天内未上线的边缘节点数据 -> %s" %(delete_num)
    send_message(mess)