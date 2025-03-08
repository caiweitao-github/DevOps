# -*- coding: utf-8 -*-

from datetime import date, timedelta
import ckutil
import dbutil

nodeops_db = dbutil.get_db_nodeops()
kdljip_db = dbutil.get_db_kdljip()

ck_nodeops = ckutil.get_ck_nodeops()

day = (date.today()).strftime("%Y-%m-%d %H:%M:%S")

yesterday = (date.today() + timedelta(-1)).strftime("%Y-%m-%d %H:%M:%S")

def get_third_party_ip_data(source):
    sql = """select distinct(public_ip) from external_changeip_history where create_time between '%s' and '%s' and provider = '%s'""" % (
        yesterday, day, source)
    rows = ck_nodeops.execute(sql)
    return { ip[0] for ip in rows }


def get_dps_ip_product_count():
    sql = """select count(distinct(ip)) as debug_count from dps_changeip_history where change_time>='%s'and change_time<='%s' and is_valid=1 and is_expected=1""" % (
        yesterday, day)
    cur = nodeops_db.execute(sql)
    row = cur.fetchone()
    return int(row[0])


def get_jde_ip_product_count():
    sql = """select count(distinct(ip)) as debug_count from jde_changeip_history where change_time>='%s'and change_time<='%s' and is_expected=1""" % (
        yesterday, day)
    cur = nodeops_db.execute(sql)
    row = cur.fetchone()
    return int(row[0])

def get_jip_ip_product_count():
    sql = """select count(distinct(edge_public_ip)) from jip_node where status = 1"""
    cur = kdljip_db.execute(sql)
    row = cur.fetchone()
    return int(row[0])


if __name__ == '__main__':
    dps_ip_count = get_dps_ip_product_count()
    jde_ip_count = get_jde_ip_product_count()
    jip_ip_count = get_jip_ip_product_count()
    xx_ip_data = get_third_party_ip_data('xiaoxiong')
    yle_ip_data = get_third_party_ip_data('yle')
    nt_ip_data = get_third_party_ip_data('nt')
    count_num = dps_ip_count + jde_ip_count + jip_ip_count + len(xx_ip_data) + len(yle_ip_data) + len(nt_ip_data)
    print("IP总数: %d" %(count_num))
    print("NT与YL的重复IP数量: %d" %(len(nt_ip_data & yle_ip_data)))  #简单处理，取交集
    print("YL与XX的重复IP数量: %d" %(len(yle_ip_data & xx_ip_data)))
    print("NT与XX的重复IP数量: %d" %(len(nt_ip_data & xx_ip_data)))