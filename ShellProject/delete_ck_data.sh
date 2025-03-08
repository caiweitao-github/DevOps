#!/bin/bash

# jyesterday=`date +"%Y%m%d" -d "-30day"`

for i in $(seq 50 440);do
    fyesterday=`date +"%Y%m%d" -d "-${i}day"`;
    # tmp=$(expr $i + 1)
    # time2=`date +"%Y%m%d" -d "-${tmp}day"`
    # echo $fyesterday
    clickhouse-client -u kdl --password yvk8fcfb -h 10.0.6.10 -d fpsstat --query="ALTER TABLE fps_request_history DETACH PARTITION ($fyesterday, $fyesterday);"
done