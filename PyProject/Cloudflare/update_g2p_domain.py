# -*- encoding: utf-8 -*-
import requests
import pymysql

config = {
    'user': '',
    'password': '',
    'host': '',
    'database': '',
    'autocommit': True
}

connection = pymysql.connect(**config)
go2proxy_db = connection.cursor()

src_code = ""
dest_code = ""


class CloudflareAPI():
    def __init__(self):
        self.domain = 'go2proxy.com'
        self.api_url = 'https://api.cloudflare.com/client/v4/zones/'
        self.auth_key = ''
        self.auth_email = ''
        self.global_head = {'Content-Type': 'application/json', 'X-Auth-Key': self.auth_key,
                            'X-Auth-Email': self.auth_email}

    def get_zone_id(self):
        r = requests.get(self.api_url, headers=self.global_head)
        if r.status_code == 200:
            data = r.json()
            for i in data['result']:
                if i['name'] == self.domain:
                    return i['id']
        else:
            err = r.json()
            print("code: %s, error: %s" % (err['errors'][0]['code'], err['errors'][0]['message']))

    def run(self):
        src_id, src_ip, src_location = self.get_src_ip()
        dest_id, dest_ip, dest_location = self.get_dest_ip()
        if all((src_id, src_ip, src_location, dest_id, dest_ip, dest_location)):
            if src_location == dest_location:
                zone_id = self.get_zone_id()
                domain_list = self.get_domain_list(src_location, src_id)
                for domains in domain_list:
                    if src_location == 'as':
                        domain_id = self.query_domain(domains[1], zone_id)
                        if domain_id:
                            res = self.update_domain_record(zone_id, domain_id, dest_ip)
                            if res:
                                self.update_db(dest_id, domains, dest_location)
                                self.inser_change_record(domains, src_id, dest_id)
                                print("update %s -> %s" %(domains[1], dest_ip))
                            else:
                                print("update %s err" %(domains[1], dest_ip))
                        else:
                            print("get domain id err")
                    elif src_location == 'us':
                        flage = True
                        for i in domains:
                            domain_id = self.query_domain(i, zone_id)
                            if domain_id:
                                res = self.update_domain_record(zone_id, domain_id, dest_ip)
                                if res:
                                    print("update %s -> %s" %(domains, dest_ip))
                                else:
                                    flage = False
                                    print("update %s err" %(domains[0], dest_ip))
                            else:
                                print("get domain id err")
                        if flage:
                            self.update_db(dest_id, domains, dest_location)
                            self.inser_change_record(domains, src_id, dest_id)
            else:
                print("src location or dest location err")
        else:
            print("get data err")

    def query_domain(self, domain_str, zone_id):
        url = self.api_url + zone_id + '/dns_records'
        p = {'name': domain_str, 'type': 'A'}
        r = requests.get(url, headers=self.global_head, params=p)
        if r.status_code == 200:            
            data = r.json()
            return data['result']['id']
        else:
            print("get domain data err %s" %(r.json()))
            return None

    def update_domain_record(self, zone_id, domain_id, ip):
        url = self.api_url + zone_id + '/dns_records/' + domain_id
        payload = {
            'content': ip,
        }
        r = requests.patch(url, headers=self.global_head, json=payload)
        if r.status_code == 200:
            return True
        else:
            return False

    def get_src_ip(self):
        sql = "select id,login_ip,location_code from fps where code = '%s'" %(src_code)
        go2proxy_db.execute(sql)
        return go2proxy_db.fetchone()

    def get_dest_ip(self):
        sql = "select id,login_ip,location_code from fps where code = '%s'" %(dest_code)
        go2proxy_db.execute(sql)
        return go2proxy_db.fetchone()

    def get_domain_list(self, location, id):
        sql = "select domain,sub_domain_%s from fps_domain where fps_%s_id = '%s'" %(location, location, id)
        go2proxy_db.execute(sql)
        return go2proxy_db.fetchall()

    def update_db(self, id, domain, location):
        sql = "update fps_domain set fps_%s_id = '%s' where domain = '%s' or sub_domain_%s = '%s'" %(location, id, domain[0], domain[1], location)
        go2proxy_db.execute(sql)

    def inser_change_record(self, domains, src_id, dest_id):
        sql = """insert into fps_domain_change(domain, sub_domain, status, update_time, create_time, memo, dest_fps_id, src_fps_id) 
        values('%s', '%s', 1, CURRENT_TIMESTAMP(), CURRENT_TIMESTAMP(), "", %d, %d)""" % (
        domains[0], domains[1], dest_id, src_id)
        go2proxy_db.execute(sql)



if __name__ == '__main__':
    c = CloudflareAPI()
    c.run()
