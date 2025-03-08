#!/usr/bin/env python
# -*- coding: utf-8 -*-
# author_='Joinishu';
# date: 2021/8/20 11:38
"""dps异常节点重启脚本

脚本功能：周期性将异常时间大于30分钟的机器通过供应商的api重启，最多连续重启两次
运行方式：cron.d
运行周期：*/30 * * * *
"""
import json
import threading

import requests
import hashlib
import random
import string
import time

import dbutil
import loggerutil
from redisutil import redisdb_crs

kdl_db = dbutil.get_db_db()
logger = loggerutil.get_logger("devops", "devops/exception_node_auto_reboot.log")


def generate_random_str(randomlength=16):
    str_list = [random.choice(string.digits + string.ascii_letters) for i in range(randomlength)]
    random_str = ''.join(str_list)
    return random_str


def vps91_reboot(vps_name):
    username = "gizaworks"
    api_key = "WGYiHOXSa5bnVFjNZTfMDHn83IRHKKc9"
    nonce = generate_random_str(16)
    sign = hashlib.md5(("%s%s%s" % (username, nonce, api_key)).encode(encoding="UTF-8")).hexdigest()
    url = "https://www.91vps.com/api/v1/cloud/action?username=%s&vpsname=%s&nonce=%s&sign=%s&operation=restart" % (
        username, vps_name, nonce, sign)
    res = requests.post(url)
    res_text = res.json().get('info')
    logger.info("[91vps][%s] results→ %s" % (vps_name, res_text))


def vps_youyi_reboot(vps_name):
    userid = "18162334197"
    userstr = "Gizavps2021kdl"
    op = "reset"
    url = "https://my.150cn.com/api/cloudapi.html?action=power&userid=%s&userstr=%s&vpsname=%s&op=%s&format=json" % (
        userid, userstr, vps_name, op)
    res = requests.get(url)
    res_text = res.json().get('msg')
    logger.info("[有益网络][%s] results→ %s" % (vps_name, res_text))


def vps_qingguoyun_reboot(vps_name):
    productId = "113"
    appid = "qg2935ee"
    timestamp = time.time()
    domain = "www.qg.net"
    app_key = "5dbfbad56ada76955d58"
    sign = hashlib.md5(("%s%s%s%s" % (appid, domain, timestamp, app_key)).encode(encoding="UTF-8")).hexdigest()
    url = "https://www.qg.net/api/cloud-operate/restart?suid=%s&productId=%s&appid=%s&time=%s&domain=%s&sign=%s" % (
        vps_name, productId, appid, timestamp, domain, sign)
    res = requests.get(url)
    res_text = res.json().get("Data").encode("utf-8")
    logger.info("[青果云][%s] results→ %s" % (vps_name, res_text))


def vps_ylf2_reboot(vps_name):
    userid = "xiaogaovps"
    passwd = "gizavps"
    userstr = hashlib.md5(("%s%s" % (passwd, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.yunlifang.cn/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&action=vpsop&op=reset" % (userid, userstr, vps_name)
    res = requests.get(url)
    res_text = res.json().get("msg").encode("utf-8")
    logger.info("[云立方2][%s] results→ %s" % (vps_name, res_text))


def vps_ylf3_reboot(vps_name):
    userid = "yangvps"
    passwd = "gizavps"
    userstr = hashlib.md5(("%s%s" % (passwd, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.yunlifang.cn/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&action=vpsop&op=reset" % (userid, userstr, vps_name)
    res = requests.get(url)
    res_text = res.json().get("msg")
    logger.info("[云立方3][%s] results→ %s" % (vps_name, res_text))


def vps_bicheng_reboot(vps_name):
    url = "http://www.xiguaip.net/whapi/vps/thirdPartyApi/vpsOperation"
    headers = {
        'app_id': '1415048307411128320',
        'safety_token': '402892e67aa197e8017aa198c49e005e'
    }
    body = {
        'vpsNames': vps_name,
        'operationType': '3'
    }
    res = requests.post(url, headers=headers, data=body)
    res_text = res.json().get("message")
    logger.info("[必诚互联][%s] results→ %s" % (vps_name, res_text))


def vps_yangguang_reboot(vps_name):
    userid = "wangvps"
    action = "reset"  # reset:重启 turnoff:关机 start:开机
    ti = time.time()
    api_key = "3f4bd6558853cf91f19b482c37565898"
    nonce = generate_random_str(16)
    sign = hashlib.md5(("%s%s%s" % (ti, nonce, api_key)).encode(encoding="UTF-8")).hexdigest()
    url = "https://api.juzhenip.cn/v1/vpsHand?userid=%s&vmname=%s&action=reset&ti=%s&nonce=%s&sign=%s" % (userid, vps_name, ti, nonce, sign)
    res = requests.post(url)
    res_text = res.json().get("errcode")
    if int(res_text) == 0:
        logger.info("[阳光net][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[阳光net][%s] results→ 重启失败" % vps_name)


def vps_chaoba_reboot(vps_name):
    userid = "gizavps"
    passwd = "Gizacb2016"
    userstr = hashlib.md5(("%s%s" % (passwd, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.adslvps.com/api/cloudapiths.asp?userid=%s&userstr=%s&vpsname=%s&action=vpsop&op=reset" % (userid, userstr, vps_name)
    res = requests.get(url)
    res_text = res.json().get("ret")
    if res_text == 'ok':
        logger.info("[超巴网络][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[超巴网络][%s] results→ 重启失败" % vps_name)


def vps_91sy_reboot(vps_name):
    userid = "gizavps"
    api_key = "Gizavps2021"
    userstr = hashlib.md5(("%s%s" % (api_key, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.91soyun.com/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&op=reset&action=vpsop" % (userid, userstr, vps_name)
    res = requests.get(url)
    res_text = res.text
    if "ret=ok" in res_text:
        logger.info("[91搜云][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[91搜云][%s] results→ 重启失败" % vps_name)


def vps_jiaxinzhaocheng_reboot(vps_name):
    userid = "gizavps"
    api_key = "giza123654"
    userstr = hashlib.md5(("%s%s" % (api_key, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.zhekou5.com/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&op=reset&action=vpsop" % (userid, userstr, vps_name)
    res = requests.get(url)
    res_text = res.text
    if "ret=ok" in res_text:
        logger.info("[嘉兴昭诚][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[嘉兴昭诚][%s] results→ 重启失败" % vps_name)


# def vps_dsy_reboot(vps_name):
def vps_wanbian_reboot(vps_name):
    userid = "gizavps"
    passwd = "giza123654"
    passwd = hashlib.md5(("%s" % (passwd)).encode(encoding="UTF-8")).hexdigest()
    userstr = hashlib.md5(("%s%s" % (passwd, userid)).encode(encoding="UTF-8")).hexdigest()
    url = "http://www.wanbianyun.com/api/v4/vps.asp?username=%s&password=%s&allsnme=%s&action=reset" % (userid, userstr, vps_name)
    res = requests.post(url)
    res_text = res.text
    if "成功" in res_text:
        logger.info("[万变云][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[万变云][%s] results→ 重启失败" % vps_name)


def vps_tjy_reboot(vps_name):
    userid = "gizavps"
    api_key = "Giza6782016"
    userstr = hashlib.md5(("%s%s" % (api_key, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.vps678.com/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&op=reset&action=vpsop" % (userid, userstr, vps_name)
    res = requests.get(url)
    res_text = res.text
    if "ret=ok" in res_text:
        logger.info("[天际云][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[天际云][%s] results→ 重启失败" % vps_name)


def vps_bengxing_reboot(vps_name):
    userid = "gizavps1"
    action = "reset"  # reset:重启 turnoff:关机 start:开机
    ti = time.time()
    api_key = "60b83565d0605970717"
    nonce = generate_random_str(16)
    sign = hashlib.md5(("%s%s%s" % (ti, nonce, api_key)).encode(encoding="UTF-8")).hexdigest()
    url = "https://api.juzhenip.cn/v1/vpsHand?userid=%s&vmname=%s&action=reset&ti=%s&nonce=%s&sign=%s" % (userid, vps_name, ti, nonce, sign)
    res = requests.post(url)
    res_text = res.json().get("errcode")
    if int(res_text) == 0:
        logger.info("[奔星网络][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[奔星网络][%s] results→ 重启失败" % vps_name)


def vps_189vps_reboot(vps_name):
    userid = "gizavps"
    api_key = "Giza123654"
    userstr = hashlib.md5(("%s%s" % (api_key, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.189vps.com/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&op=reset&action=vpsop" % (userid, userstr, vps_name)
    res = requests.get(url)
    res_text = res.text
    if "ret=ok&vpsname" in res_text:
        logger.info("[189vps][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[189vps][%s] results→ 重启失败" % vps_name)


def vps_dywl_reboot(vps_name):
    userid = "gizavps"
    api_key = "gizavps"
    userstr = hashlib.md5(("%s%s" % (api_key, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://www.diyavps.com/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&op=reset&action=vpsop" % (userid, userstr, vps_name)
    res = requests.post(url)
    res_text = res.text
    if "ret=ok&vpsname" in res_text:
        logger.info("[迪亚网络][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[迪亚网络][%s] results→ 重启失败" % vps_name)


def vps_263vps_reboot(vps_name):
    userid = "gizavps"
    api_key = "gizavps"
    userstr = hashlib.md5(("%s%s" % (api_key, "7i24.com")).encode(encoding="UTF-8")).hexdigest()
    url = "http://api.263vps.com:88/api/cloudapi.asp?userid=%s&userstr=%s&vpsname=%s&op=reset&action=vpsop" % (
    userid, userstr, vps_name)
    res = requests.post(url)
    res_text = res.text
    if "ret=ok&vpsname" in res_text:
        logger.info("[263vps][%s] results→ 重启成功" % vps_name)
    else:
        logger.error("[263vps][%s] results→ 重启失败" % vps_name)


def get_alive_list(provider):
    alive_list = []
    sql = "select provider_vps_name from dps where provider='%s' and status=1" % provider
    cursor = kdl_db.execute(sql)
    results = cursor.fetchall()
    for result in results:
        alive_list.append(result[0])
    return alive_list


def reboot_process(provider):
    exclude_list = []
    reboot_count_list = list(redisdb_crs.hgetall("dps_reboot_count").iterkeys())
    alive_list = get_alive_list(provider)
    sql = "select provider_vps_name,dead_time from dps where provider='%s' and status=3 and dead_time < DATE_SUB(now(),interval 30 minute) order by dead_time" % provider
    cursor = kdl_db.execute(sql)
    results = cursor.fetchall()
    ecs_list = ["10051OVF1FT", "10051NDDZTN", "10051164PLX", "10051R498N0", "10051KYR45Q", "10051VDS4VH", "10051HEEQDT"]
    # 清除重启次数标记
    for provider_vps_name in alive_list:
        if provider_vps_name in reboot_count_list:
            redisdb_crs.hdel("dps_reboot_count", provider_vps_name)
            logger.info("[%s][%s]  recover" % (provider, provider_vps_name))
    for row in results:
        vps_name = row[0]
        # 标记重启次数
        reboot_count = redisdb_crs.hincrby("dps_reboot_count", vps_name)
        # 即使没有自动恢复，6小时后重置标记
        if reboot_count >= 12:
            redisdb_crs.hdel("dps_reboot_count", vps_name)
        if reboot_count >= 3:
            exclude_list.append(vps_name)
            logger.debug("[%s][%s] exclude,reboot_count:%s" % (provider, vps_name, reboot_count))
        else:
            try:
                if provider == '91vps' and vps_name not in ecs_list:
                    vps91_reboot(vps_name)
                if provider == '有益网络':
                    vps_youyi_reboot(vps_name)
                if provider == '青果云':
                    vps_qingguoyun_reboot(vps_name)
                if provider == '云立方2':
                    vps_ylf2_reboot(vps_name)
                if provider == '云立方3':
                    vps_ylf3_reboot(vps_name)
                if provider == '必诚互联':
                    vps_bicheng_reboot(vps_name)
                if provider == '淘宝-阳光net':
                    vps_yangguang_reboot(vps_name)
                if provider == '超巴网络':
                    vps_chaoba_reboot(vps_name)
                if provider == '91搜云':
                    vps_91sy_reboot(vps_name)
                if provider == '嘉兴昭诚':
                    vps_jiaxinzhaocheng_reboot(vps_name)
                if provider == '万变云':
                    vps_wanbian_reboot(vps_name)
                if provider == '天际云':
                    vps_tjy_reboot(vps_name)
                if provider == '奔星网络':
                    vps_bengxing_reboot(vps_name)
                if provider == '189vps':
                    vps_189vps_reboot(vps_name)
                if provider == '迪亚网络':
                    vps_dywl_reboot(vps_name)
                if provider == '263vps':
                    vps_263vps_reboot(vps_name)
            except Exception as e:
                logger.error("[%s]ERROR-%s:%s reboot error" % (provider, repr(e), vps_name))


if __name__ == '__main__':
    thread_list = []
    provider_list = ["91vps", "有益网络", "青果云", "云立方2", "云立方3", "必诚互联", "淘宝-阳光net", "超巴网络", "91搜云", "嘉兴昭诚", "万变云", "天际云", "奔星网络", "189vps", "迪亚网络", "263vps"]
    # try:
    #     # 每个供应商创建1个重启线程
    #     for i in provider_list:
    #         t_process = threading.Thread(target=reboot_process, args=(i,))
    #         thread_list.append(t_process)
    #     # 启动线程
    #     for t in thread_list:
    #         t.setDaemon(True)  # 将子线程设成守护进程，主线程不会等待子线程结束，主线程结束子线程立刻挂
    #         t.start()
    # except Exception as e:
    #     logger.error(str(e))
    try:
        for provider in provider_list:
            reboot_process(provider)
    except Exception as e:
        logger.error(str(e))
