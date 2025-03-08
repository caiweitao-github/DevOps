#!/usr/bin/python
# -*- coding: UTF-8 -*-
"""
    代码目的: 获取IME/EC设备请求的域名,jip_node_index以及出口IP统计
"""
import random
import zipfile
import datetime
from datetime import datetime, timedelta
import tldextract
import time
import os

import dbutil
jip_node = dbutil.get_db_kdljip()
import ckutil
nodeops = ckutil.get_ck_nodeops()

import loggerutil
logger = loggerutil.get_logger("get_ime_ec_request_domain", "dps_master/get_ime_ec_request_domain.log")

from redisutil import redisdb_cq
REDIS_EDGE_LOG_KEY = "edge:log:{0}:{1}"
Expire_time = 30 * 24 * 60 * 60 # 设置超时时间

from qcloud_cos import CosConfig, CosS3Client
from qcloud_cos.cos_exception import CosClientError, CosServiceError
Tencent_SecretID = "AKIDiVEE9uSulaWds6SUMGzfEj6LSPCtV3Xx"
Tencent_SecretKey = "RlzOveLwPiuztgNwPgopfJsIbsCSfoph"
Tencent_Region = "ap-beijing"
Tencent_Scheme = "https"
Tencent_Bucket = 'request-domains-1251449757'

Length = 15 # 定义随机数15位

# # 获取当前日期
current_date = time.strftime("%Y-%m-%d", time.localtime())
# 获取当前时间
current_time = datetime.now()
# 计算今天的时间
current_date_today = current_time.strftime("%Y-%m-%d 00:00:00")
# 计算前一天的日期
previous_day = current_time - timedelta(days=1)
# 计算前一天格式化时间:
current_date_yesterday = previous_day.strftime("%Y-%m-%d 00:00:00")
# 计算前10分钟的日期
# date_time_minus_ten_minutes = current_time - timedelta(minutes=10)
# 初始化时间间隔列表
time_intervals = []
# 遍历前一天的每个小时

for i in range(24):
    start_time = previous_day.replace(hour=i, minute=0, second=0, microsecond=0)
    end_time = previous_day.replace(hour=i, minute=59, second=59, microsecond=999999)
    # 将时间间隔添加到列表中
    time_intervals.append((start_time, end_time))

Provider_List = ["ec", "ime"]

def get_deviceid_jip_index(device_id, provider):
    """
        根据deviceid获取jip_index
    """
    sql = "select provider_id from jip_node where proxy_index = '{device_id}' and provider= '{provider}';".format(device_id=device_id, provider=provider)
    result = jip_node.execute(sql).fetchone()
    if result:
        provider_id = result[0]
        return provider_id
    return None

def get_deviceid_domain(provider):
    device_domains = {}  # 初始化一个空字典来存储设备 ID 和对应的域名集合
    """
        获取deviceid 请求过的 domain
    """
    for interval in time_intervals:
        start_time = interval[0].strftime("%Y-%m-%d %H:%M:%S")
        end_time = interval[1].strftime("%Y-%m-%d %H:%M:%S")
        if provider == "ime":
            sql = """
                select device_id,ip,domain,method from node_request_history
                where request_time >= '{start_time}' and request_time <= '{end_time}' and provider = '{provider}' and device_id<>''
                group by device_id,ip,domain,method;
            """.format(start_time=start_time, end_time=end_time, provider=provider)
            result = nodeops.execute(sql)
            if result:
                for row in result:
                    device_id = row[0]
                    ip = row[1]
                    domain = tldextract.extract(row[2]).domain
                    if tldextract.extract(row[2]).suffix == "gov.cn":
                        domain = "gov.cn"
                    if tldextract.extract(row[2]).suffix:
                        domain = tldextract.extract(row[2]).domain + "." + tldextract.extract(row[2]).suffix
                    device_key = device_id + ':' + ip
                    if device_key not in device_domains:
                        device_domains[device_key] = set()
                    # 将域名添加到设备 ID 对应的集合中
                    device_domains[device_key].add(domain)
            else:
                logger.debug("{0} query result is null".format(provider))
        

        if provider == "ec":
            sql = """
                select device_id,ip,domain,method from node_request_history
                where request_time >= '{start_time}' and request_time <= '{end_time}' and provider = '{provider}' and device_id<>''
                group by device_id,ip,domain,method;
            """.format(start_time=start_time, end_time=end_time, provider=provider)
            result = nodeops.execute(sql)
            if result:
                for row in result:
                    device_id = row[0]
                    ip = row[1]
                    domain = tldextract.extract(row[2]).domain
                    if tldextract.extract(row[2]).suffix == "gov.cn":
                        domain = "gov.cn"
                    if tldextract.extract(row[2]).suffix:
                        domain = tldextract.extract(row[2]).domain + "." + tldextract.extract(row[2]).suffix
                    method = row[3]
                    if method == "GET" or method == "POST":
                        protocol = "http"
                    elif method == "CONNECT":
                        protocol = "https"
                    elif method == "socks" or method == "1":
                        protocol = "socks5"
                    domain_str = "{0} {1}".format(domain , protocol)
                    device_key = device_id + ':' + ip
                    if device_key not in device_domains:
                        device_domains[device_key] = set()
                    # 将域名添加到设备 ID 对应的集合中
                    device_domains[device_key].add(domain_str)

            else:
                logger.debug("{0} query result is null".format(provider))
                
    return device_domains

def get_file_zip_file(filename):
    """
        将文件进行打包,生成zip文件后存入OSS对象存储中并将其进行Redis缓存
    """
    config = CosConfig(Region=Tencent_Region, SecretId=Tencent_SecretID, SecretKey=Tencent_SecretKey, Scheme=Tencent_Scheme)  # 获取配置对象
    client = CosS3Client(config)  # 根据文件大小自动选择分块大小,多线程并发上传提高上传速度
    provider = filename.split('_')[0]
    date_time =  previous_day.strftime("%Y%m%d")
    random_int_num = random.randint(10**(Length-1), (10**Length)-1)
    output_filename = '{0}.zip'.format(random_int_num)
    with zipfile.ZipFile(output_filename,'w') as zf:
        zf.write(filename)
    Upload_File_Path = "log/{0}".format(output_filename)
    # 必须以二进制的方式打开文件
    # 填写本地文件的完整路径.如果未指定本地路径，则默认从示例程序所属项目对应本地路径中上传文件

    # # 使用高级接口断点续传, 失败重试时不会上传已成功的分块(这里重试10次)
    for i in range(0,10):
        try:
            response = client.upload_file(
                Bucket = Tencent_Bucket, # string -> 存储桶名称
                Key = Upload_File_Path, # string -> 分块上传路径名
                LocalFilePath = '{0}'.format(output_filename), # string -> 本地文件路径名
                PartSize= 3, # int -> 分块的大小设置, 单位为MB
                MAXThread= 10, # int -> 并发上传的最大线程数
                EnableMD5= False, # bool -> 是否打开MD5校验
                # kwargs = {} # dict -> 设置请求headers
                # return {} # dict -> 成功上传文件的元信息
                               )
            logger.info(response)
            file_download_url = "https://{0}.cos.{1}.myqcloud.com/{2}".format(Tencent_Bucket, Tencent_Region, Upload_File_Path)
            logger.info("文件上传成功, 文件下载路径为: {0}".format(file_download_url))
            redisdb_cq.set(REDIS_EDGE_LOG_KEY.format(provider, date_time), file_download_url , Expire_time)

        except CosClientError or CosServiceError as e:
            logger.error(e)
            logger.error("文件上传错误, 异常信息为: %s", e)
    os.remove(output_filename)
    logger.info("{0} 文件已删除!".format(output_filename))

def main(provider):
    date_time = previous_day.strftime("%Y_%m_%d")
    file_name = "{provider}_{date_time}_request_domain.txt".format(provider=provider, date_time=date_time)
    device_domains = get_deviceid_domain(provider)

    with open(file_name, "w") as file:
        for key, values in device_domains.items():
            provider_index = key.split(':')[0]
            deviceid_export_ip  = key.split(':')[1]
            provider_id = get_deviceid_jip_index(provider_index, provider)
            domain_list = list(values)
            # 将内容格式化为字符串
            content = "{0} | {1} | {2} | {3}\n".format(provider_id, previous_day.strftime("%Y-%m-%d"), deviceid_export_ip, domain_list)
            # 将内容写入文件
            file.write(content)

    get_file_zip_file(file_name)
    os.remove(file_name)
    logger.info("{0} 文件已删除!".format(file_name))

if __name__ == '__main__':
    for provider in Provider_List:
        main(provider)
