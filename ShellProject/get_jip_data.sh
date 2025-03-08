#!/bin/bash

query_num=${1:-500}
file_name="CKSHOW.csv"

kdlstat="数据库连接地址"

$kdlstat "select proxy_index,sum(down_time) as sum_down_time from jip_node_down_history where version = '1.4' and 
down_date between DATE_FORMAT(DATE_SUB(CURRENT_DATE(),INTERVAL 1 DAY),'%Y-%m-%d 00:00:00') and DATE_FORMAT(CURRENT_DATE(),'%Y-%m-%d 00:00:00') 
group by proxy_index having sum_down_time>=$query_num order by sum_down_time desc" > jip_index.txt

> $file_name

while read line;do
    proxy_index=$(echo $line|awk '{print $1}')
    clickhouse-client -u kdl --password yvk8fcfb -h 10.0.6.10 -d eipstat --query "select proxy_index,action,toUnixTimestamp(timestamp) as timestamp from jip_event_history where timestamp  
    BETWEEN toUnixTimestamp(toDateTime(toStartOfDay(toDateTime(now()) - interval 7 day))) AND toUnixTimestamp(toStartOfDay(toDateTime(now()))) and 
    proxy_index = '$proxy_index' FORMAT CSV" >> $file_name
done < jip_index.txt

rm -rf data.tgz

tar -zcf data.tgz $file_name