#!/bin/bash

declare -A Other=(
    [feishuUrl]=''
)
logDir="/data/log/nginx"
logFile=(fetch_access.log svip_access.log ent_access.log)
forbidConf="/etc/nginx/forbid_order.conf"
dateTime=$(date -d "1 day ago" +"%F")


send_msg() {
    message="$1"
    url_token=${Other[feishuUrl]}
    data=$(cat <<EOF
{
    "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "[Nginx无效订单封禁通知]",
                "content": [
                    [{
                        "tag": "text",
                        "text": "$message"
                    }]
                ]
            }
        }
    }
}
EOF
)

    headers="Content-Type: application/json"
    curl -X POST -H "$headers" -d "$data" "$url_token"
}

logPrint(){
    time=$(date "+%Y-%m-%d %H:%M:%S")
    echo -ne "${time} - ${1}\n" >> $logDir/expire_order_forbid.log
}

writeConf(){
    grep -E "'$1'" $forbidConf >/dev/null
    if [[ $? -eq 1 ]];then
        echo "if (\$args  ~ '"$1"') {return 403;}" >> $forbidConf
        logPrint "write $1 to $forbidConf success!"
        return $?
    else
        return 1
    fi
}

checkOrder(){
    res=$(curl -s --request GET --url "API接口地址" --header 'API-AUTH: 鉴权信息' --header "orderid: $1")
    if [[ -n $(echo $res|grep 'success') ]];then
        echo $res|awk '{print $NF}'|awk -F '}}' '{print $1}'
    fi
}

checkSecret(){
    res=$(curl -s --request GET --url "API接口地址" --header 'API-AUTH: 鉴权信息' --header "secret_id: $1")
    if [[ -n $(echo $res|grep 'success') ]];then
        echo $res|awk '{print $NF}'|awk -F '}}' '{print $1}'|sed 's/\"//g'
    fi
}


cleanConf(){
    isClen=$(cat $forbidConf|wc -l)
    if [[ $isClen -ge 100 ]];then
        > $forbidConf
        reloadNginx
        if [[ $? -eq 0 ]];then
            logPrint "Nginx reload success!"
        else
            send_msg "$(hostname)配置文件有错误,请检查!"
        fi
    fi
}



reloadNginx(){
    /usr/sbin/nginx -t
    if [ $? -eq 0 ] ; then
        /usr/sbin/nginx -s reload
        return 0
    else
        return 1
    fi
}

scpFile(){
    name=$(hostname)
    if [[ $name == 'nginx' ]];then
        scp -P8301 $1 远程服务器地址:/tmp/dps_${1#*nginx/}
    elif [[ $name == 'nginx2' ]];then
        scp -P8301 $1 远程服务器地址:/tmp/dps2_${1#*nginx/}
    fi
}

cleanConf

for i in ${logFile[@]};do
    checkLog=$logDir/${i}.${dateTime}.gz
    if [[ ! -e $checkLog ]];then
        logPrint "$checkLog not found, skip!"
        continue
    fi
    scpFile $checkLog
    orderOrsecret=$(zcat $checkLog |awk '{if ($9 == 400){print $7}}'|sort |uniq|awk -F '[ |&|?]' '{print $2}'|awk -F 'orderid=|secret_id=' '{print $2}'|grep -v '^$')
    if [[ -n $orderOrsecret ]];then
        for j in $orderOrsecret;do
            if [[ $j =~ ^[[:digit:]]+$ ]];then
                isok=$(checkOrder $j)
                if [[ $isok == '"ok"' ]];then
                    writeConf $j
                    if [[ $? -eq 0 ]];then
                        ((n++))
                    fi
                fi
            elif [[ $j =~ ^[uo][a-z0-9]{19}$ ]];then
                orderid=$(checkSecret $j)
                if [[ -n $orderid ]];then
                    isok=$(checkOrder $orderid)
                    if [[ $isok == '"ok"' ]];then
                        writeConf "$orderid|$j"
                        if [[ $? -eq 0 ]];then
                            ((n++))
                        fi

                    fi
                else
                    logPrint "unknown order or secret: $j"
                fi
            fi
        done
    else
        logPrint "Not Orderid Or Secretid!"
    fi
done

if [[ $n -gt 0 ]];then
    reloadNginx
    if [[ $? -eq 0 ]];then
        logPrint "Nginx reload success!"
    else
        send_msg "$(hostname)配置文件有错误,请检查!"
    fi
else
    logPrint "Nothing to do!"
fi
