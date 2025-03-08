# -*- encoding: utf-8 -*-

#在线率 = (总在线时长 / 86400) * 100

import pandas as pd
from datetime import datetime, timedelta, date

day_time_seconds = 86400

file_path = "/root/CKSHOW.csv"

day = 7

results = {}

def get_timestamps():
    today = datetime.now().replace(hour=0, minute=0, second=0, microsecond=0)
    timestamps = []
    for i in range(day+1):
        start_time = today - timedelta(days=i+1)
        end_time = today - timedelta(days=i)
        timestamps.append((int(start_time.timestamp()), int(end_time.timestamp())))
    return timestamps


def read_file():
    time_list = get_timestamps()
    column_names = ['ProxyIndex', 'Status', 'Time']
    data = pd.read_csv(file_path, names=column_names, header=None)
    # proxy_index = data["ProxyIndex"].unique()
    proxy_index = ['01a6241d-091c-4323-ae36-3d754f6c4dce']
    for index in proxy_index:
        for start,end in time_list:
            res = data.query("ProxyIndex == '%s' and Time >= %d and Time <= %d" %(index, start, end))
            if not res.empty:
                d = res.sort_values(by='Time')
                on, do = statistics(d, start, end)
                print("%d ~ %d %s 在线: %d, 掉线: %d" %(start, end, index, on, do))


def statistics(data, start_time, end_time):
    time = 0
    status = 0
    online_time = 0
    down_count = 0
    for index, row in enumerate(data.itertuples()):
        if status == 0:
            status = row.Status
            if status == 1:
                time = row.Time
                continue
            else:
                online_time = online_time + (row.Time - start_time)
                continue

        if row.Status == 1:
            if status == 1:
                if time == 0:
                    time = row.Time
                else:
                    online_time = online_time + (row.Time - time)
                    time = row.Time
            else:
                time = row.Time
            if index == len(data):
                online_time = online_time + (end_time - row.Time)
        else:
            down_count += 1
            if status == 1:
                online_time = online_time + (row.Time - time)
            else:
                time = 0
        status = row.Status
    return online_time, down_count


if __name__ == '__main__':
    read_file()