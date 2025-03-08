# -*- encoding: utf-8 -*-

import json
from datetime import date, timedelta
import requests

from dbutil import *

db_db = get_db_db()


wx_staff_data = {

}

ent_staff_data = {

}

tencent_qd_staff_data = {

}

tencent_qq_staff_data = {

}


def sendCardmessage(content, user_id):
  url = "https://open.feishu.cn/open-apis/bot/v2/hook/cbeadcc4-329a-4c5e-9323-b4ab60a5aef6"
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
                        "content": "跟进详情"
                        },
                        "type": "primary",
                        "multi_url": {
                        "url": "https://www.db.com/bkm1/customer/bigcustomertrace/%s/" %(user_id),
                        }
                      }
                  ]
              }
            ],
            "header": {
              "template": "green", # 消息卡片主题颜色，可选值：red、orange、yellow、green、cyan、blue、purple、pink
              "title": {
                "content": "今日需跟进数据",
                "tag": "plain_text"
              }
            }
          }}
  requests.post(url, headers=headers, data=json.dumps(message_body))

def get_customer_trace_data(start_time, end_time):
  sql = """select bt.id,bt.user_id,au.first_name,contact_channel,kefu_wx_account,
  contact_memo,trace_record from big_customer_trace bt,auth_user au, user_profile up where bt.staff_id=au.id 
  and bt.user_id =up.user_id and trace_time between '{}' and '{}' and trace_record like '%需要进一步跟进%'""".format(start_time, end_time)
  r = db_db.execute(sql).fetchall()
  if not r:
      return
  for d in r:
    if d[3] == '微信':
      user = wx_staff_data.get(d[4], "")
      sendCardmessage("**跟进人: **<at id=%s></at> **联系人: **<at id=%s></at>\n**用户ID: **%s\n**联系%s: **%s\n**客户备注: ** %s\n**跟进记录: **%s" %(get_feishu_id(d[2]), get_feishu_id(user), d[1], d[3], d[4], d[5], d[6]), d[0])
    elif d[3] == '企微':
      user = ent_staff_data.get(d[4], "")
      sendCardmessage("**跟进人: **<at id=%s></at> **联系人: **<at id=%s></at>\n**用户ID: **%s\n**联系%s: **%s\n**客户备注: ** %s\n**跟进记录: **%s" %(get_feishu_id(d[2]), get_feishu_id(user), d[1], d[3], d[4], d[5], d[6]), d[0])
    elif d[3] == '企点':
      user = tencent_qd_staff_data.get(d[4], "")
      sendCardmessage("**跟进人: **<at id=%s></at> **联系人: **<at id=%s></at>\n**用户ID: **%s\n**联系%s: **%s\n**客户备注: ** %s\n**跟进记录: **%s" %(get_feishu_id(d[2]), get_feishu_id(user), d[1], d[3], d[4], d[5], d[6]), d[0])
    elif d[3] == 'QQ':
      user = tencent_qq_staff_data.get(d[4], "")
      sendCardmessage("**跟进人: **<at id=%s></at> **联系人: **<at id=%s></at>\n**用户ID: **%s\n**联系%s: **%s\n**客户备注: ** %s\n**跟进记录: **%s" %(get_feishu_id(d[2]), get_feishu_id(user), d[1], d[3], d[4], d[5], d[6]), d[0])





def get_feishu_id(user_name):
  kdlstaff_db = DB(user='kdlstaff', password='kdlstaff', dbname='kdlstaff', host='10.0.5.15 ', port=3306, charset='utf8', autocommit=True)
  kdlstaff_db.connect()
  sql = """select feishu_user_id from user_profile where nickname = '%s'""" %(user_name)
  r = kdlstaff_db.execute(sql).fetchone()
  kdlstaff_db.close()
  return r[0]


if __name__ == '__main__':
  week = date.today().weekday()
  if week == 0:
      start_day = (date.today() + timedelta(-3)).strftime("%Y-%m-%d %H:%M:%S")
      end_day = (date.today() + timedelta(-1)).strftime("%Y-%m-%d %H:%M:%S")
  elif week in (1, 2, 3, 4):
      start_day = (date.today() + timedelta(-1)).strftime("%Y-%m-%d %H:%M:%S")
      end_day = (date.today()).strftime("%Y-%m-%d %H:%M:%S")
  get_customer_trace_data(start_day, end_day)
