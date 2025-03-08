#!/bin/bash

date=$(date +"%Y-%m-%d %H:%M:%S")
name=$(hostname)
IP=$(hostname -I | awk '{print $1}')
proxy_tdps=$(ps aux | grep "proxy_tdps" | grep -v "grep" | awk '{print $2}')
categraf_status=$(systemctl is-active categraf.service)
aps_status=$(systemctl is-active proxy_aps.service)

min=40001
max=42000
random_number=$((RANDOM % (max - min + 1) + min))

function report() {
    curl --request POST \
        -o /dev/null -s -w "%{http_code}\n" \
        --url https://test.com/modifytdpsstatus \
        --header 'API-AUTH: cd0b979ecdd34d2fd58d9a5900bd5a58' \
        --header 'Content-Type: application/json' \
        --data '{
            "ip": "'"${IP}"'",
            "code": "'"${name}"'"
        }'
}

if [[ -z "$proxy_tdps" ]]; then
    systemctl stop categraf.service
    systemctl stop proxy_aps.service
    proxy_tdps=$(ps aux | grep "proxy_tdps" | grep -v "grep" | awk '{print $2}')
    if [[ -n "$proxy_tdps" ]]; then
        kill -9 $proxy_tdps
    fi
    sleep 5

    # 找到占用端口的进程 ID
    port_Occupied=$(netstat -tulnp | grep -E ':4[0-9]{3}|:41[0-9]{2}|:420[0-9]|:4200' | awk '{print $7}' | cut -d'/' -f1 | sort -u)
    # 如果找到进程，则杀死它
    if [[ -z "$port_Occupied" ]]; then
        kill -9 $port_Occupied
        echo "$date $port_Occupied进程占用" >> /root/log/process_tdps.log
    fi

    sleep 2
    /root/tdps_linux_amd64/proxy_tdps >> /root/log/tdps_err.log 2>&1 &
    sleep 2

    proxy_tdps=$(ps aux | grep "proxy_tdps" | grep -v "grep" | awk '{print $2}')  # 更新状态
    if [[ -n "$proxy_tdps" ]]; then
        kill -9 $proxy_tdps
        sleep 3

        # 找到占用端口的进程 ID
        port_Occupied=$(netstat -tulnp | grep -E ':4[0-9]{3}|:41[0-9]{2}|:420[0-9]|:4200' | awk '{print $7}' | cut -d'/' -f1 | sort -u)
        # 如果找到进程，则杀死它
        if [[ -n "$port_Occupied" ]]; then
            kill -9 $port_Occupied
            echo "$date $port_Occupied进程占用" >> /root/log/process_tdps.log
        fi
        sleep 3
        /root/tdps_linux_amd64/proxy_tdps >> /root/log/tdps_err.log 2>&1 &

        # 再次检查进程
        sleep 2
        proxy_tdps=$(ps aux | grep "proxy_tdps" | grep -v "grep" | awk '{print $2}')  # 更新状态
        if [[ -z "$proxy_tdps" ]]; then
            for i in {1..3}; do
                sleep 1
                proxy_tdps=$(ps aux | grep "proxy_tdps" | grep -v "grep" | awk '{print $2}')  # 更新状态
                if [[ -n "$proxy_tdps" ]]; then
                    echo "$date proxy_tdps重启成功" >> /root/log/process_tdps.log
                    systemctl restart categraf.service
                    systemctl restart proxy_aps.service
                    break
                fi
                if [ "$i" -eq 3 ]; then
                    echo "$date proxy_tdps重启失败" >> /root/log/process_tdps.log
                    break
                fi
            done
        else
            echo "$date proxy_tdps重启成功" >> /root/log/process_tdps.log
            systemctl restart categraf.service
            systemctl restart proxy_aps.service
        fi
    else
        echo "$date proxy_tdps重启成功" >> /root/log/process_tdps.log
        systemctl restart categraf.service
        systemctl restart proxy_aps.service
    fi
else
    if [ "$aps_status" != "active" ]; then
        systemctl restart proxy_aps.service
    fi

    if [ "$categraf_status" != "active" ]; then
        systemctl restart categraf.service
    fi
fi
