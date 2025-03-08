#!/usr/bin/env python
# -*- encoding: utf-8 -*-
"""
KPS节点更换
db.order_kps.kps_id

命令行使用(会忽略脚本内配置)
python -u kps_changeip_email_new.py old_ip:new_ip ...

脚本内配置使用
OLD_TO_NEW = {
    '122.114.87.228': '117.50.181.155'
}

更换流程
1> order_kps.filter(kps__ip__in=old_ip_list)
2> order_kps.kps_id_old -> order_kps.kps_id_new
3> 更换一个IP补偿12小时，如果订单内有多个IP，补偿时间=IP量/12，最低为1小时
4> insert into order_kps_allocate_history
5> send_notify
"""

import time
import sys
import re
from notifyutils import send_notify
from notifyconfig import NotifyType

import dbutil
db_db = dbutil.get_db_db()

from redisutil import  redisdb_crs


# 原IP: 新IP
OLD_TO_NEW = {
    # '1.1.1.1': '2.2.2.2',
}


def check_ip_valid(ip):
    ip_check = re.compile('^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$')
    return bool(ip_check.match(ip))


def get_old_ip_info(old_ip_list):
    sql = """
        SELECT ok.id AS ok_id, ok.user_id, po.orderid, kps.ip AS kps_ip, kps.id AS kps_id, kps.code AS kps_code
        FROM order_kps ok
            LEFT JOIN proxy_order po on ok.order_id=po.id
            LEFT JOIN kps on ok.kps_id=kps.id
        WHERE ok.is_valid=1 AND kps.ip in (%s);
    """ % ','.join(repr(i.strip()) for i in old_ip_list)
    cursor = db_db.execute_dictcursor(sql)
    result = cursor.fetchall()
    cursor.close()
    return result


def get_new_ip_info(new_ip_list):
    sql = """
        SELECT kps.ip AS kps_ip, kps.id AS kps_id, kps.code AS kps_code
        FROM kps WHERE status=1 AND is_available=1 AND ip in (%s);
    """ % ','.join(repr(i.strip()) for i in new_ip_list)
    cursor = db_db.execute_dictcursor(sql)
    result = cursor.fetchall()
    cursor.close()
    result = {i['kps_ip']: i for i in result}
    return result

def write_cache(oldcode, newcode):
    keys = "k8s_kps_bak_code_%s" %(oldcode)
    redisdb_crs.set(keys, newcode, ex = 60 * 60 * 24)

def main(old_new_dict):
    old_ip_list = old_new_dict.keys()
    old_ip_info = get_old_ip_info(old_ip_list)
    if not old_ip_info:
        print 'Error: old ip info not found'
        sys.exit(1)

    new_ip_list = old_new_dict.values()
    new_ip_info = get_new_ip_info(old_new_dict.values())
    if len(new_ip_info) != len(new_ip_list):
        invalid_new_ip_list = []
        for i in new_ip_list:
            if i not in new_ip_info:
                invalid_new_ip_list.append(i)
        print 'Error: new ip not found\n\t%s\n' % '\n\t'.join(invalid_new_ip_list)
        sys.exit(1)

    userid_change_ip_info = {}
    sql_change_kps, sql_insert_history = [], []
    for info in old_ip_info:
        ok_id, user_id, orderid, old_kps_ip, old_kps_id, old_kps_code = info['ok_id'], info['user_id'], info['orderid'], info['kps_ip'], info['kps_id'], info['kps_code']
        new_kps_ip = old_new_dict[old_kps_ip]
        new_kps_id, new_kps_code = new_ip_info[new_kps_ip]['kps_id'], new_ip_info[new_kps_ip]['kps_code']
        write_cache(old_kps_code, new_kps_code)
        sql_change_kps.append("""update order_kps set kps_id=%s where id=%s;""" % (new_kps_id, ok_id))
        sql_insert_history.append(
            """insert into order_kps_allocate_history(orderid, order_kps_id, category, old_kps_id, new_kps_id, create_time, memo)
               values('%s', %s, %s, %s, %s, now(), '');
            """ % (orderid, ok_id, 2, old_kps_id, new_kps_id))
        print '[%s] %s(%s) -> %s(%s)' % (orderid, old_kps_code, old_kps_ip, new_kps_code, new_kps_ip)

        userid_change_ip_info.setdefault(user_id, {'orderid_list': [], 'change_info': []})
        userid_change_ip_info[user_id]['orderid_list'].append(orderid)
        userid_change_ip_info[user_id]['change_info'].append('原IP: %s，该IP已经自动更换为: %s；' % (old_kps_ip, new_kps_ip))

    orderid_old_ip_info_list = {}
    for info in old_ip_info:
        orderid_old_ip_info_list.setdefault(info['orderid'], []).append(info)
    sql_date_add = []
    for orderid, old_ip_info_list in orderid_old_ip_info_list.items():
        count = len(old_ip_info_list)
        add_minute = 60 if count > 12 else int((12*60)/count)
        sql_date_add.append("""update proxy_order set end_time=date_add(end_time, interval %s minute) where orderid='%s';""" % (add_minute, orderid))
        print '[%s] add minute %s' % (orderid, add_minute)

    # continue...
    yes = raw_input('\nContinue? [y/n]: ').lower()
    if yes not in ('y', 'yes'):
        sys.exit(1)
    print

    db_db.update(''.join(sql_change_kps))
    db_db.update(''.join(sql_insert_history))
    db_db.update(''.join(sql_date_add))

    for user_id, info in userid_change_ip_info.items():
        send_notify(int(user_id), NotifyType.SERVER_CHANGE_ORDER_KPS, {
            'orderid_encrypt': ','.join([orderid[:-8] + "****" + orderid[-4:] for orderid in info['orderid_list']]),
            'change_info': ''.join(info['change_info']),
        })
        time.sleep(0.2)
        print 'send email to %s %s\n%s' % (user_id, ','.join([orderid[:-8] + "****" + orderid[-4:] for orderid in info['orderid_list']]), ''.join(info['change_info']))


if __name__ == '__main__':
    if len(sys.argv) > 1:
        OLD_TO_NEW = {}
        for old_new in sys.argv[1:]:
            try:
                old, new = old_new.split(':')
            except ValueError:
                print 'Usage: python %s old_ip:new_ip ...' % sys.argv[0]
                sys.exit(1)
            OLD_TO_NEW[old.strip()] = new.strip()

    if not OLD_TO_NEW:
        print 'Error: OLD_TO_NEW must be configured\nCommand Usage: python %s old_ip:new_ip ...' % sys.argv[0]
        sys.exit(1)
    invalid_ip_list = []
    for ip in list(OLD_TO_NEW.keys()) + list(OLD_TO_NEW.values()):
        if not check_ip_valid(ip):
            invalid_ip_list.append(ip)
    if invalid_ip_list:
        print 'Error: invalid ip\n\t%s\n' % '\n\t'.join(invalid_ip_list)
        sys.exit(1)

    main(OLD_TO_NEW)
