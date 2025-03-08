# -*- coding: utf-8 -*-
# 导入common的dbutil 连接db数据库
import dbutil
import datetime
# date用来更新dps_group_relation的update_time字段
date = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
db_db = dbutil.get_db_db()

# 通过dps_1和dps_2两个code 获取两个dps的dps_id和换IP周期changeip_period
def as_dpscode_get_dpsinfo(dps_1_code,dps_2_code):
	sql="select id,changeip_period from dps where code='%s' or code='%s';" %(dps_1_code,dps_2_code) 
	cursor = db_db.execute(sql)
	#result = cursor.fetchall()
	#print result
	dps_1=cursor.fetchone()
	dps_2 = cursor.fetchone()
	#print dps_2
	return dps_1,dps_2

# 通过dpsid更换组关系dgr
# 保持dgr的 id-dpsid-groupid里 id和groupid对应关系不改变 只改变dpsid
def as_dpsid_update_dgr_group(dps_1_id,dps_2_id):
	sql="UPDATE dps_group_relation SET dps_id='%s',update_time='%s' where dps_id='%s';" %(dps_1_id,date,dps_2_id)
	db_db.execute(sql)
	sql="UPDATE dps_group_relation SET dps_id='%s',update_time='%s' where dps_id='%s' and not update_time='%s';" %(
	dps_2_id,date,dps_1_id,date)
	db_db.execute(sql)
# 通过dpsid更换换ip周期
def as_dpsid_update_changeip_period(dps_1_id,dps_1_changeip_period,dps_2_id,dps_2_changeip_period):
	if dps_1_changeip_period==dps_2_changeip_period:
		print "the same pass"
	else:
		sql="UPDATE dps SET changeip_period='%s' WHERE id='%s';" %(dps_2_changeip_period,dps_1_id)
		db_db.execute(sql)
		sql="UPDATE dps SET changeip_period='%s' WHERE id='%s';" %(dps_1_changeip_period,dps_2_id)
		db_db.execute(sql)
# 主函数
def exchange_dps(dps_1_code,dps_2_code):
	dps_1,dps_2 = as_dpscode_get_dpsinfo(dps_1_code,dps_2_code)
	
	dps_1_id,dps_1_changeip_period = dps_1[0],dps_1[1]
	dps_2_id,dps_2_changeip_period = dps_2[0],dps_2[1]
	as_dpsid_update_dgr_group(dps_1_id,dps_2_id)
	as_dpsid_update_changeip_period(dps_1_id,dps_1_changeip_period,dps_2_id,dps_2_changeip_period)
	#pass

#test
#exchange_dps("dps1400","dps2700")
