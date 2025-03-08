#!/usr/bin/python
# -*- coding: UTF-8 -*-
"""
    功能目的:扫描天互数据百度云机房上订单访问腾讯类域名失败的有效订单
    运行方式:计划任务,每30分钟执行一次
    启动命令:python -u /data/test/tps_dtect_tencent_domain.py > /dev/null 2>&1 &
"""
import sys # 导入 sys 模块，用于配置默认字符编码
reload(sys) # 导入 sys 模块，用于配置默认字符编码
sys.setdefaultencoding("utf8")  # 设置默认字符编码为 UTF-8，以确保处理 Unicode 数据时不会出现编码问题
import urllib2
import subprocess
import json
import time
import re
Logger_File = '/data/log/tps/proxy.log'
Limit_Num = 80
def get_tps_proxy_log():
    start_time = time.strftime("%H:%M:%S", time.localtime(time.time()-30*60))
    end_time = time.strftime("%H:%M:%S", time.localtime())
    command = "awk -v start='{0}' -v end='{1}' '$2>=start && $2<=end' {2} | grep '|' | grep -E '*qq.com' | grep 'operation was canceled' | awk -F '|' '{{print $4}}' | sort | uniq -c | awk '$1 > {3} {{print $1, $2}}'".format(start_time, end_time, Logger_File, Limit_Num)
    result = subprocess.check_output(command, shell=True)
    if result:
        result = result.decode('utf-8') 
    return result

def get_db_order_domain():
    host_list = []
    result = get_tps_proxy_log().strip('\n')
    if result:
        tid_list = re.findall(r'\d+', result)
        for i in range(0,len(tid_list),2):
            count= int(tid_list[i])
            tid = 't'+tid_list[i+1]          
            host_list.append((tid,count))
        if host_list:
            return host_list
    return None
                
def get_tiktok_domain():
    host_list = get_db_order_domain()
    hostname = ""
    command = "hostname"
    hostname = subprocess.check_output(command, shell=True)
    if hostname:
        hostname = hostname.decode('utf-8')
    if host_list:
        title = "%s以下订单访问腾讯系域名失败"%hostname
        content = ""
        for tid,count in host_list:
            content += "订单号:%s, 失败次数:%s\n"%(tid,count)
        url = "https://open.feishu.cn/open-apis/bot/v2/hook/c6f343b3-f8e8-4e81-b27d-ae20bd31c495"
        headers = {
            "Content-Type": "application/json; charset=utf-8",
        }
        data = {
            "msg_type": "post",
            "content": {
            "post": {
                "zh_cn": {
                    "title": "[%s]"%title,
                    "content": [
                        [{
                                "tag": "text",
                                "text": content
                        }
                        ]
                    ]
                }
            }
            }
        }
        data = json.dumps(data)  
        req = urllib2.Request(url, data=data.encode('utf-8'), headers=headers)  
        response = urllib2.urlopen(req)  
        response.close()
        
if __name__ == '__main__':
    get_tiktok_domain()
