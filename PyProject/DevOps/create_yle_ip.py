#!/usr/bin/python
# -*- coding: utf-8 -*-
'''
分配柚雷节点
'''
import requests
import loggerutil
import random
import json
from datetime import datetime, timedelta
from area_code_dict import area_dict,prov_dict
from redisutil import redisdb_crs,redisdb
import dbutil

db_kdl = dbutil.get_db_kdlnode()

logger = loggerutil.get_logger("create_node", "yle_master/create_node.log")

all_count = 0

YL_TOKEN = 'yle_token'
YL_CITY = 'yle_city'
YL_Proportion = 'YL_PP'

def get_token():
    """获取token"""
    url = "http://api.ip119.cn/api/user/login"
    data = {
        "username": "db",
        "password": "db"
    }

    try:
        r = requests.post(url, data=data)
        if r.status_code != 200:
            return r.json()["msg"]
        else:
            token = r.json()["data"]["token"]
            expire_time = r.json()["data"]["expire_time"]
            redisdb.set(YL_TOKEN,token)
            redisdb.expire(YL_TOKEN,expire_time)
            return token
    except Exception as e:
        return e

def get_city_hash():
    """获取城市hash"""
    url = "http://api.ip119.cn/api/edge/city"
    token = redisdb.get(YL_TOKEN)
    if not token:
        token = get_token()
    data = {
        "token": token
    }
    try:
        r = requests.post(url, data=data)
        if r.status_code != 200:
            return r.json()["msg"]
        else:
            city_all = r.json()["data"]["list"]
            for province, city_hash in city_all.items():
                redisdb.hset_obj(YL_CITY, province, city_hash)
    except Exception as e:
        print(e)

def get_device_post_param(token, city_hash):
    """获取设备参数"""
    return {"token": token, "geo": city_hash, "offset": 0, "num": 200000, "onlyip": 1}

def get_mysql_code(code_name):
    """获取name的数量"""
    sql = """select count(*) from yle_node where code like "%s%s";"""%(code_name,'%')
    result = db_kdl.execute(sql).fetchone()
    return result[0]

def get_mysql_id(code_name):
    """获取name的id,ip"""
    sql = """select id,login_ip from yle_server where name =  "%s";"""%(code_name)
    result = db_kdl.execute(sql).fetchall()
    id = result[0][0]
    proxy_ip = result[0][1]
    return id,proxy_ip

def get_mysql_count(province_code):
    sql = """select count(*) from yle_node where province_code = "%s";"""%(province_code)
    result = db_kdl.execute(sql).fetchone()
    return result[0]

def insert_node(node_list):
    """插入节点"""
    if not node_list:
        return
    try:
        sql = """insert into yle_node 
        (yle_server_id, code, status, proxy_ip, proxy_port, city_code, province_code, bandwidth, changeip_period, bind_port, enable, last_changeip_time, create_time, update_time) 
        VALUES  (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, now(), now(), %s);"""
        db_kdl.executemany(sql, node_list)
        logger.info("insert_node success size :%s" % len(node_list))
    except Exception as e:
        logger.exception("insert_node ERROR: %s" % repr(e))

def get_percentage(province_count, node_sum):
    """获取比例对应的省份"""
    node_count = {}
    remaining_nodes = node_sum

    for province in province_count:
        province_code = prov_dict.get(province)
        percentage_str = redisdb.hget(YL_Proportion, province)
        province_percentage = float(percentage_str.replace('%', '')) / 100
        
        # 计算根据比例应分配的节点数
        allocated_nodes = int(node_sum * province_percentage)
        mysql_count = get_mysql_count(province_code)
        
        # 调整分配的节点数
        allocated_nodes -= mysql_count
        if allocated_nodes <= 0:
            allocated_nodes = 1
        
        node_count[province] = allocated_nodes
        remaining_nodes -= allocated_nodes

    # 确保省份没有分配为0节点
    for province in node_count:
        if remaining_nodes == 0:
            break
        if node_count[province] == 0:
            node_count[province] += 1
            remaining_nodes -= 1

    # 随机分配剩余的节点
    provinces = list(node_count.keys())
    if remaining_nodes > 0:
        for _ in range(remaining_nodes):
            province = random.choice(provinces)
            node_count[province] += 1

    return node_count

def assign_city_node(code_name, node_sum, time):
    """根据省份分配节点"""
    node_list = []

    province_count = redisdb.hkeys(YL_CITY)

    sum_node = get_mysql_code(code_name) + 1
    server_id,server_ip = get_mysql_id(code_name)
    node_cz = int(node_sum - sum_node +1)
    node_count = get_percentage(province_count,node_cz)
    for province in province_count:
        city = redisdb.hget(YL_CITY,province)
        # 解析 JSON 字符串
        data = json.loads(city)
        city_all = list(data.keys())
        province_code =  prov_dict.get(province)
        province_node = node_count[province]

        for _ in range(province_node):
            # 获取当前时间
            current_time = datetime.now()
            # 计算当前时间减去200秒的时间
            time_200_seconds_ago = current_time - timedelta(seconds=200)
            # 将时间格式化为指定的字符串格式
            update_time = time_200_seconds_ago.strftime("%Y-%m-%d %H:%M:%S")

            city = random.choice(city_all)
            city = str(city.encode("utf-8") + "市")
            city_code = area_dict.get(city)
            if not city_code:
                city_code = 'NULL'
            server_code = '{}{:03d}'.format(code_name, sum_node) if sum_node >= 100 else '{}{:02d}'.format(code_name, sum_node)
            port = 27000 + sum_node
            bind_port = 17000 + sum_node

            sql = """insert into yle_node (yle_server_id, code, status, proxy_ip, proxy_port, city_code, province_code, bandwidth, changeip_period, bind_port, enable, last_changeip_time, create_time, update_time) VALUES  (%s, '%s', %s, '%s', %s, %s, %s, %s, %s, %s, %s, now(), now(), '%s');"""%(server_id, server_code, 6, server_ip, port, city_code, province_code, 10, time,bind_port, 1,update_time)
            print(sql)
            node_list.append(sql)

            if sum_node > node_sum:
                break

            sum_node += 1
    # print(node_list)
    # insert_node(node_list)

if __name__ == '__main__':
    code_name = "yle03" # 节点名称
    node_sum = 650 # 节点总数
    time = 630 # 3分钟
    assign_city_node(code_name,node_sum, time)
    print("ok")
    pass
