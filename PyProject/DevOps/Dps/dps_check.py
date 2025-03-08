#! /usr/bin/python3
# -*- coding: UTF-8 -*
# Function:使用Python查询db数据库并和各个供应商后台数据进行对比
import requests
import time
import hashlib
import random
import string
import pymysql
import json

class Result:
# __init__()是Python中定义类时常用的特殊方法之一，也称为构造函数
# 它的作用是在创建对象时进行初始化操作，即给对象的属性赋初值。
# 当一个类被实例化时，Python会自动调用__init__()方法，将创建的实例作为self参数传入，
# 同时将实例化时传入的参数传入到__init__()中进行处理。
# 在__init__()方法中，可以定义和初始化实例的属性，这些属性是与实例相关的数据。
# 通过使用self关键字，可以将这些属性绑定到实例上
    def __init__(self, provider, vpsname, vps_num):
        self.provider = provider
        self.vpsname = vpsname
        self.vps_num = vps_num
# 这段代码是定义了一个名为Result的类,它包含了三个属性:`provider`,`vpsname`,`vps_num`,以及一个构造函数`__init__()`
# 当实例化`Result`对象时,需要传递三个参数:`provider`,`vpsname`和`vps_num`.这些参数将会被传递给构造函数`__init__()`
# 构造函数`__init__()`将这些参数分别赋值给`self.provider`,`self.vps_name`和`self.vps_num`,并把它们保存在`Result`对象的属性中,以便后续使用

class db_provider:
    def db(provider):
        # 打开数据库连接
        db = pymysql.connect(
            host='10.0.5.41',#db数据库所在主机
            port=3306,
            user='db',
            passwd='db',
            db='db',
            charset='utf8'
        )

        # 打开游标
        cursor = db.cursor()
        # 执行SQL查询
        # sql = f"select provider_vps_name from dps where status !=4 and provider={provider};"
        # 为了防止SQL注入,使用参数化查询,%s是参数占位符.将参数的值作为元组传递给execute方法
        provider = provider
        sql = "select provider_vps_name from dps where status !=4 and provider=%s;"
        params = (provider,)
        cursor.execute(sql,params)
        # 获取查询结果
        result = cursor.fetchall()
        list = []
        for row in result:
            row = ''.join(row)
            list.append(row)
        # 关闭数据库连接
        db.close()
        db_vps_name = list
        db_vps_num = len(list)
        return db_vps_name,db_vps_num

    class provider:
        def NineOne_vps():
            str_random=string.ascii_lowercase+string.digits
            salt=[]
            for i in range(16):
                salt.append(random.choice(str_random))
                nonce = ''.join(salt)
            # 随机字符串
            # print(nonce)

            # 用户名
            username='gizaworks'

            # sign
            apikey="WGYiHOXSa5bnVFjNZTfMDHn83IRHKKc9"
            date=username+nonce+apikey
            sign=hashlib.md5(date.encode('utf-8')).hexdigest()

            api_url="https://www.91vps.com/api/v1/cloud/getAllCloudInfo"
            headers = {
                    "User-Agent":"Mozilla/6.0 (Windows; U; Windows NT 6.0; en-US) Gecko/2009032609 (KHTML, like Gecko) Chrome/2.0.172.6 Safari/530.7"
                    }
            data={
                'username':username,
                'host_type':'all',
                'nonce':nonce,
                'sign':sign
            }
            response=requests.post(url=api_url,headers=headers,data=data)
            # print(response.text)
            # print(response.json())
            vpsname = []
            data = response.json()['data']['lists']
            data_len = len(data)
            for i in range(data_len):
                vpsname.append(data[i]['vps_name'])
            vps_num = len(vpsname)
            return vpsname,vps_num

        def NineOneSouYun():
            str_random=string.ascii_lowercase+string.digits
            salt=[]
            for i in range(13):
                salt.append(random.choice(str_random))
                nonce = ''.join(salt)
            # 随机字符串
            # print(nonce)

            # 时间戳UTC/GMT+08:00
            now=str(time.time()).split('.')[0]


            # sign
            apikey="Gizavps2021"
            date=now+nonce+apikey
            sign=hashlib.md5(date.encode('utf-8')).hexdigest()
            # print(sign)

            url="http://zhukong.91soyun.com/v2/vmList"
            headers = {
                    "User-Agent":"Mozilla/6.0 (Windows; U; Windows NT 6.0; en-US) Gecko/2009032609 (KHTML, like Gecko) Chrome/2.0.172.6 Safari/530.7"
                    }
            data={
            "agentid":"gizavps",
                "ti":now,
                "nonce":nonce,
                "sign":sign
            }
            try:
                vpsname = []
                response=requests.post(url=url,headers=headers,data=data)
                if response.status_code==200:
                    vpsname = []
                    data = response.json()['data']
                    data_len = len(data)
                    for i in range(data_len):
                        vpsname.append(data[i]['vmname'])
                    vps_num = len(vpsname)
                else:
                    print("请检查相关参数")
                    vps_num = 0

            except ValueError:
                pass
            finally:
                pass
            return vpsname,vps_num
        def hean_yun():
            # 时间戳
            ti = str(time.time()).split('.')[0]
            # 随机字符串
            str_random = string.ascii_lowercase+string.digits
            salt = []
            for i in range(13):
                salt.append(random.choice(str_random))
                nonce = ''.join(salt)
            # print(nonce)
            # sign
            apikey="992c539736cee63267322c1c861cb2a7"
            date=ti+nonce+apikey
            sign=hashlib.md5(date.encode('utf-8')).hexdigest()

            api_url="http://106.14.117.11:8001/v2/vmList"
            headers = {
                    "User-Agent":"Mozilla/6.0 (Windows; U; Windows NT 6.0; en-US) Gecko/2009032609 (KHTML, like Gecko) Chrome/2.0.172.6 Safari/530.7"
                    }
            data={
                "agentid":"gizavps",
                "ti":ti,
                "nonce":nonce,
                "sign":sign
            }
            headers = {
                "User-Agent":"Mozilla/6.0 (Windows; U; Windows NT 6.0; en-US) Gecko/2009032609 (KHTML, like Gecko) Chrome/2.0.172.6 Safari/530.7"
                }
            try:
                response = requests.get(url=api_url,headers=headers,params=data)
                response.encoding='utf-8'
                stauts = response.status_code
                vpsname = []
                if stauts == 200:
                    vpsname = []
                    data = response.json()['data']
                    data_len = len(data)
                    for i in range(data_len):
                        vpsname.append(data[i]['vmname'])
                    vps_num = len(vpsname)
                    return vpsname,vps_num 
            except:
                pass
            finally:
                pass


        def YouYi_vps():
            vpsname = []
            api_url="https://www.150cn.com/api/cloudapi.asp"
            data={
                'userid':'18162334197',
                'userstr':'Gizavps2021kdl',
                'action':'list'
            }
            try:
                response = requests.get(url=api_url,data=data)
                if response.status_code==200:
                    data = response.json()
                    vps_name = data['result']
                    vps_num = len(vps_name)
                    for i in range(vps_num):
                        vpsname.append(vps_name[i]['name'])
                    return vpsname,vps_num
            except ValueError:
                pass
            finally:
                pass

        def XiGuaip():
            str_random=string.ascii_lowercase+string.digits
            salt=[]
            for i in range(13):
                salt.append(random.choice(str_random))
                nonce = ''.join(salt)
            # 随机字符串
            # print(nonce)

            # 时间戳UTC/GMT+08:00
            now=str(time.time()).split('.')[0]

            # sign
            apikey="98d740efd3a8520a"
            date=now+nonce+apikey
            sign=hashlib.md5(date.encode('utf-8')).hexdigest()


            # API接口地址
            api_url="http://zhukong.xiguaip.com/v2/vmList"
            headers = {
                    "User-Agent":"Mozilla/6.0 (Windows; U; Windows NT 6.0; en-US) Gecko/2009032609 (KHTML, like Gecko) Chrome/2.0.172.6 Safari/530.7"
                    }
            data={
                "agentid":"gizavps",
                "ti":now,
                "nonce":nonce,
                "sign":sign
            }
            try:
                response=requests.post(url=api_url,headers=headers,data=data)
                if response.status_code==200:
                    vpsname = []
                    data = response.json()
                    vps_num = len(data['data'])
                    for i in range(vps_num):
                        vpsname.append(data['data'][i]['vmname'])
                    vps_num = len(vpsname)
                    return vpsname,vps_num
                # print(vps_num)
            except ValueError:
                pass
            finally:
                pass


        def YunLiFang2():
            # 定义api_url
            api_url="http://api.yunlifang.cn/api/cloudapi.asp"
            # 云立方2用户名
            userid="xiaogaovps",
            password="gizavps"
            str="7i24.com"
            date=password+str
            userstr=hashlib.md5(date.encode('utf-8')).hexdigest()

            # 定义param
            data = {
                'userid':userid,  #代理用户名
                'userstr':userstr, #代理密码
                'action':'listinfo' #操作类型
            }
            try:
                response = requests.get(url=api_url,params=data)
                data = response.text
                vps_name = response.text.split("{[")[1].split("]}")[0].replace("\"",'').replace(",",'\n')
                vpsname = []
                for line in vps_name.splitlines():
                    vpsname.append(line)
                vps_num = len(vpsname)  
                # print(vps_num)
                return vpsname,vps_num

            except ValueError:
                pass
            finally:
                pass

        def YunLiFang3():
                        # 定义api_url
            api_url="http://api.yunlifang.cn/api/cloudapi.asp"
            # 云立方3信息
            userid="yangvps",
            password="Yang1234"
            str="7i24.com"
            date=password+str
            userstr=hashlib.md5(date.encode('utf-8')).hexdigest()

            params={
                'userid':userid,  #代理用户名
                'userstr':userstr, #代理密码
                'action':'listinfo' #操作类型
            }
            try:
                response = requests.get(url=api_url,params=params)
                data = response.text
                vps_name = response.text.split("{[")[1].split("]}")[0].replace("\"",'').replace(",",'\n')
                vpsname = []
                for line in vps_name.splitlines():
                    vpsname.append(line)
                vps_num = len(vpsname)
                # print(vps_num)
                return vpsname,vps_num

            except ValueError:
                pass
            finally:
                pass
    def diff_provider(db_vps_name,db_vps_num,vpsname,vps_num,provider):
        if db_vps_num == vps_num:
            pass
        elif db_vps_num > vps_num:
            if provider == "91vps":
                db_vps_name = [string.lower() for string in db_vps_name]
                vpsname = [string.lower() for string in vpsname]
                diff = set(db_vps_name)-set(vpsname)
                diff_num = len(diff)
                diff = '\n'.join(diff)
                db_provider.tiktok(provider,diff_num,diff,chooise=1)
            else:
                # print(vps_num)
                diff = set(db_vps_name)-set(vpsname)
                # print(f"{provider}在数据库中以下机器信息填写错误{diff}")
                diff_num = len(diff)
                diff = '\n'.join(diff)
                db_provider.tiktok(provider,diff_num,diff,chooise=1)
        else:
                if provider == "91vps":
                    db_vps_name = [string.lower() for string in db_vps_name]
                    vpsname = [string.lower() for string in vpsname]
                diff = [x for x in vpsname if x not in db_vps_name]
                diff_num = len(diff)
                diff = '\n'.join(diff)
                db_provider.tiktok(provider,diff_num,diff,chooise=0)

    def Encapsulation(provider,function_name):
        name = function_name
        # print(name)
        db_vps_name,db_vps_num = db_provider.db(provider)
        vpsname, vps_num = getattr(getattr(db_provider, 'provider'), name)()
        db_provider.diff_provider(db_vps_name,db_vps_num,vpsname,vps_num,provider)
        # Encapsulation 函数返回一个 Result 对象,包含 provider,vpsname和vps_num 三个属性
        return Result(provider, vpsname, vps_num)

    def tiktok(provider,diff_num,diff,chooise):
        # 机器人API接口
        api_url = "https://open.feishu.cn/open-apis/bot/v2/hook/94175274-25ea-4bf6-bb16-ac0c681555bb"
        # 获取时间:
        now = str(time.strftime("%Y年-%m月-%d日 %H:%M:%S", time.localtime()))
        # 伪造请求头
        headers = {
            "Content-Type": "application/json; charset=utf-8",
        }
        # 推送信息
        payload_message1 = {
            "msg_type": "post",
            "content": {
                "post": {
                    "zh_cn": {
                        "title": "DPS信息检测"+now,
                        "content": [
                                [
                                    {
                                        "tag": "text",
                                        # "text": f"{provider}存在{diff_num}台DPS信息有误\n以下机器信息填写错误\n{diff}",
                                        # Python3.6以上版本才支持f-string语法格式
                                        "text":provider + "存在" + str(diff_num) + "台信息有误\n以下机器信息填写错误\n" + diff,
                                        # "style": ["bold", "underline"]
                                    }
                                ]]}}}}
        payload_message2 = {
            "msg_type": "post",
            "content": {
                "post": {
                    "zh_cn": {
                        "title": "DPS信息检测"+now,
                        "content": [
                                [
                                    {
                                        "tag": "text",
                                        "text": provider+ "存在" + str(diff_num) + "台DPS信息缺失\n可能存在以下机器未装机,请检查\n" + diff
                                        # "text": provider + "存在" + diff_num "台信息缺失\n可能存在以下机器未装机,请检查\n" + diff
                                        # "style": ["bold", "underline"]
                                    }
                                ]]}}}}
        if chooise == 1:
            response = requests.post(url=api_url, data=json.dumps(payload_message1), headers=headers)
        else:
            response = requests.post(url=api_url, data=json.dumps(payload_message2), headers=headers)

provider_list = ["91vps","云立方3","263vps","有益网络","91搜云","必诚互联","云立方2"]
for provider in provider_list:
    if provider == "云立方2":
        function_name = "YunLiFang2"
        obj = db_provider.Encapsulation(provider, function_name)
    elif provider == "云立方3":
        function_name = "YunLiFang3"
        obj = db_provider.Encapsulation(provider, function_name)
    elif provider == "263vps":
        function_name = "hean_yun"
        obj = db_provider.Encapsulation(provider,function_name)
    elif provider == "91vps":
        function_name = "NineOne_vps"
        obj = db_provider.Encapsulation(provider, function_name)
    elif provider == "91搜云":
        function_name = "NineOneSouYun"
        obj = db_provider.Encapsulation(provider, function_name)
    elif provider == "必诚互联":
        function_name = "XiGuaip"
        obj = db_provider.Encapsulation(provider, function_name)
    elif provider == "有益网络":
        function_name = "YouYi_vps"
        obj = db_provider.Encapsulation(provider, function_name)
    else:
        pass