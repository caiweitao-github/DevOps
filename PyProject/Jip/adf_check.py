# -*- encoding: utf-8 -*-
from Jip.analysis import analyzer

enable_test = True

data_path = "/root/CKSHOW.csv"

count = 0
id = ''
# id = 'a032cc41-6f83-4d3c-b07e-3dc84e4bb71c'

from datetime import datetime, timedelta, date

# 获取当前时间的字符串表示
day = (date.today()).strftime("%Y-%m-%d %H:%M:%S")
# 获取 7 天前的时间字符串表示
day2 = (date.today() + timedelta(-7)).strftime("%Y-%m-%d %H:%M:%S")

# 将字符串时间转换为 datetime 对象
day_datetime = datetime.strptime(day, "%Y-%m-%d %H:%M:%S")
day2_datetime = datetime.strptime(day2, "%Y-%m-%d %H:%M:%S")

# 获取秒级时间戳
seconds_timestamp_day = int(day_datetime.timestamp())
seconds_timestamp_day2 = int(day2_datetime.timestamp())


a = analyzer(data_path, id, seconds_timestamp_day2, seconds_timestamp_day)
print("设备总数: ",len(a.all_id()))

out_path = '/root/results/'

for test_id in a.all_id():
    count = count + 1
    a.show_all_plt(test_id,out_path)
