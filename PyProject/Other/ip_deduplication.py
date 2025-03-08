# -*- coding: utf-8 -*-

from datetime import date, timedelta
import ckutil
import dbutil

nodeops_db = dbutil.get_db_nodeops()

kdljip_db = dbutil.get_db_kdljip()

ck_nodeops = ckutil.get_ck_nodeops()
ck_jde = ckutil.get_ck_eipstat()


day = '2024-10-04 00:00:00'

yesterday = '2024-10-03 00:00:00'

def get_third_party_ip_data(source):
    sql = """select distinct(public_ip) from external_changeip_history where create_time between '%s' and '%s' and provider = '%s'""" % (
        yesterday, day, source)
    rows = ck_nodeops.execute(sql)
    return { ip[0] for ip in rows }


def get_dps_ip_product_count():
    sql = """select distinct(ip) from dps_changeip_history where change_time>='%s'and change_time<='%s' and is_valid=1 and is_expected=1""" % (
        yesterday, day)
    cur = nodeops_db.execute(sql)
    rows = cur.fetchall()
    return { ip[0] for ip in rows }


def get_jde_ip_product_count():
    sql = """select distinct(ip) from jde_changeip_history where change_time>='%s'and change_time<='%s' and is_expected=1""" % (
        yesterday, day)
    cur = nodeops_db.execute(sql)
    rows = cur.fetchall()
    return { ip[0] for ip in rows }

def get_jip_ip_product_count():
    sql = """select distinct(edge_public_ip) from jip_node where status = 1"""
    cur = kdljip_db.execute(sql)
    rows = cur.fetchall()
    return { ip[0] for ip in rows }


if __name__ == '__main__':
    dps_ip_count = get_dps_ip_product_count()
    jde_ip_count = get_jde_ip_product_count()
    jip_ip_count = get_jip_ip_product_count()
    xx_ip_data = get_third_party_ip_data('xiaoxiong')
    yle_ip_data = get_third_party_ip_data('yle')
    nt_ip_data = get_third_party_ip_data('nt')
    count_num = len(dps_ip_count | jde_ip_count | jip_ip_count | xx_ip_data | yle_ip_data | nt_ip_data)
    print("IP总数: %d" %(count_num))