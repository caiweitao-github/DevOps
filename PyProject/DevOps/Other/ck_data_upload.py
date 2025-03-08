# -*- coding=utf-8

import re
import os
import sys
import time
import logging
import datetime
import requests
from qcloud_cos import CosConfig
from qcloud_cos import CosS3Client
from qcloud_cos import CosServiceError
from qcloud_cos import CosClientError
from qcloud_cos.cos_threadpool import SimpleThreadPool

logging.basicConfig(level=logging.ERROR, stream=sys.stdout)

secret_id = ''     
secret_key = ''   
region = 'ap-beijing'      


config = CosConfig(Region=region, SecretId=secret_id, SecretKey=secret_key) 
client = CosS3Client(config)

data_dir_base = '/data4/clickhouse/data'

data_dir_list = ['']

metadata_data_base = '/data/clickhouse'
metadata_data_list = ['metadata',]

exclude = [
    ''
]

def send_msg(context):
    date_time = time.strftime("%H:%M:%S", time.localtime())
    data = {
	    "msg_type": "post",
	    "content": {
		    "post": {
			    "zh_cn": {
				    "title": "【clickhouse】  [kdlmain13]  %s" %(date_time),
				    "content": [
					    [{
                            "tag": "text",
                            "text": context
						},
                        {
                            "tag": "at",
	                        "user_id": "5c852geg",
	                        "user_name": "weitaocai"
                        }
					    ]
				    ]
			   }
		    }
	    }
    }
    headers = {"Content-Type": "application/json"}
    url_token = ""
    requests.post(url_token, json=data, headers=headers)

def update_all_data():
    today = datetime.datetime.now()
    time = today.strftime('%Y%m%d')
    for dirs in data_dir_list:
        uploadDir = os.path.join(data_dir_base, dirs)
        bucket = 'cos地址'
        g = os.walk(uploadDir, followlinks=True)
        pool = SimpleThreadPool()
        for path, dir_list, file_list in g:
            ret = re.compile(".*%s.*" %(time))
            dir_list[:] = [d for d in dir_list if not ret.search(d)]
            for file_name in file_list:
                srcKey = os.path.join(path, file_name)
                cosObjectKey = srcKey.strip('/')
                exists = False
                try:
                    response = client.head_object(Bucket=bucket, Key=cosObjectKey)
                    exists = True
                except CosServiceError as e:
                    if e.get_status_code() == 404:
                        exists = False
                    else:
                        print("Error happened, reupload it.")
                if not exists:
                    print("File %s not exists in cos, upload it", srcKey)
                    pool.add_task(client.upload_file, bucket, cosObjectKey, srcKey)


        pool.wait_completion()
        result = pool.get_result()
        if not result['success_all']:
            print("Not all files upload sucessed. you should retry")

def update_metadata():
    for dirs in metadata_data_list:
        uploadDir = os.path.join(metadata_data_base, dirs)
        bucket = 'cos地址'
        g = os.walk(uploadDir, followlinks=True)
        pool = SimpleThreadPool()
        for path, dir_list, file_list in g:
            for file_name in file_list:
                srcKey = os.path.join(path, file_name)
                cosObjectKey = srcKey.strip('/')
                exists = False
                try:
                    response = client.head_object(Bucket=bucket, Key=cosObjectKey)
                    exists = True
                except CosServiceError as e:
                    if e.get_status_code() == 404:
                        exists = False
                    else:
                        print("Error happened, reupload it.")
                if not exists:
                    print("File %s not exists in cos, upload it", srcKey)
                    pool.add_task(client.upload_file, bucket, cosObjectKey, srcKey)


        pool.wait_completion()
        result = pool.get_result()
        if not result['success_all']:
            print("Not all files upload sucessed. you should retry")

def get_file_path(root_path):
    today = datetime.datetime.now()
    offset = datetime.timedelta(days=-1)
    time = (today + offset).strftime('%Y%m%d')
    bucket = 'cos地址'
    ret = re.compile("%s.*" %(time))
    dir_or_files = os.listdir(root_path)
    pool = SimpleThreadPool()
    for dir_file in dir_or_files:
        file_list = []
        dir_file_path = os.path.join(root_path,dir_file)
        if os.path.isdir(dir_file_path):
            get_file_path(dir_file_path)
        elif os.path.isfile(dir_file_path):
            x = dir_file_path.split("/")
            if ret.search(x[-2]):
                file_list.append(dir_file_path)
        for file in file_list:
            srcKey = file
            cosObjectKey = srcKey.strip('/')
            exists = False
            try:
                response = client.head_object(Bucket=bucket, Key=cosObjectKey)
                exists = True
            except CosServiceError as e:
                if e.get_status_code() == 404:
                    exists = False
                else:
                    print("Error happened, reupload it.")
            if not exists:
                print("File %s not exists in cos, upload it" %(srcKey))
                pool.add_task(client.upload_file, bucket, cosObjectKey, srcKey, StorageClass='ARCHIVE')
    pool.wait_completion()
    result = pool.get_result()
    if not result['success_all']:
        print("Not all files upload sucessed. you should retry")

def check_upload_is_success():
    today = datetime.datetime.now()
    offset = datetime.timedelta(days=-1)
    time = (today + offset).strftime('%Y%m%d')
    for check_dir in data_dir_list:
        dir_path = os.path.join(data_dir_base, check_dir)
        dir_path = dir_path.strip("/")
        response = client.list_objects(Bucket='cos地址', Prefix=dir_path + "/", Delimiter='/')
        if 'CommonPrefixes' in response:
            check_list = []
            for folder in response['CommonPrefixes']:
                check_list.append(folder['Prefix'])
            for d in check_list:
                response = client.list_objects(Bucket='cos地址', Prefix=d + "%s" %(time))
                x = d.rstrip("/")
                if 'Contents' not in response and os.path.basename(x) not in exclude:
                    send_msg("%s 上传 %s 数据失败, 或者数据目录没有当天的数据!" %(d, time))

if __name__ == '__main__':
    update_metadata()
    for dir in data_dir_list:
        root_path = os.path.join(data_dir_base, dir)
        get_file_path(root_path)
    check_upload_is_success()
