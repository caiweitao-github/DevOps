# -*- encoding: utf-8 -*-
"""
定期检查可用域名数量是否低于阈值
低于阈值则会创建一定数量的域名
"""
import time

import random
import string

from util import AliAPI

import dbutils

import logging
# 配置日志记录
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    filename='/data/log/fps_domain_check.log',  # 日志文件名
    filemode='a'  # 以追加模式写入日志文件
)
logger = logging.getLogger(__name__)

go2proxy_db = dbutils.get_db_go2proxy()

domain_num = 10
create_num = 2


class DomainUtile():
    def __init__(self):
        self.base_domain = '.go2proxy.net'
        self.us_domain_suffix = '.us'
        self.as_domain_suffix = '.as'
        self.master_domain_prefix = '.g'

    def creat_domain_name(self, as_id, us_id):
        count = 0
        while 1:
            if count == create_num:
                break
            domain_list = self.create_domain_str()
            is_ex = self.query_doamin(domain_list)
            if not is_ex:
                us_node = self.get_node(us_id)
                as_node = self.get_node(as_id)
                domain_data = {
                    'domain': [d for d in us_node] + [domain_list[0]],  # 数据格式[fps_id, fps_code, fps_ip, domain]
                    'domain_us': [d for d in us_node] + [domain_list[1]],
                    'domain_as': [d for d in as_node] + [domain_list[2]]
                }
                for domain in domain_data.values():
                    r = AliAPI.create_domain(domain[-1], domain[2])
                    if r:
                        logger.info("create %s ---> %s(%s) success!" % (domain[-1], domain[2], domain[1]))
                    else:
                        logger.error("create error.")
                        return
                count += 1
                self.inser_into_db(domain_data['domain'][-1], domain_data['domain_as'][-1],
                                   domain_data['domain_us'][-1], domain_data['domain_as'][0],
                                   domain_data['domain_us'][0])
            else:
                logger.info("domain %s is exist, skip!" % (domain))
                continue

    def create_domain_str(self):
        charset = string.ascii_lowercase + string.digits
        random_chars = ''.join(random.choice(charset) for _ in range(10))
        return [random_chars + self.master_domain_prefix, random_chars + self.us_domain_suffix,
                random_chars + self.as_domain_suffix]

    def query_doamin(self, domain_str):
        sql = "select * from fps_domain where domain = '%s' and sub_domain_us = '%s' and sub_domain_as = '%s'" %(domain_str[0] + self.base_domain, domain_str[1] + self.base_domain, domain_str[2] + self.base_domain)
        r = go2proxy_db.execute(sql).fetchone()
        return bool(r)

    def inser_into_db(self, domain, as_domain, us_domain, as_id, us_id):
        sql = """insert into fps_domain(domain, sub_domain_as, sub_domain_us, status, update_time, create_time, memo, fps_as_id, fps_us_id) 
        values('%s', '%s', '%s', 1, CURRENT_TIMESTAMP(), CURRENT_TIMESTAMP(), "", %d, %d)""" % (
        domain + self.base_domain, as_domain + self.base_domain, us_domain + self.base_domain, as_id, us_id)
        go2proxy_db.execute(sql)

    def get_node(self, id):
        sql = """select id,code,login_ip from fps where id = '%s' """ % (id)
        r = go2proxy_db.execute(sql).fetchone()
        return r
    
    def check_domain(self):
        sql1 = """select domain,fps_as_id,fps_us_id from fps_domain where status=1"""
        rows1 = go2proxy_db.execute(sql1).fetchall()
        sql2 = """select id,location_code from fps where status in (1,3)"""  # 正常和异常的机器
        rows2 = go2proxy_db.execute(sql2).fetchall()
        as_fps_id_list = []
        us_fps_id_list = []
        for fps_id, location_code in rows2:
            if location_code == 'as':
                as_fps_id_list.append(fps_id)
            elif location_code == 'us':
                us_fps_id_list.append(fps_id)

        # 机器两两组合
        machine_pairs = {}
        for as_fps_id in as_fps_id_list:
            for us_fps_id in us_fps_id_list:
                machine_pairs[(as_fps_id, us_fps_id)] = []

        # 现有域名加入map
        for domain, as_fps_id, us_fps_id in rows1:
            pair_key = (as_fps_id, us_fps_id)
            if pair_key in machine_pairs:
                machine_pairs[pair_key].append(domain)

        # 确保每对机器有至少n个域名
        for pair_key, domain_list in machine_pairs.items():
            if len(domain_list) < domain_num:
                # 为这两台机器创建域名
                as_fps_id = pair_key[0]
                us_fps_id = pair_key[1]
                self.creat_domain_name(as_fps_id, us_fps_id)


if __name__ == '__main__':
    while True:
        try:
            logger.info('check domain start')
            D = DomainUtile()
            D.check_domain()
        except Exception as exc:
            logger.error("error: %s" % exc)
        time.sleep(3 * 60)
