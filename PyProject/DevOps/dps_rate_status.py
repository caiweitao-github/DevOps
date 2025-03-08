# -*- coding: utf-8 -*-

"""
采用ab评分节点
"""

import re
import time
import datetime
import subprocess
import urllib2
import zlib
from hashlib import md5
import json
import loggerutil
import ckutil


from redisutil import redisdb, redisdb_crs

ck = ckutil.get_ck_nodeops()
logger = loggerutil.get_logger("devops", "devops/dps_rate_status.log")
MASTER_DOMAIN = "http://tsdpsmaster.gizaworks.com"
API_TOKEN = "solau0zerz8x09np"


def api_get(url):
    req = urllib2.Request(url)
    auth = md5('%s_%s' % (API_TOKEN, "kdl")).hexdigest()
    req.add_header('API-AUTH', auth)
    req.add_header('Accept-Encoding', 'gzip')
    res = urllib2.urlopen(req)

    return res

def api_post(url, data, compress=False, timeout=60):
    post_data = zlib.compress(data) if compress else data
    req = urllib2.Request(url, post_data)
    auth = md5('%s_%s' % (API_TOKEN, "kdl")).hexdigest()
    req.add_header('API-AUTH', auth)
    req.add_header('Accept-Encoding', 'gzip')
    if compress:
        req.add_header('content-encoding', 'zlib')

    res = urllib2.urlopen(req)
    return res

def read_content(res):
    accept_encoding = res.headers.get('content-encoding', '')
    content = res.read()
    if 'gzip' in accept_encoding.lower():
        if content:
            return zlib.decompress(content, 16 + zlib.MAX_WBITS)
    return content

def get_dps_list():
    node_list = []
    res = api_get(MASTER_DOMAIN + "/getratedpslist")
    if res.getcode() != 200:
        logger.info('master getratedpslist error, response code: %d' % res.getcode())
        return node_list
    content = read_content(res)
    json_obj = json.loads(content)
    if json_obj["code"] == 0:
        node_list = json_obj["data"]
        logger.info("get dps count: %d" % len(node_list))
    else:
        logger.info("getratedpslist ERROR(%d): %s" % (json_obj["code"], json_obj["msg"]))
    return node_list

def choose_list(dps_list):
    choose_list = []
    dps_list = [dps for dps in dps_list if dps['status'] == 1]

    dps_list = sorted(dps_list, key=lambda x: int(time.mktime(time.strptime(x['last_changeip_time'], '%Y-%m-%d %H:%M:%S'))), reverse=True)
    key = "check_rate_" + datetime.datetime.now().strftime("%Y%m%d")
    for i in range(len(dps_list)):
        dps = dps_list[i]
        code = dps['code']
        if not redisdb.hexists(key, code):
            # redisdb.hset(key, code, 1)
            # logger.info("choose [%s], last change at:%s" % (code, dps["last_changeip_time"]))
            choose_list.append(dps)
    return choose_list[:5]

def choose_one(dps_list):
    node = None

    dps_list = sorted(dps_list, key=lambda x: int(time.mktime(time.strptime(x['last_changeip_time'], '%Y-%m-%d %H:%M:%S'))), reverse=True)
    key = "check_rate_" + datetime.datetime.now().strftime("%Y%m%d")
    for dps in dps_list:
        code = dps['code']
        if not redisdb.hexists(key, code):
            logger.info("choose [%s], last change at:%s" % (code, dps["last_changeip_time"]))
            node = dps
            break
    return node

def do_ab_test(node):
    proxy = '%s:%s' % (node['ip'], node['port'])
    logger.info("start test [%s]%s" % (node['code'], proxy))
    cmd = 'ab -c 2 -n 20 -X %s -P dpstest:3syu1otf http://m.baidu.com/' % proxy
    p1 = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
    content = p1.stdout.read()
    m = re.findall('Time per request:\s*([0-9\.]*)', content, re.S)
    result = {}
    if m:
        coast = float(m[-1])
        logger.info("Time per request:%.2f" % coast)
        result["time_per_request"] = coast
    m = re.findall('Transfer rate:\s*([0-9\.]*)', content, re.S)
    if m:
        rate = float(m[0])
        logger.info("Transfer rate:%.2f" % rate)
        result["transfer_rate"] = rate
    logger.info("end test")
    return result

def get_score(result):
    if result < 0:
        return 0
    elif result < 600:
        return 1
    elif result < 1000:
        return 2
    else:
        return 3

def report_rate(rate_list):
    url = MASTER_DOMAIN + "/reportrate"
    data = {"rate_list": rate_list}
    res = api_post(url, json.dumps(data), compress=True)

    if res.getcode() != 200:
        logger.info('master getratedpslist error, response code: %d' % res.getcode())
        return False

    content = read_content(res)
    json_obj = json.loads(content)
    if json_obj["code"] == 0:
        return True

    return False

def main():
    while True:
        try:
            dps_list = get_dps_list()
            node = choose_one(dps_list)

            if node:
                t0 = int(time.time())
                for i in range(10):
                    if int(time.time()) - t0 < 30:
                        result = do_ab_test(node)
                        if result:
                            code = node["code"]

                            # 上报
                            score = get_score(result["time_per_request"])
                            rate_list = [{"code": code, "score": score}]
                            if report_rate(rate_list):
                                # 检查标记
                                key = "check_rate_" + datetime.datetime.now().strftime("%Y%m%d")
                                redisdb.hset(key, code, 1)

                            break
                        else:
                            time.sleep(1)

        except Exception as e:
            logger.exception(str(e))
        time.sleep(5)

if __name__ == '__main__':
    main()

