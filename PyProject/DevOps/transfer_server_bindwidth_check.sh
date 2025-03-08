#!/bin/bash
LOG_FILE="/root/log/proxy.log"
HOSTNAME=$(hostname)
function report(){
    while IFS= read -r line;do
        ip=$(echo "$line" | awk -F',' '{print $1}')
        in_bandwidth=$(echo "$line" | awk -F',' '{print $2}')
        out_bandwidth=$(echo "$line" | awk -F',' '{print $3}')
        order_info=$(echo "$line" | awk -F',' '{print $4}')
        line_count=$(echo "$line" | awk -F',' '{print $5}')
        bandwidth_MB=$(echo "$line" | awk -F',' '{print $6}')
        request_mean=$(echo "$line" | awk -F',' '{print $7}')
        format_output+="$(printf "客户端IP:%-5s, 订单信息:%-5s, 入带宽:%-5s, 出带宽:%-5s, 总消耗带宽: %-2s, 请求次数: %-2s, 请求均值: %-2s" "$ip" "$order_info" "$in_bandwidth" "$out_bandwidth" "$bandwidth_MB" "$line_count" "$request_mean")\n"
    done < ./iftop_network_content.txt
    content=$(echo "$format_output")
    echo "$content"
    # Feishu webhook URL  
    WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/eac844f1-6234-43ad-b0d0-0d226f30aa26"  
    # Local hostname and time  
    HOSTNAME=$(hostname)
    LOCAL_TIME=$(date +"%Y-%m-%d %H:%M:%S")
    ALERT_HEADERS="柚雷服务器带宽预警查询[主机名:$HOSTNAME]"
    MESSAGE_BODY='{
                    "msg_type": "post",
                    "content": {
                        "post": {
                        "zh_cn": {
                            "title": "'$ALERT_HEADERS'",
                            "content": [
                            [
                                {
                                    "tag": "text",
                                    "text": "'$content'"
                                },
                                {
                                    "tag": "text",
                                    "text": "查询时间为:'$LOCAL_TIME'"
                                }
                            ]
                        ]
                    }
                }
            }
        }'
    curl -X POST -H "Content-Type: application/json" -d "$MESSAGE_BODY" "$WEBHOOK_URL"   
}



function get_iftop_network_ip() {
    # 获取iftop网络流量信息
    iftop -t -s 1 -n -F 'IPs' 2>/dev/null | sed -n '4,23p' > ./iftop_network_ip.txt
}

function get_network_traffic(){
    # 网络带宽单位转换
    num=$(echo $1 | awk -F 'Kb' '{print $1}')
    if [[ "$1" == *'Kb' ]];then
        mb_value=$(echo "scale=4; $num / 1000" | bc)
        printf "%.2fMb" $mb_value
    else
        echo $1
    fi
}

function get_order_info(){
    # 获取远程主机对应的order_info
    order_info=$(cat $LOG_FILE | grep "$1" | grep 'public' | grep '|' | head -n 1 | awk -F '|' '{print $5}')
    if [[ -z $order_info ]];then
        echo "未找到有关订单信息"
    fi
    echo $order_info
}

function Decryption_order_info(){
    # 解密订单信息获取订单号或用户id
    # 定义加密字典
    declare -A encrypt_dict
    encrypt_dict=([1]="0" [9]="1" [5]="2" [6]="3" [8]="4" [7]="5" [2]="6" [4]="7" [0]="8" ["b"]="9")

    # 解密order_id函数
    local order_id=$1
    local result=""

    # 遍历order_id中的每一个字符并解码
    for ((i=0; i<${#order_id}; i++)); do
        char="${order_id:$i:1}"
        result+="${encrypt_dict[$char]}"
    done

    if [[ $1 =~ ^t ]];then
        order_info="tunnel_id:$1"
    elif [[ $1 =~ ^b ]];then
        order_info="order_id:$result"
    else
        order_info="user_id:$result"
    fi
    echo $order_info

}

function get_use_bandwidth(){
    # 获取当前时间
    current_time=$(date +"%Y-%m-%d %H:%M:%S")
    # 获取当前时间的前15分钟
    current_time_30min=$(date -d "5 minutes ago" +"%Y-%m-%d %H:%M:%S")
    # 获取日志文件的开始时间
    current_log_time=$(head -n 1 $LOG_FILE | awk -F '.' '{print $1}')

    # 将时间字符串转换为 Unix 时间戳
    current_time_unix=$(date -d "$current_time" +%s)
    current_time_30min_unix=$(date -d "$current_time_30min" +%s)
    current_log_time_unix=$(date -d "$current_log_time" +%s)

    if [ $current_log_time_unix -lt $current_time_30min_unix ]; then
        # 如果日志文件的开始时间早于当前时间的30分钟
        Start_Time=$current_log_time
    else
        Start_Time=$current_time_30min
    fi

    # 定义在目标时间段内某个IP的行数
    line_count=$(awk -v start="$Start_Time" -v end="$current_time" '$1 " " $2 >= start && $1 " " $2 <= end' "$LOG_FILE" | grep "public" |grep "|" | grep "$1" | wc -l)
    # 打印数组中的结果
    # 根据用户和时间过滤日志
    bandwidth=$(awk -v start="$Start_Time" -v end="$current_time" '$1 " " $2 >= start && $1 " " $2 <= end' "$LOG_FILE" | grep "public" |grep "|" | grep "$1" | awk -F'|' '{if (NF >= 3) print $(NF-3)}')
    bandwidth_all=0
    for i in $bandwidth; do
        bandwidth_all=$((bandwidth_all + i))
    done
    if [ $bandwidth_all -ne 0 ]; then
        # 将带宽转换为MB
        bandwidth_MB=$(echo "scale=4; $bandwidth_all / (1024 * 1024)" | bc)
        # 请求均值
        request_mean=$(echo "scale=4; $bandwidth_MB / $line_count" | bc)
    else
        bandwidth_MB=0
        request_mean=0
    fi
    
    # 订单信息
    order_info=$(Decryption_order_info "$1")
    local values=(
        "$order_info"
        "$line_count"
        "$bandwidth_MB"
        "$request_mean"
    )
    echo "${values[@]}"
}


function get_network_ip() {
   # 读取iftop_network_ip.txt文件获取网络流量信息,并进行带宽单位转换
   while IFS= read -r line1 && IFS= read -r line2
   do
        # iftop监控本机主机IP
        localPublic_IpAddress=$(echo $line1 | awk '{print $2}')
        # iftop监控本机IP出口带宽
        out_Bandwidth=$(get_network_traffic $(echo $line1 | awk '{print $4}'))
        # iftop监控远程主机IP
        remotePublic_IpAddress=$(echo $line2 | awk '{print $1}')
        # iftop监控本机IP入口带宽
        in_Bandwidth=$(get_network_traffic $(echo $line2 | awk '{print $4}'))
        data_values=($(get_use_bandwidth "$(get_order_info $remotePublic_IpAddress)"))
        order_info="${data_values[0]}"
        line_count="${data_values[1]}"
        bandwidth_MB="${data_values[2]}"
        request_mean="${data_values[3]}"
        printf "%s, %s, %s, %s, %s, %.2fMb, %.2fMb\n" "$remotePublic_IpAddress" "$in_Bandwidth" "$out_Bandwidth" "$order_info" "$line_count" "$bandwidth_MB" "$request_mean" >>  ./iftop_network_content.txt
   done < ./iftop_network_ip.txt
}

function main(){
    get_iftop_network_ip
    get_network_ip
    report
    rm -rf ./iftop_network_ip.txt
    rm -rf ./iftop_network_content.txt
}
main