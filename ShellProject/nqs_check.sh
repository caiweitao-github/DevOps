#!/bin/bash

echo "$(date +%F-%H:%M:%S) - run..." >> /root/bin/nqs_check.log

cpu_num=$(nproc)

node_load=$(uptime |awk '{print $10}'|awk -F ',' '{print $1}'|awk -F '.' '{print $1}')

threshold=$(awk -v n1="$cpu_num" 'BEGIN{printf (n1-1)}')

data_time=$(date +%F-%H:%M:%S)

pid=$(top -cbn1 | grep 'tps_nqs_gun.conf' | sort -k9nr | head -n 1|awk '{print $1}')

if [[ $node_load -ge $threshold ]];then
    py-spy record -o /data/log/tps/nqs_cpu.$(date +%F-%H:%M:%S).svg --pid $pid --duration 60
fi