#!/bin/bash
#执行方式：crontab */10 * * * * bash /root/bin/check_sproxy.sh
#检查redis哈希表中status占比低于50%，就kill掉柚雷的sproxyserver进程

# 设置百分比值
percent=50

# 获取tdps:yle_proxy哈希表中的字段数量
total_fields=$(redis-cli -h localhost -p 6379 -n 1 HGETALL "tdps:yle_proxy" | awk 'NR % 2 == 0' | wc -l)

# 获取status等于1的数量
status_1_count=$(redis-cli -h localhost -p 6379 -n 1 HGETALL "tdps:yle_proxy" | awk 'NR % 2 == 0' | jq -r '.status' | grep '^1$' | wc -l)

# 计算status等于1的数量占总字段数量的百分比
percentage=$(echo "scale=2; $status_1_count / $total_fields * 100" | bc)

# 如果status等于1的数量占总字段数量的百分比低于50%，则重启进程
if (( $(echo "$percentage < $percent" | bc -l) )); then
    CURRENT_TIME=$(date +"%Y-%m-%d %H:%M:%S")
    echo "$CURRENT_TIME 正常占比为$percentage%" >> /root/log/check_sproxy.log
    # 终止 /opt/eproxy/bin/sproxyserver 进程
    pkill -f /opt/eproxy/bin/sproxyserver && echo "sproxyserver进程已kill重启" >> /root/log/check_sproxy.log
fi