#!/usr/bin/env python
# -*- coding: utf-8 -*-
# author_='Joinishu';
# date: 2021/9/10 10:26
# desc: ip独享型续费提醒
# 运行方式: cron每天早上9:10推送

from datetime import datetime
import parser
import time

import dbutil
import loggerutil
import staffnotify
from dateutil import parser

kdl_db = dbutil.get_db_db()
get_time = datetime.now()
logger = loggerutil.get_logger("devops", "devops/exception_node_auto_reboot.log")
report_list = []


def timedelta_total_seconds(timedelta):
    """得到timedelta的秒数"""
    return (timedelta.microseconds + 0.0 + (timedelta.seconds + timedelta.days * 24 * 3600) * 10 ** 6) / 10 ** 6


def timedelta_muti(expire_time, end_time, only_delta=False):
    end_time = parser.parse(end_time)
    expire_time = parser.parse(expire_time)
    delta = timedelta_total_seconds(end_time - expire_time)

    if only_delta: return delta
    if (delta < 0): return '<span class="price">已过期</span>'
    if (delta < 60): return "%s秒" % delta
    if (delta < 3600): return "%s分钟" % int(round(float(delta) / 60))
    if (delta < 3600 * 24): return "%.1f小时" % round(float(delta) / 3600, 2)
    return "%.1f天" % round(float(delta) / (3600 * 24), 2)


def get_order(pre_day):
    sql = "select proxy_order.orderid,kps.code,kps.provider,kps.ip,proxy_order.end_time,kps.expire_time from order_kps,kps,proxy_order where " \
          "order_id=proxy_order.id and kps_id=kps.id and kps.expire_time between DATE_FORMAT(NOW(), '%%Y-%%m-%%d 00:00:00') and DATE_FORMAT((DATE_SUB(NOW(),INTERVAL %s day)),'%%Y-%%m-%%d 23:59:59') and kps.level=5 and is_valid=1 " \
          "order by proxy_order.end_time asc" % pre_day
    cursor = kdl_db.execute(sql)
    rows = cursor.fetchall()
    if rows:
        for row in rows:
            orderid = row[0]
            code = row[1]
            provider = row[2]
            ip = row[3]
            order_end_time = row[4]
            kps_expire_time = row[5]
            if order_end_time and kps_expire_time and kps_expire_time < order_end_time:
                if kps_expire_time < get_time:
                    title = "%s [%s -- %s] 已在订单 %s 到期前到期！" % (code, ip, provider, orderid)
                else:
                    interval = timedelta_muti(str(kps_expire_time), str(order_end_time))
                    title = "%s [%s -- %s] 将在订单 %s 到期前%s到期！" % (code, ip, provider, orderid, interval)
                report_list.append(title)
    if len(report_list) != 0:
        return report_list
    else:
        return False


if __name__ == '__main__':
    report_list = get_order(-1)
    descr = '独享型kps续费提醒'
    if report_list:
        report_list = '\n'.join(report_list)
        try:
            res = staffnotify.staff_api_post("https://staff.gizaworks.com/api/create_task/", tt=4, ts='11', descr=descr, pu='https://stat.gizaworks.com/stat/kps_exip_order', memo=report_list)
        except Exception as e:
            logger.error("任务发送失败：%s" % str(e))
