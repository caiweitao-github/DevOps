# -*- encoding: utf-8 -*-
"""
定期检查可用域名数量是否低于阈值
低于阈值则会创建一定数量的域名
"""
import pymysql
import requests
import random
import string

domain_num = 5
create_num = 5

config = {
    'user': '',
    'password': '',
    'host': '',
    'database': '',
    'autocommit': True
}

connection = pymysql.connect(**config)
go2proxy_db = connection.cursor()

class CloudflareAPI():
    def __init__(self):
        self.domain = 'go2proxy.com'
        self.base_domain = '.go2proxy.com'
        self.us_domain_suffix = '.us' + self.base_domain
        self.as_domain_suffix = '.as' + self.base_domain
        self.master_domain_prefix = '.g' + self.base_domain
        self.api_url = 'https://api.cloudflare.com/client/v4/zones/'
        self.auth_key = ''
        self.auth_email = ''
        self.global_head = {'Content-Type': 'application/json', 'X-Auth-Key': self.auth_key, 'X-Auth-Email': self.auth_email}

    def get_zone_id(self):
        r = requests.get(self.api_url, headers=self.global_head)
        if r.status_code == 200:
            data = r.json()
            for i in data['result']:
                if i['name'] == self.domain:
                    return i['id']
        else:
            err = r.json()
            print("code: %s, error: %s" %(err['errors'][0]['code'], err['errors'][0]['message']))

    def creat_domain_name(self):
        zone_id = self.get_zone_id()
        url = self.api_url + zone_id + '/dns_records'
        count = 0
        while 1:
            if count == create_num:
                break
            domain_list = self.create_domain_str()
            is_ex = self.query_doamin(url, domain_list[0])
            if not is_ex:
                us_node = self.get_node('us')
                as_node = self.get_node('as')
                domain_data = {
                    'domain': [d for d in us_node] + [domain_list[0]],  #数据格式[fps_id, fps_code, fps, fps_ip, domain]
                    'domain_us': [d for d in us_node] + [domain_list[1]],
                    'domain_as': [d for d in as_node] + [domain_list[2]]
                }
                for domain in domain_data.values():
                    payload = {
                        'content': domain[2],
                        'name': domain[-1],
                        'proxied': False,   #代理状态，一般不修改
                        'type': 'A',
                        # 'comment': '',  #备注
                        'ttl': 600
                    }
                    print(payload)
                    # r = requests.post(url, headers=self.global_head, json=payload)
                    # if r.status_code == 200:
                    #     print("create %s ---> %s(%s) success!" %(domain[-1], domain[2], domain[1]))
                    # else:
                    #     err = r.json()
                    #     print("code: %s, error: %s" %(err['errors'][0]['code'], err['errors'][0]['message']))
                    #     return
                count+=1
                self.inser_into_db(domain_data['domain'][-1], domain_data['domain_as'][-1], domain_data['domain_us'][-1], domain_data['domain_as'][0], domain_data['domain_us'][0])
            else:
                print("domain %s is exist, skip!" %(domain))
                continue
    
    def create_domain_str(self):
        charset = string.ascii_lowercase + string.digits
        random_chars = ''.join(random.choice(charset) for _ in range(10))
        return [random_chars + self.master_domain_prefix, random_chars + self.us_domain_suffix, random_chars + self.as_domain_suffix]

    def query_doamin(self, url, domain_str):
        p = {'name': domain_str, 'type': 'A'}
        r = requests.get(url, headers=self.global_head, params=p)
        res = r.json()
        return bool(res['result'])
    
    def inser_into_db(self, domain, as_domain, us_domain, as_id, us_id):
        sql = """insert into fps_domain(domain, sub_domain_as, sub_domain_us, status, update_time, create_time, memo, fps_as_id, fps_us_id) 
        values('%s', '%s', '%s', 1, CURRENT_TIMESTAMP(), CURRENT_TIMESTAMP(), "", %d, %d)""" %(domain, as_domain, us_domain, as_id, us_id)
        print(sql)
        # go2proxy_db.execute(sql)
    
    def get_node(self, location_code):
        sql = """select id,code,login_ip from fps where location_code = '%s' """ %(location_code)
        go2proxy_db.execute(sql)
        r = go2proxy_db.fetchall()
        return random.choice(list(r))

def get_domain_num():
    """
    检查可用域名的总数，而不是检查单个机器的可用域名数量
    """
    sql = """select count(*) as num from fps_domain where status = 1"""   #并没有做带宽筛选，后期有需要可以加上。
    go2proxy_db.execute(sql)
    r = go2proxy_db.fetchone()
    return int(r[0]) < domain_num
        

if __name__ == '__main__':
    is_create = get_domain_num()
    if is_create:
        Cloudflare = CloudflareAPI()
        Cloudflare.creat_domain_name()