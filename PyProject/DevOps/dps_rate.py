# -*- coding: utf-8 -*-

"""

周期性运行，给特定机器评分。
将打好分的机器写回数据库里。
取出未打分的机器，交换到特殊组里。

查找过去3天内数据

耗时：
机器默认品质 = 3，数据小于1400次不进入统计，
品质1 time_connect  > 10000 发生 1400 次 的天数，大于3天
品质2 time_connect  > 5000 发生 700 次 的天数，大于3天
默认3 time_connect  < 5000 发生 2100 次 的天数，大于1天
品质4 time_connect  < 1000 发生 2100 次 的天数，大于2天
品质5 time_connect  < 500 发生 2100 次 的天数，大于3天

带宽：
机器默认品质 = 3，数据小于450次(全天450条数据)不进入统计，排序，取中段数值
品质1 带宽 1M
品质2 带宽 3M
默认3 带宽 5M
品质4 带宽 7M
品质5 带宽 10M

运行方式：cron.d运行
启动命令：python -u /home/httpproxy/devops/dps_rate.py >/dev/null 2>&1 &

最终得分 = 耗时 * 0.5 + 带宽 * 0.5，向下取整

"""
import sys
import time
import datetime
import random

import loggerutil
import dbutil
import ckutil
from redisutil import redisdb_crs

logger = loggerutil.get_logger("devops", "devops/dps_rate.log")
db = dbutil.get_db_db()
ck = ckutil.get_ck_nodeops()

RETA_CONNECT_COUNT = 1400
RATE_CONNECT_DAY = 3
RATE_CONNECT_CONFIG = [
    {"connect" : "connect<1000", "num": 2100, "day" : 2, "score" : 4},
    {"connect" : "connect<500", "num" : 2100, "day" : 3, "score" : 5},
    {"connect" : "connect>5000", "num" : 700, "day" : 3, "score" : 2},
    {"connect" : "connect>10000", "num" : 1400, "day" : 3, "score" : 1 },
    ]
RETA_NETWORK_COUNT = 450
RATE_NETWORK_CONFIG = {5:655360, 4:458752, 3:327680, 2:233168}

def process_connect(codes):
    result = {}
    # 有效的都默认3分
    valid_list = process_connect_valid(codes)
    result.update({code: 3 for code in valid_list})
    logger.info("valid length:%d" % (len(valid_list)))
    # 根据情况打分
    for config in RATE_CONNECT_CONFIG:
        score = config["score"]
        dps_list = process_connect_config(config, codes)
        result.update({code:score for code in dps_list})
        logger.info("[score=%d]:%d" % (score, len(dps_list)))

    return result

def process_connect_valid(codes):
    result = []
    num = RETA_CONNECT_COUNT
    edge_start = (datetime.datetime.now() - datetime.timedelta(days=1+RATE_CONNECT_DAY)).strftime("%Y-%m-%d 00:00:00")
    edge_end = (datetime.datetime.now() - datetime.timedelta(days=1)).strftime("%Y-%m-%d 23:59:59")
    sql = """select code, toDate(stat_time) as st, count(*) as ct from node_connect_history 
    where stat_time > '%s' and stat_time < '%s' group by code,st having ct>%d 
    """ % (edge_start, edge_end, num)
    rows = ck.execute(sql)
    if rows:
        days = {code:0 for code in set([item[0] for item in rows ])}
        for row in rows:
            code = row[0]
            if (codes and code in codes) or (not codes):
                days[code] = days[code] + 1
        result = [k for k,v in days.items() if v>=RATE_CONNECT_DAY]

    return result

def process_connect_config(config, codes):
    result = []
    connect = config["connect"]
    num = config["num"]
    day = config["day"]
    edge_start = (datetime.datetime.now() - datetime.timedelta(days=1+RATE_CONNECT_DAY)).strftime("%Y-%m-%d 00:00:00")
    edge_end = (datetime.datetime.now() - datetime.timedelta(days=1)).strftime("%Y-%m-%d 23:59:59")
    sql = """select code, toDate(stat_time) as st, count(*) as ct from node_connect_history 
    where stat_time > '%s' and stat_time < '%s' and %s group by code, st having ct>%d 
    """ % (edge_start, edge_end, connect, num)
    rows = ck.execute(sql)
    if rows:
        days = {code:0 for code in set([item[0] for item in rows ])}
        for row in rows:
            code = row[0]
            if (codes and code in codes) or (not codes):
                days[code] = days[code] + 1
        result = [k for k,v in days.items() if v>=day]

    return result

def process_network(codes):
    result = {}
    edge_start = (datetime.datetime.now() - datetime.timedelta(days=1)).strftime("%Y-%m-%d 00:00:00")
    edge_end = (datetime.datetime.now() - datetime.timedelta(days=1)).strftime("%Y-%m-%d 23:59:59")
    sql = """select code, bandwidth_recv, bandwidth_send from node_sysload_history 
        where create_time > '%s' and create_time < '%s' 
        """ % (edge_start, edge_end, )
    rows = ck.execute(sql)
    if rows:
        networks = {code: [] for code in set([item[0] for item in rows])}
        for row in rows:
            code, recv = row[0], row[1]
            if (codes and code in codes) or (not codes):
                networks[code].append(recv)
        for code, dps_network in networks.items():
            if dps_network and len(dps_network) >= RETA_NETWORK_COUNT:
                # 取中段
                dps_network.sort()
                index = int(len(dps_network) / 3)
                dps_network = dps_network[index:-index]
                bandwidth = sum(dps_network) / len (dps_network)
                # 得分
                score = 1
                keys = list(RATE_NETWORK_CONFIG.keys())
                keys.sort(reverse=True)
                for key in keys:
                    if RATE_NETWORK_CONFIG[key] <= bandwidth:
                        score = key
                        break
                result[code] = score
                # logger.info("[%s]bandwidth=%d, network score=%d" % (code, bandwidth, score))
    return result

def process(connect_dict, network_dict, codes):
    rated_dict = {}
    if codes:
        for code in codes:
            redisdb_crs.hset("dps_rate", code, 0)
    else:
        redisdb_crs.delete("dps_rate")
        codes = list(set(connect_dict.keys()) & set(network_dict.keys()))

    for code in codes:
        if code in connect_dict and code in network_dict:
            connect = connect_dict[code]
            network = network_dict[code]
            score = int(connect * 0.5 + network * 0.5)
            redisdb_crs.hset("dps_rate", code, score)
            rated_dict[code] = score
    return rated_dict

def get_rate_dps_list():
    edge_time = (datetime.datetime.now() - datetime.timedelta(days=1 + RATE_CONNECT_DAY)).strftime("%Y-%m-%d 00:00:00")
    sql = """select d.code from dps d, dpsgroup dgp, dps_group_relation dgr WHERE d.id=dgr.dps_id and dgr.group_id=dgp.id 
             and d.status=1 and dgr.update_time<'%s' and dgp.code='rate';""" % (edge_time, )
    cursor = db.execute(sql)
    rows = cursor.fetchall()

    dps_list = [item[0] for item in rows]
    logger.info("dps count=%d" % (len(dps_list)))
    return dps_list

def write_score(rated_dict):
    data_list = [(score, code) for code,score in rated_dict.items() if score]
    dps_list = [item[1] for item in data_list]
    sql = "update dps set quality=%s where code=%s"
    db.executemany(sql, data_list)
    logger.info("dps rated length:%d" % (len(data_list)))
    return dps_list

def get_all_dps_dict():
    sql = """select code, id, status, changeip_period, quality, is_ops_node from dps where status in (1, 3)"""
    cursor = db.execute(sql)
    rows = cursor.fetchall()

    dps_dict = {item[0]: {"id": int(item[1]), "status": int(item[2]), "changeip_period": int(item[3]), "quality": int(item[4]), "is_ops_node": int(item[5])} for item in rows}
    return dps_dict

def switch_node(rated_list):
    dps_dict = get_all_dps_dict()
    done_list = [item for key, item in dps_dict.items() if key in rated_list]
    todo_list = [item for key, item in dps_dict.items() if item["status"] == 1 and item["quality"] == 0 and item["is_ops_node"] == 0]
    update_list = zip(done_list, todo_list)
    map(update_dps_group, update_list)

# 交换分组只需要交换group_id对应的dps_id即可,这样不论节点有多少组关系都能更换
def update_dps_group((node1, node2)):
    sql = """update dps set changeip_period=%s where id=%s"""
    change_list = [(node1["changeip_period"], node2["id"]), (node2["changeip_period"], node1["id"])]
    date = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    sql2 = """update dps_group_relation set dps_id=%s,update_time='%s' where dps_id=%s""" % (
    node1["id"], date, node2["id"])
    sql3 = """update dps_group_relation set dps_id=%s,update_time='%s' where dps_id=%s and not update_time='%s'""" % (
    node2["id"], date, node1["id"], date)
    try:
        logger.info("switch %d <-> %d" % (node1["id"], node2["id"]))
        db.executemany(sql, change_list)
        db.execute(sql2)
        db.execute(sql3)
    except Exception as e:
        logger.exception(str(e))
        return False
    return True

def main():
    # 取出rate组里的机器，并进行打分。
    codes = get_rate_dps_list()
    connect_dict = process_connect(codes)
    network_dict = process_network(codes)
    rated_dict = process(connect_dict, network_dict, codes)
    # 将打好分的机器写回数据库里
    rated_list = write_score(rated_dict)
    # 取出未打分的机器，进行交换
    switch_node(rated_list)

if __name__ == '__main__':
    main()