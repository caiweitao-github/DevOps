# -*- coding: utf-8 -*-

"""

周期性运行，根据换IP记录和负载记录，给机器打分

运行方式：cron.d运行
启动命令：python -u /home/httpproxy/dps_master/update_dps_cache2.py 2>&1 &

"""
import sys
import datetime

import dbutil
from redisutil import redisdb, redisdb_crs
import loggerutil

logger = loggerutil.get_logger("devops", "devops/dps_devops_grade.log")
db = dbutil.get_db_db()
db_nodeops = dbutil.get_db_nodeops()

default_unexpected_rule = [(200, -2), (320, -2), (630, -3), (2760, -5), (3660, -8), (10800, -10), (14400, -12), (sys.maxint, -12)]
default_timeout_rule = [(60, -1), (90, -3), (180, -5), (sys.maxint, -10)]
default_load_rule = [(3.0, -1), (5.0, -2), (8.0, -5), (10.0, -10), (15.0, -20), (100.0, -30), (sys.maxint, -50)]

class ExceptionCategory:
    Unexpected = 1
    Timeout = 2
    Load = 3

def add_record(dps_code, exception_time, category, info, score, constant=False):
    logger.info("[%s] %d:%d" % (dps_code, category, score))
    sql = """insert into dps_devops_grade (dps_code, exception_time, category, info, score, is_constant, grade_time )
    values('%s', '%s', %d, '%s', %d, %d, now())""" % (dps_code, exception_time, category, info, score, constant)
    # logger.info(sql)
    db_nodeops.execute(sql)

def changeip_grade():
    # 读取规则
    unexpected_rule = default_unexpected_rule
    try:
        rule = redisdb_crs.get("devops_unexpected_rule")
        if rule:
            unexpected_rule = eval(rule)
    except Exception as e:
        pass

    timeout_rule = default_timeout_rule
    try:
        rule = redisdb_crs.get("devops_timeout_rule")
        if rule:
            timeout_rule = eval(rule)
    except Exception as e:
        pass

    # 读取最近半小时换IP记录
    sql = """select dps_code, is_expected, change_time, change_interval, changeip_period, timeuse
            from dps_changeip_history
            where change_time >= CURRENT_TIMESTAMP - INTERVAL 30 MINUTE"""
    cursor = db_nodeops.execute(sql)
    rows = cursor.fetchall()

    codes = set([item[0] for item in rows])
    constant_dict = {code: False  for code in codes}

    for r in rows:
        dps_code, is_expected, change_time, change_interval, changeip_period, timeuse = r[0], r[1], r[2], r[3], r[4], r[5]

        # 异常换IP，当天次数过多，则扣分也多。同时长效IP，单次发生的扣分也多。
        if not is_expected:
            score = 0
            constant = True if constant_dict[dps_code] else False
            constant_dict[dps_code] = True
            for item in unexpected_rule:
                if changeip_period <= item[0]:
                    score = item[1]
                    break
            add_record(dps_code, change_time, ExceptionCategory.Unexpected, '', score, constant)
        else:
            constant_dict[dps_code] = False

        # 换IP时间过长，超过60s，则扣分，越长扣分越多。
        if timeuse > 60:
            score = 0
            for item in timeout_rule:
                if timeuse <= item[0]:
                    score = item[1]
                    break
            add_record(dps_code, change_time, ExceptionCategory.Timeout, '', score)

def get_alive_dps():
    alive_dps_list = []
    sql = "select code from dps where status<>4"
    cursor = db.execute(sql)
    rows = cursor.fetchall()
    if rows:
        alive_dps_list = [row[0] for row in rows]
    return alive_dps_list

def parse_loadinfo(sysload):
    usage = 0
    try:
        if len(sysload.split("|")) == 4:
            timestamp, load, meminfo, netinfo = sysload.split("|")
            usage = load.strip().split(" ")[0]
    except Exception as e:
        pass
    return float(usage)

def stat_grade():
    # 读取规则
    load_rule = default_load_rule
    try:
        rule = redisdb_crs.get("devops_load_rule")
        if rule:
            load_rule = eval(rule)
    except Exception as e:
        pass

    # 负载异常，大于阈值时长越长，则扣分越多。大于阈值数值越大，扣分越多。
    alive_dps_list = get_alive_dps()
    for dps_code in alive_dps_list:
        sysload_list = redisdb_crs.hget_obj("dps_sysload_hour", dps_code)
        if sysload_list:
            sysload_list = sysload_list[:30]
            loads = map(parse_loadinfo, sysload_list)
            load = sum(loads) / len(loads)
            if load > 3.0:
                score = 0
                for item in load_rule:
                    if load <= item[0]:
                        score = item[1]
                        break
                add_record(dps_code, datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S') , ExceptionCategory.Load, '', score)


if __name__ == '__main__':
    changeip_grade()
    stat_grade()