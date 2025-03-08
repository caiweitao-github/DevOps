# -*- coding=utf-8
import logging
import sys
import json
import os

from qcloud_cos import CosConfig, CosServiceError
from qcloud_cos import CosS3Client
from qcloud_cos.cos_threadpool import SimpleThreadPool

logging.basicConfig(level=logging.INFO, stream=sys.stdout)

secret_id = ''
secret_key = ''   
region = 'ap-beijing'
                      

config = CosConfig(Region=region, SecretId=secret_id, SecretKey=secret_key)
client = CosS3Client(config)

bucket = 'cos名称'
data_dir_base = 'data4/clickhouse/data'

data_dir_list = ['']
# 使用默认的空分隔符可以列出目录下面的所有子节点，实现类似本地目录递归的效果,
# 如果 delimiter 设置为 "/"，则需要在程序里递归处理子目录
delimiter = ''



def listCurrentDir(path_dir, data_time):
    time = data_time
    response = client.list_objects(Bucket='cos名称', Prefix=path_dir + "/", Delimiter='/')
    if 'CommonPrefixes' in response:
        download_list = []
        check_list = []
        for folder in response['CommonPrefixes']:
            check_list.append(folder['Prefix'])
        for d in check_list:
            if not time:
                response = client.list_objects(Bucket='cos名称', Prefix=d)
            else:
                response = client.list_objects(Bucket='cos名称', Prefix=d + "%s" %(time))
                
            if 'Contents' in response:
                for content in response['Contents']:
                    download_list.append(content['Key'])

    return download_list


# 下载文件到本地目录，如果本地目录已经有同名文件则会被覆盖；
# 如果目录结构不存在，则会创建和对象存储一样的目录结构
def downLoadFiles(file_infos):
    localDir = "/"

    pool = SimpleThreadPool()
    for file in file_infos:
        localName = localDir + file

        if not os.path.exists(os.path.dirname(localName)):
            os.makedirs(os.path.dirname(localName))

        if str(localName).endswith("/"):
            continue

        pool.add_task(client.download_file, bucket, file, localName)



    pool.wait_completion()
    return None


def downLoadDirFromCos(prefix, data_time=None):
    global file_infos

    try:
        file_infos = listCurrentDir(prefix, data_time)

    except CosServiceError as e:
        print(e.get_origin_msg())
        print(e.get_digest_msg())
        print(e.get_status_code())
        print(e.get_error_code())
        print(e.get_error_msg())
        print(e.get_resource_location())
        print(e.get_trace_id())
        print(e.get_request_id())

    downLoadFiles(file_infos)
    return None


if __name__ == "__main__":
    # for d in data_dir_list:
    #     download_dir = os.path.join(data_dir_base, d)
    #     dir_path = download_dir.strip("/")
    #     downLoadDirFromCos(dir_path)
    pass
