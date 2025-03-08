# -*- coding: utf-8 -*-
"""
检测第三方节点实时可用ip，小于10%则说明有问题飞书通知到运维告警群，一小时内只告警一次
名称：tp_ipnumber.py
运行机器：kdlmain10
运行方式：cron
周期：1分钟运行一次
"""
import time
import requests
import MySQLdb
import redis

# 连接到Redis
redis_client = redis.StrictRedis(host='localhost', port=6379, db=0)

# 连接到MySQL数据库
conn = MySQLdb.connect(
    host='10.0.3.17',
    user='kdlnode_r',
    passwd='kdlnode_r@2023',
    db='kdlnode'
)

def send_msg_if_needed(message):
    last_alert_time = float(redis_client.get('last_alert_time') or 0)
    current_time = time.time()

    # 计算当前时间戳减去redis上一次存储的时间戳大于1小时则发送飞书告警
    if current_time - last_alert_time > 3600:
        send_msg(message)
        redis_client.set('last_alert_time', str(current_time))

def send_msg(message):
    date_time = time.strftime("%H:%M:%S", time.localtime())
    data = {
        "msg_type": "post",
        "content": {
            "post": {
                "zh_cn": {
                    "title": "[第三方节点资源减少通知]  %s" % (date_time),
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
    url_token = "https://open.feishu.cn/open-apis/bot/v2/hook/3f3de577-5a27-4bb1-864f-eac4792ea12b"
    requests.post(url_token, json=data, headers=headers)

def get_counts():
    try:
        cursor = conn.cursor()

        cursor.execute("SELECT COUNT(*) FROM node WHERE source NOT IN ('dps', 'kps', 'jde')")
        total_count = cursor.fetchone()[0]

        cursor.execute("SELECT COUNT(*) FROM node WHERE source NOT IN ('dps', 'kps', 'jde') AND status = 1")
        normal_count = cursor.fetchone()[0]

        return total_count, normal_count

    except Exception as e:
        print("发生异常：", e)
        return None, None

    finally:
        if 'cursor' in locals():
            cursor.close()

try:
    total_count, normal_count = get_counts()

    if total_count is not None and normal_count is not None:
        abnormal_count = total_count - normal_count
        percentage = 100 - (float(normal_count) / total_count) * 100 if total_count != 0 else 0

        if percentage > 10:
            message = "警告：可用节点数量！总数：%d，异常数：%d，占比：%d" % (total_count, abnormal_count, percentage)
            send_msg_if_needed(message)

except Exception as e:
    print("发生异常：", e)

finally:
    # 关闭数据库连接
    if 'conn' in locals():
        conn.close()