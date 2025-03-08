#!/usr/bin/env python
# -*- coding: utf-8 -*-

from datetime import date, timedelta
import json

import requests

import dbutil

db_db = dbutil.get_db_db()

day_today = (date.today()).strftime("%F")
start_day = (date.today() + timedelta(-7)).strftime("%Y-%m-%d %H:%M:%S")
end_day = (date.today()+ timedelta(-1)).strftime("%Y-%m-%d %H:%M:%S")

category = {
    1: "到期跟进",
    2: "流失跟进",
    3: "日常跟进",
    4: "潜在流失跟进",
    5: "异常跟进"
}

issues_category_exact = {
    0: "其他",
    1: "反爬",
    2: "建议",
    3: "质量",
    4: "价格",
    5: "客户本身的问题",
    6: "产品相关",
    7: "业务相关",
    8: "购买相关"
}

def send_mess(body):
    try:
        url = ''
        headers = {"Content-Type": "application/json"}
        body = body
        requests.post(url,headers=headers, data=json.dumps(body))
    except Exception as e:
        print(e)

def get_renew_track_count():
    data = {}
    sql = """select trace_category,sum(renew_price),count(*) from big_customer_trace 
    where  create_time between '%s' and '%s' and trace_feedback = 0 and renew_price > 0 group by trace_category order by trace_category""" %(start_day, end_day)
    r = db_db.execute(sql).fetchall()
    for i in r:
        data[i[0]] = {float(i[1]): i[2]}
    return data


# def get_user_trace_data(d1, d2, d3, d4, d5, d6, d7, d8):
#     res = get_renew_track_count()
#     count = sum([d1, d2, d3, d4, d5, d6, d7, d8])
#     return ("**[用户跟进数据]**\n"
#         "流失跟进: %s 人(有效跟进: %s 无效跟进: %s) %s\n"
#         "日常跟进: %s 人(有效跟进: %s 无效跟进: %s) %s\n"
#         "潜在流失跟进: %s人(有效跟进: %s 无效跟进: %s) %s\n"
#         "到期跟进: %s 人(有效跟进: %s 无效跟进: %s) %s\n"
#         "用户跟进总数: %s 人(有效跟进: %s, 无效跟进: %s)" % 
#         (d1+d2, d1, d2, "已续费: %s 人 续费总金额: %s 元" %(res.get(2).values()[0], res.get(2).keys()[0]) if res.get(2) else "", 
#         d3+d4, d3, d4, "已续费: %s 人 续费总金额: %s 元" %(res.get(3).values()[0], res.get(3).keys()[0]) if res.get(3) else "",
#         d5+d6, d5, d6, "已续费: %s 人 续费总金额: %s 元" %(res.get(4).values()[0], res.get(4).keys()[0]) if res.get(4) else "",
#         d7+d8, d7, d8, "已续费: %s 人 续费总金额: %s 元" %(res.get(1).values()[0], res.get(1).keys()[0]) if res.get(1) else "",
#         count, d1+d3+d5+d7, d2+d4+d6+d8))


def get_user_trace_data(d1, d2, d3, d4, d5, d6, d7, d8, d9, d10):
    count = sum([d1, d2, d3, d4, d5, d6, d7, d8, d9, d10])
    return ("**[用户跟进数据]**\n"
        "流失跟进: %s 人(有效跟进: %s 无效跟进: %s)\n"
        "日常跟进: %s 人(有效跟进: %s 无效跟进: %s)\n"
        "潜在流失跟进: %s 人(有效跟进: %s 无效跟进: %s)\n"
        "到期跟进: %s 人(有效跟进: %s 无效跟进: %s)\n"
        "异常跟进: %s 人(有效跟进: %s, 无效跟进: %s)\n"
        "用户跟进总数: %s 人(有效跟进: %s, 无效跟进: %s)" % 
        (d1+d2, d1, d2, d3+d4, d3, d4, d5+d6, d5, d6, d7+d8, d7, d8, d9+d10, d9, d10, count, d1+d3+d5+d7+d9, d2+d4+d6+d8+d10))

def get_user_trace_invalid(vaild_data, invaild_data):
    daily_invalid_ex = 0
    daily_invalid_loss = 0
    daily_invalid_daily = 0
    daily_invalid_potential_loss = 0
    daily_invalid_abnormal = 0
    daily_invalid = {}
    text = "**[详细跟进数据]**\n"
    sql = """select au.first_name 'new_user',trace_category,count(*) from django_admin_log dl  
    left join big_customer_trace bt on bt.id =dl.object_id  left join auth_user au on au.id = dl.user_id 
    where  dl.action_time between '{}' and '{}'  and bt.trace_time between '{}' and '{}' and dl.change_message 
    like '%trace_record%'  and dl.change_message like '%trace_time%' and content_type_id =82 and action_flag =2 
    group by au.first_name,trace_category order by new_user""".format(start_day, end_day, start_day, end_day)
    res = db_db.execute(sql).fetchall()
    for i in res:
        key = "%s" %(i[0])
        if daily_invalid.get(key):
            daily_invalid[key].update({int(i[1]): int(i[2])})
        else:
            daily_invalid[key] = {int(i[1]): int(i[2])}
        if int(i[1]) == 1:
            daily_invalid_ex += int(i[2])
        elif int(i[1]) == 2:
            daily_invalid_loss += int(i[2])
        elif int(i[1]) == 3:
            daily_invalid_daily += int(i[2])
        elif int(i[1]) == 4:
            daily_invalid_potential_loss += int(i[2])
        elif int(i[1]) == 5:
            daily_invalid_abnormal += int(i[2])
    for x in invaild_data.values():
        for status_type, count_num in x.items():
            if int(status_type) == 1:
                daily_invalid_ex += int(count_num)
            elif int(status_type) == 2:
                daily_invalid_loss += int(count_num)
            elif int(status_type) == 3:
                daily_invalid_daily += int(count_num)
            elif int(status_type) == 4:
                daily_invalid_potential_loss += int(count_num)
            elif int(status_type) == 5:
                daily_invalid_abnormal += int(count_num)
    r = merge_dicts(invaild_data, daily_invalid)
    result = merge_dicts(vaild_data, r)
    for k,v in result.items():
        num = 0
        for n in v.values():
            num += n
        dict_key = "%s" %(k)
        if vaild_data.get(dict_key):
            text = text + '%s 总人数: %d\n\t有效跟进: %d人 (' %(k, num, sum(vaild_data[dict_key].values()))
            for txt, count in vaild_data[dict_key].items():
                text = text + '%s:%s ' %(category[int(txt)], count)
            text = text + ')' + '\n'

        if r.get(dict_key):
            text = text + '\t无效跟进: %d人 (' %(sum(r[dict_key].values()))
            for txt, count in r[dict_key].items():
                text = text + '%s:%s ' %(category[int(txt)], count)
            text = text + ')' + '\n'
    return text, daily_invalid_loss, daily_invalid_daily, daily_invalid_potential_loss, daily_invalid_ex, daily_invalid_abnormal
    

def get_user_trace_valid():
    daily_valid_ex = 0
    daily_valid_loss = 0
    daily_valid_daily = 0
    daily_valid_potential_loss = 0
    daily_valid_abnormal = 0
    daily_valid, daily_invalid = {}, {}
    daily_valid_sql = """select au.first_name,trace_category,trace_feedback,count(*) from big_customer_trace bt,auth_user au where bt.staff_id=au.id and 
    trace_time between '%s' and '%s' group by trace_category,first_name,trace_feedback order by first_name""" %(start_day, end_day)
    res = db_db.execute(daily_valid_sql).fetchall()
    for i in res:
        key = "%s" %(i[0])
        if i[2] == 1:
            if daily_invalid.get(key):
                daily_invalid[key].update({int(i[1]): int(i[-1])})
            else:
                daily_invalid[key] = {int(i[1]): int(i[-1])}
            continue
        else:
            if daily_valid.get(key):
                daily_valid[key].update({int(i[1]): int(i[-1])})
            else:
                daily_valid[key] = {int(i[1]): int(i[-1])}
        if int(i[1]) == 1:
            daily_valid_ex += int(i[-1])
        elif int(i[1]) == 2:
            daily_valid_loss += int(i[-1])
        elif int(i[1]) == 3:
            daily_valid_daily += int(i[-1])
        elif int(i[1]) == 4:
            daily_valid_potential_loss += int(i[-1])
        elif int(i[1]) == 5:
            daily_valid_abnormal += int(i[-1])
    return daily_valid, daily_invalid, daily_valid_loss, daily_valid_daily, daily_valid_potential_loss, daily_valid_ex, daily_valid_abnormal


def merge_dicts(a, b):
    merged_dict = {}
    for key in set(a.keys()).union(b.keys()):
        if key in a and key in b:
            if isinstance(a[key], dict) and isinstance(b[key], dict):
                merged_dict[key] = merge_dicts(a[key], b[key])
            else:
                merged_dict[key] = a[key] + b[key]
        elif key in a:
            merged_dict[key] = a[key]
        else:
            merged_dict[key] = b[key]
    return merged_dict


def get_issues_category_data():
    text = ["**[问题类型详情]**"]
    sql = """select issues_category,count(*)  from(
        select issues_category from django_admin_log dl 
        left join big_customer_trace bt on bt.id =dl.object_id 
        left join auth_user au on au.id = dl.user_id
        where dl.action_time between '{}' and '{}' and dl.change_message like '%trace_record%'
        and bt.trace_time between '{}' and '{}' and dl.change_message like '%trace_time%' 
        and content_type_id =82 and action_flag =2
        union all select issues_category from big_customer_trace bt,auth_user au
        where bt.staff_id=au.id and trace_time between '{}' and '{}') a group by issues_category""".format(start_day, end_day, start_day, end_day, start_day, end_day)
    res = db_db.execute(sql).fetchall()
    for i in res:
        text.append("%s: %s人" %(issues_category_exact[int(i[0])], i[1]))
    return '\n'.join(text)

# 飞书body模板，勿随意修改
def make_body(data2_overview, data4_overview, data5_overview):
    body = {
        "msg_type": "interactive",
        "update_multi": True,
        "card": {
            "config": {
                "wide_screen_mode": True
            },
            "header": {
                "template": "green",
                "title": {
                    "content": "🏳️‍🌈 技术支持数据周报 %s" % (day_today),
                    "tag": "plain_text"
                }
            },
            "elements": [
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": data2_overview.decode('utf-8'),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": data4_overview.decode('utf-8'),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": data5_overview,
                        "tag": "lark_md"
                    }
                },
            ],
        }
    }
    send_mess(body)


if __name__ == '__main__':
    try:
        trace_overview_valid, trace_overview_invalid_tmp, d1, d2, d3, d4, d9 = get_user_trace_valid()
        trace_overview_invalid, d5, d6, d7, d8, d10 = get_user_trace_invalid(trace_overview_valid, trace_overview_invalid_tmp)
        user_trace_overview = get_user_trace_data(d1, d5, d2, d6, d3, d7, d4, d8, d9, d10)
        issues_category_data = get_issues_category_data()
        make_body(user_trace_overview, trace_overview_invalid, issues_category_data)
    except Exception as e:
        print(e)
