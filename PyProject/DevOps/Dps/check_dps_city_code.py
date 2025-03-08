#!/usr/bin/python
# -*- encoding: utf-8 -*-
"""
主要功能:定期检测dps表中province_code,city_code为空的正常dps,并将它们按照location进行填写
运行方式:计划任务运行,每天运行一次
"""
import sys # 导入 sys 模块，用于配置默认字符编码
reload(sys) # 导入 sys 模块，用于配置默认字符编码
sys.setdefaultencoding("utf8")  # 设置默认字符编码为 UTF-8，以确保处理 Unicode 数据时不会出现编码问题
import dbutil;kdl_db = dbutil.get_db_db()
import loggerutil;logger = loggerutil.get_logger('check_province_city_code','dps_master/check_dps_info.log')
from areautil import get_jde_province_code_by_name

"""根据location拆分省+市"""
def Split_location(location):
    if location in ["北京市","重庆市","上海市","天津市"]:
        province_name = location.split('市')[0]
        city_name = location.decode('utf-8')
    else:
        province_name = location.split("省")[0]
        city_name = location.split("省")[1].split("市")[0]
    info_dict =  get_jde_province_code_by_name(province_name,city_name)
    if info_dict is None:
        logger.warning("No provincial and municipal information was found at this address. Please check whether the address information is correct.--->%s"%location)
        return None
    return info_dict
"""查询province_code,city_code为空的dps"""
def get_dps_code():
    get_dps_code_sql = "select code,location from dps where status !=4 and location!=' ' and province_code = ' ' and city_code = ' ';"
    result = kdl_db.execute(get_dps_code_sql).fetchall()
    for row in result:
        code = row[0]
        location = row[1]
        info_dict = Split_location(location)
        if info_dict is None:
            continue
        else:
            province_code = info_dict['prov_code']
            city_code = info_dict['city_code']
            update_sql = "update dps set province_code='%s',city_code='%s' where code='%s' and location='%s';"%(province_code,city_code,code,location)
            try:
                kdl_db.update(update_sql)
                logger.info("%s的位置信息为:%s,省编码为:%s,市编码为:%s,现已修复省市编码为空的情况"%(code,location,province_code,city_code))
            except Exception as e:
                logger.error(e)
"""查询dps表中location与省市编码不符的bug"""
def get_dps_location_code_bug():
    get_dps_location = "select location from dps where status!=4 and location !=' ' group by location;"
    result = kdl_db.execute(get_dps_location).fetchall()
    for row in result:
        location = row[0]
        info_dict = Split_location(location)
        if info_dict is None:
            continue
        else:
            province_code = info_dict['prov_code']
            city_code = info_dict['city_code']
            get_error_city_code_sql = "select code,city_code from dps where status!=4 and location='%s' and city_code!='%s';"%(location,city_code)
            result = kdl_db.execute(get_error_city_code_sql).fetchall()
            for row in result:
                code = row[0]
                error_city_code = row[1]
                update_sql = "update dps set province_code='%s',city_code='%s' where code='%s';"%(province_code,city_code,code)
                try:
                    kdl_db.update(update_sql)
                    logger.info("%s的位置信息为:%s,省编码为:%s,市编码为:%s,原错误市编码为:%s,现已修复数据库中位置信息与省市编码不符的情况"%(code,location,province_code,city_code,error_city_code))
                except Exception as e:
                    logger.error(e)
"""定义主函数"""
def main():
    get_dps_code()
    get_dps_location_code_bug()
if __name__ == '__main__':
    main()
