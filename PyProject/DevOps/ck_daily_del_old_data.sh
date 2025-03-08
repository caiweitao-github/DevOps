#!/bin/sh

#定期删除ck里的老数据，释放磁盘空间

#main13 nodeops.node_request_history&tpsstat.tunnel_request_history表
#node_request_history保留30天的数据，tunnel_request_history保留60天的数据
nyesterday=`date +"%Y%m%d" -d "-30day"`;
tyesterday=`date +"%Y%m%d" -d "-50day"`;
clickhouse-client -u kdl --password yvk8fcfb -h 10.0.6.10 -d nodeops --query="ALTER TABLE node_request_history DETACH PARTITION $nyesterday;"
clickhouse-client -u kdl --password yvk8fcfb -h 10.0.6.10 -d tpsstat --query="ALTER TABLE tunnel_request_history DETACH PARTITION $tyesterday;"
echo "`date +'%Y-%m-%d %H:%M:%S'` - [nodeops.node_request_history] detach $nyesterday"
echo "`date +'%Y-%m-%d %H:%M:%S'` - [tpsstat.tunnel_request_history] detach $tyesterday"
#detach目录下的内容由root用户的cron任务每天删除
#######
