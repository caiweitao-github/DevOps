#!/usr/bin/env python
# coding: utf-8

"""
@File  :   drop_ck_partition.py
@Time  :   2020/09/18 11:08:21
@Author:   aladdinding
@Desc  :   删除clickhouse分区

@RunHostName :   None
@RunMode     :   None
@RunPeriod   :   None
"""
import ckutil
import datetime
ck = ckutil.get_ck_apistat()


def drop_partition(partition):
    sql = "ALTER TABLE api_request_history DETACH PARTITION %s" % partition
    ck.client.execute(sql)


class CkClient(object):
    @staticmethod
    def get_ck_client(databases):
        if databases == "nodeops":
            return ckutil.get_ck_nodeops()
        elif databases == "tpsstat":
            return ckutil.get_ck_tpsstat()
        elif databases == "apistat":
            return ckutil.get_ck_apistat()
        else:
            raise Exception


def partition_generater(startdate, enddate):
    startdate = datetime.datetime.strptime(startdate, "%Y%m%d")
    enddate = datetime.datetime.strptime(enddate, "%Y%m%d")
    while startdate <= enddate:
        yield datetime.datetime.strftime(startdate, "%Y%m%d")
        startdate = startdate + datetime.timedelta(days=1)


def main():
    """
    Steps:
        1. 
        2. 
        3. 
    """
    partition_gen = partition_generater("20200218", "20200518")
    for partition in partition_gen:
        drop_partition(partition)


if __name__ == "__main__":
    main()
