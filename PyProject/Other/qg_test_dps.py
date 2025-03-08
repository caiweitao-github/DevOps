#!/usr/bin/env python
# -*- encoding: utf-8 -*-

from gevent import monkey; monkey.patch_all()
import gevent
import requests
from uautil import get_ua
import dbutil
import log

db = dbutil.get_db_qingguo()

username = "E8424BEF"
password = "B9FBA12721C5"

tunnel = "tunnel6.qg.net:15226"

status_dict = {'城域网': 1, '数据中心': 2, '': 3}

class Check_ip():
    def __init__(self):
        self.username = ""
        self.password = ""
        self.tunnel = ""
        self.proxies = {
            "http": "http://%(user)s:%(pwd)s@%(proxy)s/" % {"user": self.username, "pwd": self.password, "proxy": self.tunnel},
            "https": "http://%(user)s:%(pwd)s@%(proxy)s/" % {"user": self.username, "pwd": self.password, "proxy": self.tunnel}
        }
        
    def checkip(self):
        url = 'https://ip.useragentinfo.com/json'
        try:
            headers = {"User-Agent": get_ua()}
            response = requests.get(url=url, headers=headers, proxies=self.proxies, timeout=5)
            up, down = self.test_proxy_bandwidth()
            if response:
                self.write_to_db(response.json()['ip'], status_dict.get(response.json()['net'], 4), response.json()['isp'], up, down, response.json()['city'])
        except Exception as e:
            log.error(e)
       
    @staticmethod 
    def write_to_db(ip, status, isp, up, down, location):
        sql = "insert into ip_attribution(ip, status, carrier, upload, download, location, create_time) values('%s', '%s', '%s', '%s', '%s', '%s', now())" %(ip, status, isp, up, down, location)
        db.execute(sql)
        log.info("write db success!")

if __name__ == '__main__':
    ip = Check_ip()
    while 1:
        try:
            gevent.joinall([gevent.spawn(ip.checkip) for _ in range(20)])
            gevent.sleep(0.5)
        except Exception as e:
            log.error(e)