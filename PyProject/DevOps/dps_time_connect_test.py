# -*- coding: utf-8 -*-

"""

周期性运行，查找最近一分钟内，time_connect 耗时异常的机器。
先打日志观察。

运行方式：cron.d运行
启动命令：python -u /home/httpproxy/devops/dps_time_connect_test.py >/dev/null 2>&1 &

"""
import sys
import datetime
import random
import time
import datetime
import json
import requests

import loggerutil
import dbutil
import ckutil

from redisutil import redisdb, redisdb_crs

logger = loggerutil.get_logger("devops", "devops/dps_time_connect_test.log")
db = dbutil.get_db_db()
ck = ckutil.get_ck_nodeops()

TIME_CONNECT_THRESHOLD = 10000
TIME_CONNECT_COUNT = 1000
DAY_THRESHOLD = 5


def get_dps_dpsgroup_dict():
    """得到所有dps节点对应的分组情况"""
    sql = """select d.code, d.changeip_period, dg.code as group_code from dps d,dps_group_relation dgr ,dpsgroup dg where d.id=dgr.dps_id and dg.id=dgr.group_id;"""
    cursor = db.execute(sql)
    rows = cursor.fetchall()
    codes = set([row[0] for row in rows])
    dps_dict = {row[0]:row[1] for row in rows}
    dps_dpsgroup_dict = {code:[] for code in codes}
    for row in rows:
        dps_dpsgroup_dict[row[0]].append(row[2])

    return dps_dict, dps_dpsgroup_dict

def get_connect_dict():
    """查找最近24小时内，机器的平均耗时"""
    data = {}
    edge_end = (datetime.datetime.now() - datetime.timedelta(days=1)).strftime("%Y-%m-%d %H:00:00")
    key = "time_connect_last_day_" + edge_end
    if not redisdb.exists(key):
        sql = """select code, avg(connect) from node_connect_history 
        where stat_time > '%s' group by code
        """ % (edge_end, )
        rows = ck.execute(sql)
        if rows:
            data = {row[0]:int(row[1]) for row in rows}
            redisdb.set_obj(key, data, 4000)
    else:
        data = redisdb.get_obj(key)

    return data

def process_data(dps_dict, dpsgroup_dict, connect_dict):
    exclude_set = set(['c7', 'c20', 'c9', 'c30', 'c23', 'c5', 'c26', 'c1', 'c4', 'c21', 'c22', 'c25', 'c39', 'c27', 'c8', 'g14' ,'g15', 'g22', 'g21', 'g25'])  # 排除组
    current_connect_dict = redisdb_crs.hgetall("code_connected_timeuse")
    for code, dpsgroup_list in dpsgroup_dict.items():
        changeip_period = dps_dict[code]
        if 600 < changeip_period < 2000:
            #c g 组才统计
            if not (set(dpsgroup_list) & exclude_set) :
                # 排除组外的才统计
                if code in connect_dict:
                    avg = int(connect_dict[code])
                    current = int(current_connect_dict[code])
                    if avg and avg < 5000 and (current > avg + 30000):
                        # 超过平均时间30s
                        logger.info("[%s] avg=%s, current=%s" % (code, avg, current))


def main():
    dps_di, dpsgroup_di = get_dps_dpsgroup_dict()
    connect_di = get_connect_dict()
    process_data(dps_di, dpsgroup_di, connect_di)

if __name__ == '__main__':
    main()