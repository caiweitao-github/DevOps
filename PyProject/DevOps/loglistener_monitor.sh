#!/bin/bash

#服务器名称
hostname=$(cat /etc/hostname)
proxy_tdps=$(ps aux | grep "proxy_tdps" | grep -v "grep" | awk '{print $2}')

# 将20MB转换为字节
twenty_mb=$((20 * 1024 * 1024))

# 初始化上一次文件大小变量
last_file_size_save_tdps=$(stat -c %s /root/log/loglistener/save.log)
last_file_size_cls_tdps=$(stat -c %s /root/log/cls.log)

current_file_size_aps_offse=$(cat /root/loglistener/aps_offset)
last_file_size_cls_aps=$(stat -c %s /data/log/aps/cls.log)

function report(){
    # Feishu webhook URL  
    WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/7c6fe762-b427-463a-bc47-2858356b0991"
    # Local hostname and time  
    HOSTNAME=$(hostname)
    LOCAL_TIME=$(date +"%Y-%m-%d %H:%M:%S")

    # Alert headers and content  
    ALERT_HEADERS="进程监控告警"
    ALERT_CONTENT="故障时间:  $LOCAL_TIME\n$hostname $1"

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
                "template": "red",  
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

function check_logs {
    counter=0
    if [ -n "$proxy_tdps" ]; then
        while true; do
            ((counter++))
            if [ "$counter" -gt 3 ]; then
                break
            fi

            # 等待5秒后再次执行
            sleep 15

            # 获取当前文件大小并进行对比
            current_file_size_save=$(stat -c %s /root/log/loglistener/save.log)
            current_file_size_cls=$(stat -c %s /root/log/cls.log)

            if [ "$current_file_size_save" -eq "$last_file_size_save_tdps" ]; then
                if [ "$current_file_size_cls" -ne "$last_file_size_cls_tdps" ]; then
                    report "loglistener_tdps 日志僵死，请尽快检查"
                    break
                fi
            fi

            # aps日志判断
            current_file_size_aps_offse=$(cat /root/loglistener/aps_offset)
            current_file_size_cls_aps=$(stat -c %s /data/log/aps/cls.log)

            chazhi=$(($current_file_size_aps_offse - $current_file_size_cls_aps))

            if [ "$chazhi" -gt "$twenty_mb" ]; then
                report "loglistener_aps 日志僵死，请尽快检查"
                break
            fi

        done
    fi
}

# 调用函数
check_logs
