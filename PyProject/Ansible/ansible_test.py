#!/usr/bin/env python3

import os
import json
import dbutil
import argparse

dbdb = dbutil.get_db_db()

inventory = {"_meta": {"hostvars": {}},"dps": {"hosts": [],"vars": {}}}

def build_inventory(data):
    if data.startswith('dps'):
        sql = "select code,login_ip,login_port from dps where code = '%s'" %(data)
    else:
        sql = "select dps.code,dps.login_ip,dps.login_port from dps,dps_group_relation,dpsgroup where dps_group_relation.group_id = dpsgroup.id and dps_group_relation.dps_id = dps.id and dpsgroup.code = '%s'" %(data)
    cursor = dbdb.execute_dictcursor(sql)
    res = cursor.fetchall()
    _ = list(map(lambda host : (inventory['_meta']['hostvars'].update({host['code']: {'ansible_host': host['login_ip'], 'ansible_port': host['login_port']}}), inventory['dps']['hosts'].append(host['code'])), res))
    

if __name__ == '__main__':
    if os.path.exists('code'):
        parser = argparse.ArgumentParser()
        parser.add_argument('--list', action = 'store_true')
        parser.add_argument('--host', action = 'store_true')
        args = parser.parse_args()
        with open('code') as f:
            for line in f:
                build_inventory(line.strip())
        if args.list:
            print(json.dumps(inventory, indent=4))
        elif args.host:
            print(json.dumps(inventory['dps'], indent=4))
    else:
        print("not found code file!")