#! /usr/bin/python
# -*- coding: UTF-8 -*-


"""
检测策略：检测过去220内秒的IP产出量，若>2则认为其code重复导致循环上报
"""

#基础库
import time
import requests
#日志
import loggerutil
logger = loggerutil.get_logger("devops", "devops/code_repeat_check.log")

#数据库
import dbutil
db_nodeops = dbutil.get_db_nodeops()
db_db = dbutil.get_db_db()
def send_msg(mesg,error_info):
    """
    飞书告警
    """
    if error_info == "":
        error_info = ""
    mesg = mesg
    date_time = time.strftime("%H:%M:%S", time.localtime())
    data = {
            "msg_type": "post",
            "content": {
                    "post": {
                            "zh_cn": {
                                    "title": "[dps code重复监测]  %s [kdlmain13]" % (date_time),
                                    "content": [
                                            [{
                            "tag": "text",
                            "text": "%s\n%s" % (mesg,error_info)
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


def get_except_code():
    "获取海外vps的code。不检测海外vps的重复上报！！！"
    li = []
    sql = "select code from dps where dps_type = 2 and status in (1,3)"
    cursor = db_db.execute(sql)
    code_info = cursor.fetchall()
    for i in code_info:
        li.append(i[0])
    return li

def get_node():
    "通过nodeops表查看重复ip的dps机器！！！"
    try:
        except_list = get_except_code()
        To_Handle_list = []
        sql = '''
            select dps_code,ip,count(ip) from dps_changeip_history where change_time > DATE_SUB(NOW(), INTERVAL 200 second) group by dps_code,ip having count(ip)>2;
        '''
        cursor = db_nodeops.execute(sql)
        if cursor is not None:
            dps_code_results = cursor.fetchmany(2)          #提取俩条记录,code重复检测每次都是报俩次！！！
            if dps_code_results:
                for i in dps_code_results:
                    if i[0] not in except_list:             #判断dps是否为海外dps,海外不检测重复上报！！！
                        Handle_info = i[0],i[1],i[2]
                        To_Handle_list.append(Handle_info)
                return To_Handle_list
            else:
                return True
    except Exception as e:
        logger.error(e)
        # print(e)

def Duplicate_IP_vps():
    "获取这批机器的ip地址"
    import subprocess
    try:
        result = get_node()
        print(result)
        if result != True:
            send_list = []
            for item in result:
                dps_code = item[0]
                area_ip = item[1]
                num = item[2]
                command = ["curl", "-s", "cip.cc/%s" % area_ip]
                process = subprocess.Popen(command, stdout=subprocess.PIPE)
                output, _ = process.communicate()
                if process.returncode == 0:
                    time.sleep(0.5)             #运行过快可能导致ip.curl不到
                    output_lines = output.strip().split('\n')
                    area = output_lines[1].split(':')[1].strip().replace(' ','').replace('中国','')
                    dps_info = "%s->%s->%s->%s" % (dps_code,area_ip,area,num)
                    logger.info("以下code存在ip重复上报：%s" % dps_info)
                    send_list.append(dps_info)
                    print(dps_info)
            send_msg("以下code存在ip重复上报:",error_info='\n'.join(send_list))
        else:
            logger.info("未发现code重复ip上报！！！")
    except Exception as e:
        logger.exception("code重复上报检测：%s" % e)
if __name__ == '__main__':
    Duplicate_IP_vps()
    
