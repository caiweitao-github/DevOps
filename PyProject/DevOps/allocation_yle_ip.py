# -*- coding: utf-8 -*-
'''
分配完成后可以有这个脚本润色一下
'''
import sys
reload(sys)
sys.setdefaultencoding('utf-8')

import random
from area_code_dict import area_dict,prov_dict
from redisutil import redisdb_crs,redisdb
import dbutil

db_kdl = dbutil.get_db_kdlnode()

regions = [
    "四川", "重庆", "陕西", "甘肃", "青海", "宁夏", "新疆", 
    "安徽", "云南", "辽宁", "吉林", "黑龙江", "海南", "广东", 
    "广西", "湖北", "湖南", "河南", "北京", "河北", "天津", 
    "内蒙古", "山西", "浙江", "西藏", "上海", "山东", 
    "江西", "福建", "贵州", "江苏"
]

data = {
    "四川": "2.36%",
    "重庆": "1.25%",
    "陕西": "2.32%",
    "甘肃": "0.77%",
    "青海": "0.18%",
    "宁夏": "0.19%",
    "新疆": "0.81%",
    "安徽": "3.95%",
    "云南": "1.29%",
    "辽宁": "8.98%",
    "吉林": "4.21%",
    "黑龙江": "6.23%",
    "海南": "0.38%",
    "广东": "5.76%",
    "广西": "2.24%",
    "湖北": "3.24%",
    "湖南": "5.46%",
    "河南": "13.09%",
    "北京": "4.16%",
    "河北": "14.44%",
    "天津": "1.14%",
    "内蒙古": "2.38%",
    "山西": "3.52%",
    "浙江": "1.42%",
    "西藏": "0.09%",
    "上海": "2.30%",
    "山东": "12.03%",
    "江西": "1.76%",
    "福建": "2.27%",
    "贵州": "1.42%",
    "江苏": "2.56%",
}

def get_mysql_count(province_code):
    sql = """select count(*) from yle_node where province_code = "%s";""" % (province_code)
    result = db_kdl.execute(sql).fetchone()
    return result[0]

def get_province_count():
    province_count = {}
    provinces = data.keys()
    for province in provinces:
        province_code = prov_dict.get(province)
        count = get_mysql_count(province_code)
        province_count[province] = count
        # print("%s %s"%(province,count))

    return province_count

def difference_num(provinces_negative,provinces_positive):
    # 拿到差值后，分配数量
    for province in provinces_negative.keys():
        difference = abs(provinces_negative[province])

        #随机拿到一个正省份
        positive_province = random.choice(list(provinces_positive.keys()))

        positive = provinces_positive[positive_province]
        num = int(positive) - difference

        province_code = prov_dict.get(province) # 多了的省份code
        positive_province_code = prov_dict.get(positive_province)

        if num == 0:
            list_num = abs(difference)
            sql = '''UPDATE yle_node SET province_code = '%s' 
            WHERE id IN (SELECT id FROM (SELECT id FROM yle_node WHERE province_code = '%s' ORDER BY RAND() LIMIT %s ) AS temp_table);''' %(positive_province_code,province_code,list_num)
            del provinces_positive[positive_province]
            del provinces_negative[province]
            print(sql)
        elif num > 0:
            list_num = abs(difference)
            sql = '''UPDATE yle_node SET province_code = '%s' 
            WHERE id IN (SELECT id FROM (SELECT id FROM yle_node WHERE province_code = '%s' ORDER BY RAND() LIMIT %s ) AS temp_table);''' %(positive_province_code,province_code,list_num)
            provinces_positive[positive_province] = num
            del provinces_negative[province]
            print(sql)
        else:
            sql = '''UPDATE yle_node SET province_code = '%s' 
            WHERE id IN (SELECT id FROM (SELECT id FROM yle_node WHERE province_code = '%s' ORDER BY RAND() LIMIT %s ) AS temp_table);''' %(positive_province_code,province_code,positive)
            del provinces_positive[positive_province]
            provinces_negative[province] = num
            print(sql)

def node_configuration(num):
    provinces = data.keys()
    old_province_count = get_province_count()
    provinces_negative = {}
    provinces_positive = {}
    for province in provinces:
        province_proportion = data[province]
        province_percentage = float(province_proportion.replace('%', '')) / 100
        province_num = int(num * province_percentage)
        print("%s %s"%(province,province_num))
        old_count = old_province_count[province]

        count = province_num - old_count
        if count > 0:
            provinces_positive[province] = count
        elif count < 0:
            provinces_negative[province] = count
        else:
            continue

    # 检查是否还有遗漏的
    while True:
        if len(provinces_negative) > 0:
            if len(provinces_positive) > 0:
                difference_num(provinces_negative,provinces_positive)
            else:
                print(provinces_negative)
                break
        else:
            print('分配完成')
            break

if __name__ == '__main__':
    num = 2000
    node_configuration(num)
