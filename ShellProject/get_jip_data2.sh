#!/bin/bash

file_name="CKSHOW.csv"

kdlstat="数据库连接地址"

$kdlstat "select proxy_index from jip_node where status = 1 limit 1000" > jip_index.txt

echo '' > $file_name

while read line;do
    proxy_index=$(echo $line|awk '{print $1}')
    clickhouse-client -u kdl --password yvk8fcfb -h 10.0.6.10 -d eipstat --query "select proxy_index,action,toUnixTimestamp(timestamp) as timestamp from jip_event_history where timestamp
    BETWEEN toUnixTimestamp(toDateTime(toStartOfDay(toDateTime(now()) - interval 3 day))) AND toUnixTimestamp(toDateTime(toStartOfDay(toDateTime(now()) - interval 1 day))) and
    proxy_index = '$proxy_index' FORMAT CSV" >> $file_name
done < jip_index.txt

rm -rf data.tgz

tar -zcf data.tgz $file_name