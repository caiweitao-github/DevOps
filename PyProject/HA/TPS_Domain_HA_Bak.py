#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import re
import sys
import log
import dbutil
from tps_tools import *
from HA.tps_ssh_util import *
from Aliyun_functions import *
from redisutil import redisdb

dbdb = dbutil.get_db_db()

feishu = get_feishu_tpsHA()

TPS_HA_IP = '221.131.165.14'
directory = sys.argv[0]

CREATE_TPS_COUNT = 'create_tps_count'
TPS_HA_INFO_KEY = '_ha_info'
EXCEPTION_TPS_COUNT = 'exception_tps'
CMD_LIST = [
    "sed -ri 's/(.*)tpsb1/\\1tpsb%s/g' /root/.bash_profile",
    "sed -ri 's/(.*)tpsb1/\\1tpsb%s/g' /root/tps/tps.cfg",
    "sed -ri 's/(.*)nqsb1/\\1nqsb%s/g' /root/tps_nqs/tps_nqs.cfg",
    "sed -ri 's/tpsb1/tpsb%s/g' /root/tps_go/config.yaml",
    "sed -ri 's/nqsb1/nqsb%s/g' /root/tps_go/config.yaml",
    "cd /root;source .bash_profile;cd /root/tps; bash restart_auth.sh;sleep 1;python transfer_restart.py;cd /root/tps_nqs;bash restart_auth.sh;bash restart_webapi.sh;sleep 1;cd /root/tps_go/;bash restart_transfer.sh"
    ]

def exclude_domain():
    sql = "select host,host2 from tunnel where user_id in (61301, 74342) and status = 1"
    cursor = dbdb.execute(sql)
    doamin_tuple = cursor.fetchall()
    if doamin_tuple:
        domain = map(lambda x : str(re.split('[.]', x)[0]), doamin_tuple[0])
        return tuple(domain)
    return False


def get_tps_code():
    sql = "select code from tps where status in ('1', '3')"
    cursor = dbdb.execute(sql)
    tps_code = cursor.fetchall()
    cursor.close()
    return tps_code

def get_exception_tps():
    sql = "select count(*) from tps where status = 3"
    cursor = dbdb.execute(sql)
    exception_tps_num = cursor.fetchone()
    cursor.close()
    if exception_tps_num[0] >= 5:
        if not redisdb.set(EXCEPTION_TPS_COUNT, 1, nx=True):
            redisdb.incr(EXCEPTION_TPS_COUNT)
    else:
        redisdb.set(EXCEPTION_TPS_COUNT, 0)
    res = redisdb.get(EXCEPTION_TPS_COUNT)
    if int(res) >= 3:
        return True
    return False

    
def update_tps_status(tps_info, tps_ip=None, nqs_name=None):
    sql_list = [ "update tps set status = 1, login_ip = '%s', nqs_domain = '%s.gizaworks.com', nqs_code = '%s' where code = '%s'" ,"update tps set status = 2 where login_ip = '%s'" ]
    check = re.compile('^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$')
    if check.match(tps_info):
        sql = sql_list[1] %(tps_info)
    else:
        sql = sql_list[0] %(tps_ip, nqs_name, nqs_name, tps_info)
    cursor = dbdb.execute(sql)
    cursor.close()

def get_tps_domain():
    tps_code_tuple = get_tps_code()
    for code in tps_code_tuple:
        sql = "select domain from tps_domain where tps_id = (select id from tps where code = '%s')" %(code[0])
        cursor = dbdb.execute(sql)
        domain_list = cursor.fetchall()
        list_name = code[0] + "_domain_list"
        domain_name = []
        for x in domain_list:
            domain_name.append(x[0])
        for i in domain_name:
            key_is_exists = redisdb.sismember(list_name, i)
            if not key_is_exists:
                redisdb.sadd(list_name, i)
        get_all_tps_domain = redisdb.smembers(list_name)
        for i in get_all_tps_domain:
            i = i.decode()
            if i not in domain_name:
                redisdb.srem(list_name, i)
    cursor.close()

    
def exec_tps_cmd(ip, cmd):
    ip_check = re.compile('^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$')
    if ip_check.match(ip):
        tps_ssh = TpsSSH()
        login_result = tps_ssh.login_tps(ip)
        if login_result:
            tps_ssh.exec_cmd(cmd)
        return True
    else:
        log.error("%s ---> 格式检查失败!" %(ip))
    tps_ssh.close()
    
def check_tps_status():
    result = get_exception_tps()
    sql = "select id,code,login_ip,status from tps where status in ('1', '3') and code not REGEXP 'tpsb.*|tpsT.*|tpstest|tpsysy'"
    cursor = dbdb.execute(sql)
    tps_info = cursor.fetchall()
    cursor.close()
    for id,code,ip,status in tps_info:
        tps_redis_status = redisdb.hget(code, 'status')
        tps_code_is_exists = redisdb.exists(code)
        if not tps_code_is_exists:
            if int(status) == 3:
                redisdb.hmset(code, {'id':id, 'ip': ip, 'status': 0, 'ha': 0, 'ha_ip': 'None'})
                feishu.error("node %s %s down!" %(code, ip))
            else:
                redisdb.hmset(code, {'id':id, 'ip': ip, 'status': status, 'ha': 0, 'ha_ip': 'None'})
        elif int(status) != int(tps_redis_status) and int(status) == 3:
            if int(tps_redis_status) > -2:
                feishu.error("node %s %s down!" %(code, ip))
                redisdb.hincrby(code, 'status', amount=-1)
            status_num = redisdb.hget(code, 'status')
            ha_status = redisdb.hget(code, 'ha')
            if int(status_num) == -2 and int(ha_status) == 0 and ip != TPS_HA_IP:
                exclude_user_domain = exclude_domain()
                domain_list = []
                feishu.error("node %s %s 连续三次检测为异常，切换所有域名至备用机。" %(code, ip))
                list_name = code + "_domain_list"
                for tps_domain in redisdb.sscan_iter(list_name):
                    domain_name = re.split('[.]',tps_domain.decode())[0]
                    domain_list.append(domain_name)
                if not result:
                    for domain in domain_list:
                        if exclude_user_domain:
                            if domain not in exclude_user_domain:
                                update_tpsdomain_record(domain, TPS_HA_IP, TTL_time=1)
                            else:
                                domain_list.remove(domain)
                                feishu.info("排除域名： %s" %(domain))
                        else:
                            update_tpsdomain_record(domain, TPS_HA_IP, TTL_time=1)
                    redisdb.hmset(code, {'ha': 1, 'ha_ip': TPS_HA_IP})
                    feishu.error("node %s 域名：%s 切换至%s。" %(code, domain_list, TPS_HA_IP))     
                else:
                    if not redisdb.set(CREATE_TPS_COUNT, '1', ex=3600 * 24, nx=True):
                        redisdb.incr(CREATE_TPS_COUNT)
                    tps_num = redisdb.get(CREATE_TPS_COUNT)
                    if int(tps_num) >= 5:
                        
                        feishu.error("创建机器数量超出限制！")
                    else:
                        tps_name = int(redisdb.get(CREATE_TPS_COUNT))
                        tps_new_ip, server_id = create_tps("ap-beijing-6", "S6.2XLARGE16", 100, "tpsb%s-nqsb%s" %(tps_name, tps_name))
                        if tps_new_ip and server_id:
                            redisdb.hmset(code, {'ha': 1, 'ha_ip': tps_new_ip})
                            redisdb.hmset(code + TPS_HA_INFO_KEY, {'server_id': server_id, 'ip': tps_new_ip})
                            for index, cmd in enumerate(CMD_LIST):
                                if index == 5:
                                    cmd_res = exec_tps_cmd(tps_new_ip, cmd)
                                else:
                                    cmd_res = exec_tps_cmd(tps_new_ip, cmd %(tps_name))
                            if cmd_res:
                                for domain in domain_list:
                                    if exclude_user_domain:
                                        if domain not in exclude_user_domain:
                                            update_tpsdomain_record(domain, tps_new_ip, TTL_time=1)
                                        else:
                                            domain_list.remove(domain)
                                            feishu.info("排除域名： %s" %(domain))
                                    else:
                                        update_tpsdomain_record(domain, tps_new_ip, TTL_time=1)
                                nqs_name = 'nqsb' + str(tps_name)
                                domain_info = query_some_domain_record(nqs_name, DomainName="gizaworks.com", show_print=0)
                                for i in domain_info:
                                    if i['RR'] == nqs_name and i['Remark'] == nqs_name:
                                        update_domain_record(RecordId=i['RecordId'], new_Value=tps_new_ip, RR=i['RR'], TTL_time='1',Type='A')
                                        log.info("nqsb%s ---> %s" %(tps_name, tps_new_ip))
                                update_tps_status(tps_ip=tps_new_ip, nqs_name=nqs_name, tps_info='tpsb%s' %(tps_name))
                                feishu.info("node %s 域名：%s 切换至%s。" %(code, domain_list, tps_new_ip))
                            else:
                                feishu.error("node %s ---> 远程连接失败！" %(tps_new_ip))
        elif int(status) != int(tps_redis_status) and int(status) == 1:
            redisdb.hincrby(code, 'status', amount=1)
            ha_status = redisdb.hget(code, 'ha')
            ha_ip = redisdb.hget(code, 'ha_ip').decode()
            if int(ha_status) == 1 and ip != TPS_HA_IP and ha_ip != 'None':
                domain_list = []
                if int(tps_redis_status) == 0:
                    redisdb.hset(code, 'ha', 0)
                    list_name = code + "_domain_list"
                    for domain in redisdb.sscan_iter(list_name):
                        domain_name = re.split('[.]',domain.decode())[0]
                        domain_list.append(domain_name)
                        update_tpsdomain_record(domain_name, ip, TTL_time=600)
                    redisdb.hset(code, 'ha_ip', 'None')
                    if ha_ip != TPS_HA_IP:
                        redisdb.expire(code + TPS_HA_INFO_KEY ,time=60 * 30)
                    feishu.info("node %s %s 恢复，切回所有域名 %s。" %(code, ip, domain_list))
        
def del_server():
    key_list = redisdb.keys('*%s' %(TPS_HA_INFO_KEY))
    if key_list:
        for key in key_list:
            valid_time = redisdb.ttl(key)
            if int(valid_time) < 900 and int(valid_time) != -1:
                ip = redisdb.hget(key, 'ip').decode()
                server_id = str(redisdb.hget(key, 'server_id').decode())
                if server_id:
                    tps = Tps(region_key='ap-beijing-6')
                    results = tps.destruction_server(server_id)
                    if results:
                        update_tps_status(tps_info=ip)
                        redisdb.delete(key)
                        feishu.info("实例：%s ---> %s 销毁成功" %(server_id, ip))
    
    
if __name__ == '__main__':
    log.info('run %s...' %(os.path.basename(directory)))
    try:
        get_tps_domain()
        check_tps_status()
        del_server()
    except Exception as e:
        log.error(e)
