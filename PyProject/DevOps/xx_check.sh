#!/bin/bash
# 每小时检测一次小熊ip的余额，3分钟版小于100w，10分钟版小于30w告警
# 运行机器：n9e-server
# 执行方式：crontab：0 * * * * bash /root/bin/xx_check.sh

# 获取当前时间
dt=$(date '+%Y-%m-%d %T')

# 3分钟版
# 获取ip余额
ip_num_3min=$(curl -s "https://u.ipapiaddress.com/api/remainIpCount/126e0b8c249145de" | awk -F':' '{print $4}' | sed 's/}//g')
# 记录检测时间和检测时的ip数量到日志
echo "[$dt] 3分钟版当前可用ip数量为：$ip_num_3min" >> /root/log/xx_check.log
# 判断ip数量是否小于100w，小于100万则告警
if [[ $ip_num_3min -lt 5000 ]]; then
   # 记录检测时间和检测时的ip数量到日志
   #echo "[$dt] 3分钟版当前可用ip数量为：$ip_num_3min" >> /root/log/xx_check.log
    Time=$(date '+%T')
    prompt="当前剩余ip数量为：$ip_num_3min，请及时补充"
    curl -X POST -H "Content-Type: application/json"\
    -d '{"msg_type":"post","content":{"post":{"zh_cn":{"title":"【小熊3分钟版告警】 '$Time'","content":[[{"tag":"text","text":"'$prompt'"},{"tag": "at","user_id": "57d87g78","user_name": "王景>高（jerrywang）"}]]}}}}'\
    https://open.feishu.cn/open-apis/bot/v2/hook/3f3de577-5a27-4bb1-864f-eac4792ea12b >/dev/null 2>&1
fi

# 5分钟版
# 获取ip余额
ip_num_5min=$(curl -s "https://u.ipapiaddress.com/api/remainIpCount/e0ec385d9f15f816" | awk -F':' '{print $4}' | sed 's/}//g')
# 记录检测时间和检测时的ip数量到日志
echo "[$dt] 5分钟版当前可用ip数量为：$ip_num_5min" >> /root/log/xx_check.log
# 判断ip数量是否小于30w，小于30万则告警
if [[ $ip_num_5min -lt 300000 ]]; then
    # 记录检测时间和检测时的ip数量到日志
    #echo "[$dt] 5分钟版当前可用ip数量为：$ip_num_5min" >> /root/log/xx_check.log
    Time=$(date '+%T')
    prompt="当前剩余ip数量为：$ip_num_5min，请及时补充"
    curl -X POST -H "Content-Type: application/json" \
    -d '{"msg_type":"post","content":{"post":{"zh_cn":{"title":"【小熊5分钟版告警】'$Time'","content":[[{"tag":"text","text":"'$prompt'"},{"tag": "at","user_id": "57d87g78","user_name": ">王景>高（jerrywang）
"}]]}}}}' \
    https://open.feishu.cn/open-apis/bot/v2/hook/3f3de577-5a27-4bb1-864f-eac4792ea12b >/dev/null 2>&1
fi

# 10分钟版
# 获取ip余额
ip_num_10min=$(curl -s "https://u.ipapiaddress.com/api/remainIpCount/b4453c393cc7f6fd" | awk -F':' '{print $4}' | sed 's/}//g')
# 记录检测时间和检测时的ip数量到日志
echo "[$dt] 10分钟版当前可用ip数量为：$ip_num_10min" >> /root/log/xx_check.log
# 判断ip数量是否小于30w，小于30万则告警
if [[ $ip_num_10min -lt 20000 ]]; then                                               
    # 记录检测时间和检测时的ip数量到日志
    #echo "[$dt] 10分钟版当前可用ip数量为：$ip_num_10min" >> /root/log/xx_check.log
    Time=$(date '+%T')
    prompt="当前剩余ip数量为：$ip_num_10min，请及时补充"
    curl -X POST -H "Content-Type: application/json" \
    -d '{"msg_type":"post","content":{"post":{"zh_cn":{"title":"【小熊10分钟版告警】'$Time'","content":[[{"tag":"text","text":"'$prompt'"},{"tag": "at","user_id": "57d87g78","user_name": ">王景
高（jerrywang）"}]]}}}}' \
    https://open.feishu.cn/open-apis/bot/v2/hook/3f3de577-5a27-4bb1-864f-eac4792ea12b >/dev/null 2>&1
fi