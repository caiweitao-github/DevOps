#!/usr/bin/env python
# -*- encoding: utf-8 -*-

import datetime
import dbutil
import time
import log
dbnodeops = dbutil.get_db_nodeops()

TABLET_LIST = ['external_changeip_history']

today = datetime.datetime.now()
offset_3 = datetime.timedelta(days=-3)
three_day_ago = (today + offset_3).strftime('%Y-%m-%d 00:00:00')

delete_num = 5000

def get_data(table_name):
    sql = "select min(id),max(id) from %s where change_time < '%s'" % (table_name, three_day_ago)
    data = dbnodeops.execute(sql).fetchone()
    return all((data)) and data or False

def delete_data(table_name):
    lable = 1
    data = get_data(table_name)
    if data and int(data[0]) < int(data[1]):
        min_id, max_id = int(data[0]), int(data[1])
        while lable:
            min_id = min_id + delete_num
            if min_id > max_id:
                min_id = max_id
                lable = 0
            sql = "delete from %s where id <= %s" % (table_name, min_id)
            dbnodeops.execute(sql)
            time.sleep(1)
        return False
    else:
        log.error('%s get data error' %(table_name))
        return False

if __name__ == '__main__':
    try:
        for i in TABLET_LIST:
            start_time = time.time()
            log.info('start delete %s ...' %(i))
            while 1:
                res = delete_data(i)
                if not res:
                    end_time = time.time()
                    time_user = end_time - start_time
                    log.info('delete %s done, time use: %s' %(i, time_user))
                    break
        dbnodeops.close()
    except Exception as e:
        dbnodeops.close()
        log.error(e)