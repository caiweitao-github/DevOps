#!/usr/bin/python
# -*- coding: utf-8 -*-
"""主要功能:修复数据库中位置,运营商,登录IP和端口为空的异常数据,如若修复失败飞书推送至运维群内
运行机器: main13
启动命令: 1 8-21 * * * bash /home/httpproxy/DevOps/Dps/Dps_check_info.sh >> /dev/null 2>&1
运行方式：定时任务
周期： 每天一次
"""
import requests
from datetime import datetime
import dbutil

db = dbutil.get_db_db()

def notify(content):
    date_time = datetime.now().strftime("%H:%M:%S")# 格式化datetime.now()对象
    # 机器人webhook
    titok_url = ''
    headers = {
        "Content-Type": "application/json; charset=utf-8",
    }
    payload_message = {
       "msg_type": "post",
         "content": {
            "post": {
               "zh_cn": {
                  "title": "DPS信息检测异常 %s" %(date_time),
                  "content": [
                     [{
                              "tag": "text",
                              "text": "%s" %(content)
                     }
                     ]
                  ]
               }
            }
        }
    }
    requests.post(url=titok_url, json=payload_message, headers=headers)


def get_dps_info():
    """ 查询DPS状态不是下架机器,且位置,供应商,登录IP和登录端口为空的code等信息 """
    dps_info_list = []
    check_info_sql = "select code,provider_vps_name,login_ip,provider,ip from dps where (location = '' or carrier = '' or login_ip = '' or login_port = '')  and status = 1;"
    result = db.execute(check_info_sql).fetchall()
    for dps_info in result:
        code = dps_info[0]
        provider_vps_name = dps_info[1]
        login_ip = dps_info[2]
        provider = dps_info[3]
        ip = dps_info[4]
        dps_info_list.append((code,provider_vps_name,login_ip,provider,ip))
    return dps_info_list

# 调用接口获取代理IP解析的IP地址
def get_ipadd_location(ip):
    getByAddress_url = "http://ip.plyz.net/ip.ashx?ip=%s" % ip
    header = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.124 Safari/537.36"
    }
    res = requests.get(getByAddress_url,header).text
    
    province = res.split('|')[-1].split(' ')[1]
    if not province.endswith(u"省") :
         province = province+u"省"
    city = res.split('|')[-1].split(' ')[2]         
    if city in [u"北京市",u"重庆市",u"上海市",u"天津市"]:
            location = city
    else:
            location = province + city
    carrier = res.split('|')[-1].split(' ')[-1]

    return location,carrier

def get_location():
    dps_info = get_dps_info()
    content_list = []
    for i in dps_info:
        code = i[0]
        provider_vps_name = i[1]
        login_ip = i[2]
        provider = i[3]
        ip = i[4]
        if ip != '':
           location,carrier = get_ipadd_location(ip)
        else:
           if login_ip == '':
               content = u"异常机器为: %s,机器名称为:%s,暂无代理IP和远程登录IP,请联系%s供应商解决"%(code,provider_vps_name,provider)
               content = content.strip()
               content_list.append(content)
               continue
           else:
                select_sql = "SELECT location,carrier FROM dps WHERE login_ip = '%s' and location !='' GROUP BY location,carrier limit 1;"% login_ip
                result = db.execute(select_sql).fetchone()
                if result is not None:
                    location = result[0]
                    carrier = result[1]
                else:
                    content = u"异常机器为: %s,机器名称为:%s,暂无代理IP和远程登录IP,请联系%s供应商解决"%(code,provider_vps_name,provider)
                    content = content.strip()
                    content_list.append(content)
                    continue
        update_sql = "update dps set location = '%s',carrier='%s' where provider='%s' and code='%s';"%(location,carrier,provider,code)
        db.update(update_sql)
     
    if len(content_list) > 0:
       content = '\n'.join(content_list).strip()
       notify(content)

if __name__ == '__main__':
    get_location()
