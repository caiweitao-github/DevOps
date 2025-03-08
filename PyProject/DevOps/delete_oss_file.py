#!/usr/bin/python
# -*- coding: UTF-8 -*-
"""
    代码目的: 删除30天前的oss文件
    运行方式: 计划任务
    运行周期: 每天运行一次
"""
from qcloud_cos import CosConfig, CosS3Client
from qcloud_cos.cos_exception import CosClientError, CosServiceError

import loggerutil
logger = loggerutil.get_logger('delete_oss_file', 'dps_master/delete_oss_file.log')

from datetime import datetime, timedelta
# 获取当前时间
current_time = datetime.now()
# 计算n天前的时间
Days = 31
Days_time = current_time - timedelta(days=Days)
Days_date = datetime(Days_time.year, Days_time.month, Days_time.day)

class Tencent:
    def __init__(self):
        self.id = "AKIDiVEE9uSulaWds6SUMGzfEj6LSPCtV3Xx"
        self.key = "RlzOveLwPiuztgNwPgopfJsIbsCSfoph"
        self.region = "ap-beijing"
        self.scheme = "https"
        self.bucket = 'request-domains-1251449757'
        self.config = CosConfig(Region=self.region, SecretId=self.id, SecretKey=self.key, Scheme=self.scheme)  # 获取配置对象
        self.client = CosS3Client(self.config)  # 根据文件大小自动选择分块大小,多线程并发上传提高上传速度
    
    def get_oss_file(self):
        response = self.client.list_objects(Bucket=self.bucket) # 单次调用 list_objects 接口一次只能查询1000个对象，如需要查询所有的对象，则需要循环调用
        content_list = response.get('Contents')
        for content in content_list:
            key = content.get('Key')
            if key == "log/":
                continue
            else:
                LastModified = content.get('LastModified')
                LastModified_Year = LastModified.split('-')[0]
                LastModified_Month = LastModified.split('-')[1]
                LastModified_Day = LastModified.split('-')[2].split('T')[0]
                LastModified_date = datetime(int(LastModified_Year), int(LastModified_Month), int(LastModified_Day))
                if LastModified_date > Days_date:
                    continue
                else:
                    response = self.client.delete_object(Bucket=self.bucket, Key=key)
                    logger.info("Deleted successfully for Key: %s"%key)

if __name__ == '__main__':
    tencent = Tencent()
    tencent.get_oss_file()