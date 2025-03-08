#!/usr/bin/env python
# -*- coding: utf-8 -*-
from dbutil import *
import staffnotify
import logging, os, sys

logger = logging.getLogger()
handler = logging.FileHandler('/data/kdl/log/devops/mysql_ha.log')
logger.setLevel(logging.INFO)
formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
handler.setFormatter(formatter)
logger.addHandler(handler)


def Slave():
    try:
        sql = "show slave status"
        db = DB(user='db_r', password='db_r', dbname='db', host='10.0.3.17', port=3306, autocommit=True)
        conn = db.connect()
        cursor = db.execute_dictcursor(sql)
        result = cursor.fetchall()
        db.close()
        IO = result[0]["Slave_IO_Running"]
        SQL = result[0]["Slave_SQL_Running"]
        IO_Log = result[0]["Last_IO_Error"]
        SQL_Log = result[0]["Last_SQL_Error"]
    except Exception as e:
        logging.error(e)
        logger.removeHandler(handler)
        return False
    else:
        dirc = sys.argv[0]
        logging.info('run %s...' %(os.path.basename(dirc)))
        logger.removeHandler(handler)
        return IO, SQL, IO_Log, SQL_Log



if __name__ == '__main__':
    if Slave():
        IO, SQL, IO_Log, SQL_Log = Slave()
        if IO != 'Yes' or SQL != 'Yes':
            staffnotify.notify_feishu2({'text': "IO_Status: %s\nSQL_Status: %s\nIO_log: %s\nSQL_log: %s" % (IO, SQL, IO_Log, SQL_Log)}, 'mysql_ha',tag='数据库告警')