#!/usr/bin/python
# -*- coding: UTF-8 -*-
"""
    主要功能: 
        逐个测试JIP节点,验证带宽是否稳定大于10M的节点. 连续测试3天,每天测试3次.如果带宽均大于10M,则输出节点信息.
    前提准备:
        1. 安装pycurl模块
            # Ubuntu/Debain
                apt-get -y install libcurl4-openssl-dev libssl-dev mariadb-server
            # Centos/Redhat:
                yum -y install libcurl-devel

            pip install pycurl
        2. 安装Mariadb-server
            apt-get install  mariadb-server -y
            SET PASSWORD FOR 'root'@'localhost' = PASSWORD('123456'); # 设置root用户密码
            FLUSH PRIVILEGES; # 刷新权限表
            # 创建数据库
            # 创建数据表
            # 插入数据
"""
import sys
reload(sys)
sys.setdefaultencoding('utf8')
import time
import json
from io import BytesIO
import requests
import pycurl
import pymysql
import logging
# 配置日志记录
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    filename='speedtest.log',  # 日志文件名
    filemode='a'  # 以追加模式写入日志文件
)
logger = logging.getLogger(__name__)


class SpeedJipTest:
    def __init__(self):
        self.db, self.cursor = self.connect_mysql() # 声明数据库连接
        self.SpeedUrl = "https://dldir1.qq.com/qqfile/qq/PCQQ9.7.5/QQ9.7.5.28965.exe"
        self.Proxies = "http://jip:Gizajip0301@127.0.0.1:17001"
        self.PublicUrl = "http://ip.plyz.net/ip.ashx?ip=myip"
        self.max_speed = 0 
        # 数据库表
        self.mysql_table_name = "kdl_jip_bandwidth" # 需要插入的数据表名称
        self.jip_server_code = "jips03" # 修改jips服务器的code
 
    def get_notify_method(self):
        url = "https://open.feishu.cn/open-apis/bot/v2/hook/01d1c3ba-35ed-4368-b413-ebff488cedcc"
        headers = {
            "Content-Type": "application/json; charset=utf-8",
        }
        data = {
            "msg_type": "post",
            "content": {
            "post": {
                "zh_cn": {
                    "title": "[%s 服务器带宽测试完毕]"% self.jip_server_code,
                    "content": [
                        [{
                                "tag": "text",
                                "text": "JIP服务器带宽测试完毕, 请上机器查看详细信息或开启下一轮测试"
                        }
                        ]
                    ]
                }
            }
            }
        }
        requests.post(url=url, data=json.dumps(data), headers=headers)

    @staticmethod 
    def connect_mysql():
        db = pymysql.connect(
            host='127.0.0.1', user='root', password='123456', port=3306, db='kdljip_speed'
        )
        cursor = db.cursor()
        return db, cursor

    def insert_to_mysql(self, data):
        try:
            self.db.ping(reconnect=True) 
            """ping方法检测数据库是否正常连接,若连接失败则重新连接"""
        except Exception as e:
            logger.error(e)
            self.cursor.close()
            self.db.close()
            self.db, self.cursor = self.connect_mysql()
        sql = "insert into kdl_jip_bandwidth(proxy_index, public_ip, speed_bandwidth, create_time) values (%s, %s, %s, %s)"
        try:
            self.cursor.execute(sql, data)
            self.db.commit()
        except Exception as e:
            logger.error("[insert_to_mysql] Error, sql: %s, error: %s" % (sql, e))
            self.db.rollback() # 回滚

    def get_jip_proxy_index(self):
        """
            获取jip代理索引
            :return: list 
        """
        url = "http://device.kdlapi.com/listedge?JIPS-CODE={jip_server_code}".format(jip_server_code=self.jip_server_code)
        headers = {
            "API-AUTH": "soj9v7qsfrinm1hu"
        }
        response = requests.get(url, headers=headers)
        
        if response.json()["code"] == 0:
            return response.json()["data"]["nodes"]
        else:
            return None
    
    def progress(self, download_total, downloaded, upload_total, uploaded):
        """
            该函数用于计算并更新下载和上传的进度. 函数参数包括总下载量,已下载量,总上传量,已上传量.
            根据已下载量和开始时间计算当前下载速度，并与历史最大下载速度比较，若当前速度更大，则更新最大下载速度
        """
        if downloaded > 0:
            current_speed = (downloaded / (time.time() - self.start_time))
            if current_speed > self.max_speed:
                self.max_speed = current_speed
    
    def get_jip_bandwidth(self, proxy_index, retries):
        """
            测试JIP节点下载带宽
        """
        headers = [
            "Kdl-Proxy-Index: {proxy_index}".format(proxy_index=proxy_index),
            "kdl-Proxy-Bandwidth: 15"
        ]
        buffer = BytesIO()
        c = pycurl.Curl()
        c.setopt(c.URL, self.SpeedUrl)
        c.setopt(c.WRITEDATA, buffer)
        c.setopt(c.PROXY, self.Proxies)
        c.setopt(c.HTTPHEADER, headers)
        c.setopt(c.NOPROGRESS, False)
        c.setopt(c.XFERINFOFUNCTION, self.progress)
        c.setopt(c.RANGE, '0-10485759')  # 只下载前10MB数据
        c.setopt(c.TIMEOUT, 40)  # 设置超时时间
        c.setopt(c.LOW_SPEED_TIME, 15)  # 低速连接最大持续时间
        c.setopt(c.LOW_SPEED_LIMIT, 1)  # 低速连接的速度阈值
        
        self.start_time = time.time()
        attempt = 0

        while attempt < retries:
            try:
                logger.info("Starting download test, attempt {}".format(attempt + 1))
                c.perform()
                logger.info("Download test completed")
                break  # 成功后退出循环
            except pycurl.error as e:
                logger.error(u"An error occurred: {}".format(e))
                attempt += 1
                if attempt < retries:
                    logger.info("Retrying...")
                else:
                    logger.error("Max retries reached. Download test failed.")
            finally:
                c.close()

        SpeedJip_Bandwidth = "{:.2f}".format(self.max_speed / 1024 / 1024 * 8)
        self.max_speed = 0 
        return SpeedJip_Bandwidth


    def get_detection_public_ip(self):
        """
            检测JIP节点的公网IP是否与节点的IP一致
        """
        jip_proxy_index_list = self.get_jip_proxy_index()
        if jip_proxy_index_list:
            for jip_proxy_index in jip_proxy_index_list:
                public_ip = jip_proxy_index.get('public_ip') # 从接口中获取公网IP
                proxy_index = jip_proxy_index.get('proxy_index') # 从接口中获取代理索引
                headers = {
                    "Kdl-Proxy-Index": "{proxy_index}".format(proxy_index=proxy_index),
                    "kdl-Proxy-Bandwidth": "15"
                }
                proxies = {
                    "http" : self.Proxies,
                    "https" : self.Proxies
                }
                try:
                    response = requests.get(url = self.PublicUrl, headers = headers , proxies = proxies, timeout = 20)
                    if response.status_code == 200:
                        real_public_ip = response.text.split('|')[0]
                        if real_public_ip == public_ip:
                            SpeedJip_Bandwidth = self.get_jip_bandwidth(proxy_index, 3)
                            now_time =  time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
                            data = proxy_index, public_ip, SpeedJip_Bandwidth, now_time
                            self.insert_to_mysql(data)
                            print(proxy_index, public_ip, SpeedJip_Bandwidth)
                        else:
                            logger.error("真实公网IP与接口获取不一致, 不进入测试带宽, 接口获取IP为: {}, 真实公网IP为: {}".format(public_ip, real_public_ip))
                            continue
                    else:
                        logger.error("%s节点请求失败, 状态码为: %s"%(proxy_index, response.status_code))
                        continue
                except Exception as e:
                    logger.error(e)
                    continue
            self.get_notify_method()
        else:
            logger.error("接口异常或无空闲机器")
            return None
        
if __name__ == '__main__':
    speedTest = SpeedJipTest()
    speedTest.get_detection_public_ip()