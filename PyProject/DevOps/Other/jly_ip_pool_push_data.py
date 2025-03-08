#!/usr/bin/env python
# -*- coding: utf-8 -*-

import redis
import staffnotify
from datetime import date, timedelta


day_today = (date.today()).strftime("%F")
yesterday = (date.today() + timedelta(-1)).strftime("%F")
datetime_week_dict = {0: '周一', 1: '周二', 2: '周三', 3: '周四', 4: '周五', 5: '周六', 6: '周日'}

redisdb = redis.StrictRedis(host='localhost', port=6379, db=3, decode_responses=True)

key = ":Pool" + "-*-" + yesterday

long_period = [7200, 10800, 14400]

long_period_6_12 = [21600, 43200]

def get_data():
    pool_periods = {
    '60_300': [],
    '300_600': [],
    '600_1200': [],
    '1200_1800': [],
    '1800_3600': [],
    '7200_14400': [],
    '21600_43200': [],
    }
    for i in range(60, 3660, 60):
        keys = redisdb.keys('%s' %(str(i) + key))
        sorted_keys = sorted(keys)
        ip_num, data = calculate_data(sorted_keys)
        period = str(i) + "s"
        if i >= 60 and i < 300:
            pool_periods['60_300'].append("%s: %s(%s)" % (period, ip_num, ' '.join(data)))
        elif i >= 300 and i < 600:
            pool_periods['300_600'].append("%s: %s(%s)" % (period, ip_num, ' '.join(data)))
        elif i >= 600 and i < 1200:
            pool_periods['600_1200'].append("%s: %s(%s)" % (period, ip_num, ' '.join(data)))
        elif i >= 1200 and i < 1800:
            pool_periods['1200_1800'].append("%s: %s(%s)" % (period, ip_num, ' '.join(data)))
        elif i >= 1800 and i <= 3600:
            pool_periods['1800_3600'].append("%s: %s(%s)" % (period, ip_num, ' '.join(data)))

    for t in long_period:
        period = str(t) + "s"
        keys = redisdb.keys('%s' %(str(t) + key))
        sorted_keys = sorted(keys)
        ip_num, data = calculate_data(sorted_keys)
        pool_periods['7200_14400'].append("%s: %s(%s)" % (period, ip_num, ' '.join(data)))

    for d in long_period_6_12:
        period = str(d) + "s"
        keys = redisdb.keys('%s' %(str(d) + key))
        sorted_keys = sorted(keys)
        ip_num, data = calculate_data(sorted_keys)
        pool_periods['21600_43200'].append("%s: %s(%s)" % (period, ip_num, ' '.join(data)))
    return pool_periods


def structure_data(data):
    s = ""
    count = 1
    for i in data:
        if count == 2:
            s = s + " " + str(i) + '\n'
            count = 1
        else:
            s = s + str(i)
            count += 1
    return s

def calculate_data(keys):
    count = 0
    source_data = []
    for i in keys:
        r = redisdb.get(i)
        source_data.append("%s : %s" %(i.split('-', 2)[1], r))
        count +=int(r)
    return count, source_data

# 飞书body模板，勿随意修改
def make_body(data_overview):
    body = {
        "open_chat_id": '',
        "msg_type": "interactive",
        "update_multi": True,
        "card": {
            "config": {
                "wide_screen_mode": True
            },
            "header": {
                "template": "green",
                "title": {
                    "content": "📊 IP池数量统计 %s（%s）" % (day_today, datetime_week_dict.get((date.today()).weekday())),
                    "tag": "plain_text"
                }
            },
            "elements": [
                {
                    "tag": "div",
                    "text": {
                        "content": "[1-5分钟IP池]\n%s" %('\n'.join(data_overview['60_300'])),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": "[5-10分钟IP池]\n%s" %(structure_data(data_overview['300_600'])),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": "[10-20分钟IP池]\n%s" %(structure_data(data_overview['600_1200'])),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": "[20-30分钟IP池]\n%s" %(structure_data(data_overview['1200_1800'])),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": "[30-60分钟IP池]\n%s" %(structure_data(data_overview['1800_3600'])),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": "[2-4小时IP池]\n%s" %(' '.join(data_overview['7200_14400'])),
                        "tag": "lark_md"
                    }
                },
                {
                    "tag": "hr"
                },
                {
                    "tag": "div",
                    "text": {
                        "content": "[6-12小时IP池]\n%s" %(' '.join(data_overview['21600_43200'])),
                        "tag": "lark_md"
                    }
                },
            ],
        }
    }
    staffnotify.notify_feishu_card_customized(body)



if __name__ == '__main__':
    data = get_data()
    make_body(data)
