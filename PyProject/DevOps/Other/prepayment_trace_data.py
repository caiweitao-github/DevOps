# -*- encoding: utf-8 -*-

import json
from datetime import date, timedelta
import requests
import dbutil

db_db = dbutil.get_db_db()


def sendCardmessage(content, id):
  url = ""
  headers={'Content-Type': 'application/json'}
  message_body={
          "msg_type": "interactive",
          "card": {
            "config": {
              "wide_screen_mode": True
            },
            "elements": [
              {
                "tag": "div",
                "text": {
                  "content": content,
                  "tag": "lark_md"
                }
              },
              {
                "tag": "action",
                "actions": [
                      {
                        "tag": "button",
                        "text": {
                        "tag": "plain_text",
                        "content": "订单流程详情"
                        },
                        "type": "primary",
                        "multi_url": {
                        "url": "https://url/%s/" %(id),
                        }
                      }
                  ]
              }
            ],
            "header": {
              "template": "blue", # 消息卡片主题颜色，可选值：red、orange、yellow、green、cyan、blue、purple、pink
              "title": {
                "content": "预付款需跟",
                "tag": "plain_text"
              }
            }
          }}
  requests.post(url, headers=headers, data=json.dumps(message_body))

def get_customer_trace_data(start_time, end_time):
  sql = """select op.id,company_name,contact_name,next_pay_time  from order_process op 
  where status in (1,2) and next_pay_time between  DATE_SUB(CURDATE(), INTERVAL 2 DAY) 
  and  DATE_ADD(CURDATE(), INTERVAL 2 DAY)""".format(start_time, end_time)
  r = db_db.execute(sql).fetchall()
  if not r:
      return
  for d in r:
    sendCardmessage("**公司名称: **%s\n**下次付款时间: **%s" %(d[1], d[3]), d[0])


if __name__ == '__main__':
  week = date.today().weekday()
  if week == 0:
      start_day = (date.today() + timedelta(-3)).strftime("%Y-%m-%d %H:%M:%S")
      end_day = (date.today() + timedelta(-1)).strftime("%Y-%m-%d %H:%M:%S")
  elif week in (1, 2, 3, 4):
      start_day = (date.today() + timedelta(-1)).strftime("%Y-%m-%d %H:%M:%S")
      end_day = (date.today()).strftime("%Y-%m-%d %H:%M:%S")
  get_customer_trace_data(start_day, end_day)
