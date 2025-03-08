#!/bin/bash
# 故障自愈脚本: 磁盘占用超过90%就清除之前的日志文件
hostname=$(hostname)
LOG_DIR=/data/log/tps
if [ $hostname == 'tfs03' ];then 
    DISK_USAGE=$(df -h | awk '/\/data$/ {print $(NF-1)}' | tr -s '%' ' ')
else
    DISK_USAGE=$(df -h | awk '/\/$/ {print $(NF-1)}' | tr -s '%' ' ')
fi

function report(){ 
    # Feishu webhook URL  
    WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/98c518dd-8faa-4ce0-b188-c939e1c34f4a"  
    # Local hostname and time  
    HOSTNAME=$(hostname)  
    LOCAL_TIME=$(date +"%Y-%m-%d %H:%M:%S")  
    
    # Alert headers and content  
    ALERT_HEADERS="夜莺监控故障自愈"  
    ALERT_CONTENT="🌟故障自愈成功!👏请查阅💪\n主机名称:  $hostname\n当前磁盘使用占比:  $DISK_USAGE%\n故障恢复时间:  $LOCAL_TIME"
    
    # Message body in JSON format  
    MESSAGE_BODY='{  
        "msg_type": "interactive",  
        "card": {  
            "config": {  
                "wide_screen_mode": true  
            },  
            "elements": [  
                {  
                    "tag": "div",  
                    "text": {  
                        "content": "'$ALERT_CONTENT'",  
                        "tag": "lark_md"  
                    }  
                }  
            ],  
            "header": {  
                "template": "green",  
                "title": {  
                    "content": "'$ALERT_HEADERS'",  
                    "tag": "plain_text"  
                }  
            }  
        }  
    }'  
    
    # Send the POST request to Feishu  
    curl -X POST -H "Content-Type: application/json" -d "$MESSAGE_BODY" "$WEBHOOK_URL"  
}

declare -a thresholds=(30 10 7 3 0)
for days in "${thresholds[@]}";do
    echo $days
    if [ $DISK_USAGE -ge 90 ];then
        find "$LOG_DIR" \( -name "*.log.*" -o -name "*.statlog.*" \) -mtime +"$days" -type f -exec rm -f {} \;
        echo "$days天前的日志文件已清除"
        if [ $DISK_USAGE -le 90 ];then
            report
        fi
    fi
done