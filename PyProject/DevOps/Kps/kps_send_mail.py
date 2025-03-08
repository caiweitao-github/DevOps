#!/usr/bin/env python
# -*- encoding: utf-8 -*-

import time
import sys
import re
from notifyutils import send_notify
from notifyconfig import NotifyType

import dbutil
db_db = dbutil.get_db_db()


# 原IP: 新IP
OLD_TO_NEW = {
    # '1.1.1.1': '2.2.2.2',
}


def check_ip_valid(ip):
    ip_check = re.compile('^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$')
    return bool(ip_check.match(ip))


def get_old_ip_info(old_ip_list):
    sql = """
        SELECT ok.user_id, po.orderid, kps.ip AS kps_ip, kps.code AS kps_code
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
        SELECT kps.ip AS kps_ip, kps.code AS kps_code
        FROM kps WHERE status=1 AND is_available=1 AND ip in (%s);
    """ % ','.join(repr(i.strip()) for i in new_ip_list)
    cursor = db_db.execute_dictcursor(sql)
    result = cursor.fetchall()
    cursor.close()
    result = {i['kps_ip']: i for i in result}
    return result


def main(old_new_dict):
    old_ip_list = old_new_dict.keys()
    old_ip_info = get_old_ip_info(old_ip_list)
    if not old_ip_info:
        print('Error: old ip info not found')
        sys.exit(1)

    new_ip_list = old_new_dict.values()
    new_ip_info = get_new_ip_info(old_new_dict.values())
    if len(new_ip_info) != len(new_ip_list):
        invalid_new_ip_list = []
        for i in new_ip_list:
            if i not in new_ip_info:
                invalid_new_ip_list.append(i)
        print('Error: new ip not found\n\t%s\n') % '\n\t'.join(invalid_new_ip_list)
        sys.exit(1)

    userid_change_ip_info = {}
    for info in old_ip_info:
        user_id, orderid, old_kps_ip, old_kps_code = info['user_id'], info['orderid'], info['kps_ip'], info['kps_code']
        new_kps_ip = old_new_dict[old_kps_ip]
        new_kps_code =  new_ip_info[new_kps_ip]['kps_code']
        print('[%s] %s(%s) -> %s(%s)') % (orderid, old_kps_code, old_kps_ip, new_kps_code, new_kps_ip)

        userid_change_ip_info.setdefault(user_id, {'orderid_list': [], 'change_info': []})
        userid_change_ip_info[user_id]['orderid_list'].append(orderid)
        userid_change_ip_info[user_id]['change_info'].append('原IP: %s，需要在24小时内进行更换为新IP：%s；' % (old_kps_ip, new_kps_ip))

    yes = raw_input('\nContinue? [y/n]: ').lower()
    if yes not in ('y', 'yes'):
        sys.exit(1)
    print

    for user_id, info in userid_change_ip_info.items():
        send_notify(int(user_id), NotifyType.KPS_CHANGE_ORDER, {
            'orderid_encrypt': ','.join([orderid[:-8] + "****" + orderid[-4:] for orderid in info['orderid_list']]),
            'change_info': ''.join(info['change_info']),
        })
        time.sleep(0.2)
        print('send email to %s %s\n%s') % (user_id, ','.join([orderid[:-8] + "****" + orderid[-4:] for orderid in info['orderid_list']]), ''.join(info['change_info']))


if __name__ == '__main__':
    if len(sys.argv) > 1:
        OLD_TO_NEW = {}
        for old_new in sys.argv[1:]:
            try:
                old, new = old_new.split(':')
            except ValueError:
                print('Usage: python %s old_ip:new_ip ...') % sys.argv[0]
                sys.exit(1)
            OLD_TO_NEW[old.strip()] = new.strip()

    if not OLD_TO_NEW:
        print('Error: OLD_TO_NEW must be configured\nCommand Usage: python %s old_ip:new_ip ...') % sys.argv[0]
        sys.exit(1)
    invalid_ip_list = []
    for ip in list(OLD_TO_NEW.keys()) + list(OLD_TO_NEW.values()):
        if not check_ip_valid(ip):
            invalid_ip_list.append(ip)
    if invalid_ip_list:
        print('Error: invalid ip\n\t%s\n') % '\n\t'.join(invalid_ip_list)
        sys.exit(1)

    main(OLD_TO_NEW)
