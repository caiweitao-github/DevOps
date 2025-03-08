#!/bin/bash
declare -A dpsDate

declare -A text

declare -A SqlStr=(
 [allGroup]='select code from dpsgroup where code  REGEXP "^a.*|^c.*|^g.*|^k.*" and code not in ("ab", "cb", "gb", "kb")'
 [gId]='select id from dpsgroup where code = '
 [groupNum]='select count(*) from dps_group_relation where group_id = '
 [dpsGroupRelation]='select changeip_period,count(*) from dps,dpsgroup,dps_group_relation where dps.id=dps_group_relation.dps_id and dpsgroup.id=dps_group_relation.group_id  and dpsgroup.code="%s" group by changeip_period'
 [dpsId]='select dps_id from dps_group_relation where group_id = '
 )

declare -A Other=(
    [feishuUrl]=''
)

Db='mysql -udb -pdb -h10.0.5.41 -Ddb -N -e'


send_msg() {
    message="$1"
    url_token=${Other[feishuUrl]}
    data=$(cat <<EOF
{
    "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "[dps分组异常通知]",
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

aGroup(){
    sql=$(printf "${SqlStr[dpsGroupRelation]}" "$1")
    data=($($Db "$sql"))
    for i in ${!data[@]};do
        if [[ ${data[$i]} -eq 200 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        elif [[ ${data[$i]} -eq 320 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        fi
    done
    dpsId=$($Db "${SqlStr[dpsId]}'"$2"'")
    if [[ ${dpsDate[200]} -ne 5 ]] || [[ ${dpsDate[320]} -ne 2 ]];then
        for i in $dpsId;do
            ((count++))
            if [ $count -le 2 ];then
                $Db 'update dps set changeip_period = "320" where id = "'$i'";'
                echo "Dps ID: $i 设置为320."
            else
                $Db 'update dps set changeip_period = '200' where id = "'$i'";'
                echo "Dps ID: $i 设置为200."
            fi
        done
        count=0
        unset dpsDate
        return $?
    fi
    return 1
}

cGroup(){
    sql=$(printf "${SqlStr[dpsGroupRelation]}" "$1")
    data=($($Db "$sql"))
    for i in ${!data[@]};do
        if [[ ${data[$i]} -eq 630 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        elif [[ ${data[$i]} -eq 1860 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        fi
    done
    dpsId=$($Db "${SqlStr[dpsId]}'"$2"'")
    if [[ ${dpsDate[1860]} -ne 3 ]] || [[ ${dpsDate[630]} -ne 7 ]];then
        for i in $dpsId;do
            ((count++))
            if [ $count -le 3 ];then
                $Db 'update dps set changeip_period = "1860" where id = "'$i'";'
                echo "Dps ID: $i 1860."
            else
                $Db 'update dps set changeip_period = '630' where id = "'$i'";'
                echo "Dps ID: $i 设置为630."
            fi
        done
        count=0
        unset dpsDate
        return $?
    fi
    return 1
}


gGroup(){
    sql=$(printf "${SqlStr[dpsGroupRelation]}" "$1")
    data=($($Db "$sql"))
    for i in ${!data[@]};do
        if [[ ${data[$i]} -eq 1860 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        elif [[ ${data[$i]} -eq 2760 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        elif [[ ${data[$i]} -eq 3660 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        fi
    done
    dpsId=$($Db "${SqlStr[dpsId]}'"$2"'")
    if [[ ${dpsDate[1860]} -ne 17 ]] || [[ ${dpsDate[2760]} -ne 6 ]] || [[ ${dpsDate[3660]} -ne 5 ]];then
        for i in $dpsId;do
            ((count++))
            if [ $count -le 17 ];then
                $Db 'update dps set changeip_period = "1860" where id = "'$i'";'
                echo "Dps ID: $i 设置为1860."
            elif [[ $count -gt 17 ]] && [[ $count -le 23 ]];then
                $Db 'update dps set changeip_period = "2760" where id = "'$i'";'
                echo "Dps ID: $i 设置为2760."
            else
                $Db 'update dps set changeip_period = '3660' where id = "'$i'";'
                echo "Dps ID: $i 设置为3660."
            fi
        done
        count=0
        unset dpsDate
        return $?
    fi
    return 1
}

kGroup(){
    sql=$(printf "${SqlStr[dpsGroupRelation]}" "$1")
    data=($($Db "$sql"))
    for i in ${!data[@]};do
        if [[ ${data[$i]} -eq 10800 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        elif [[ ${data[$i]} -eq 14400 ]];then
            dpsDate[${data[$i]}]=${data[$i+1]}
        fi
    done
    dpsId=$($Db "${SqlStr[dpsId]}'"$2"'")
    if [[ ${dpsDate[10800]} -ne 12 ]] || [[ ${dpsDate[14400]} -ne 2 ]];then
        for i in $dpsId;do
            ((count++))
            if [ $count -le 12 ];then
                $Db 'update dps set changeip_period = "10800" where id = "'$i'";'
                echo "Dps ID: $i 设置为10800."
            else
                $Db 'update dps set changeip_period = '14400' where id = "'$i'";'
                echo "Dps ID: $i 设置为14400."
            fi
        done
        count=0
        unset dpsDate
        return $?
    fi
    return 1
}

Process(){
    case $1 in
    a[0-9]*)
        aGroup $code $gId
    ;;
    c[0-9]*)
        cGroup $code $gId
    ;;
    g[0-9]*)
        gGroup $code $gId
    ;;
    k[0-9]*)
        kGroup $code $gId
    ;;
        *)
        echo "Usage：bash $0 'group code'"
    ;;
esac
}

for code in $($Db "${SqlStr[allGroup]}");do
    #FIXME: 需要处理备用组
    gId=$($Db "${SqlStr[gId]}'"$code"'")
    groupNum=$($Db "${SqlStr[groupNum]}'"$gId"'")
    if ([[ $code =~ ^a[0-9]{1,3}$ ]] && [[ $groupNum -ne 7 ]]) || ([[ $code =~ ^c[0-9]{1,3}$ ]] && [[ $groupNum -ne 10 ]]) || ([[ $code =~ ^g[0-9]{1,3}$ ]] && [[ $groupNum -ne 28 ]]) || ([[ $code =~ ^k[0-9]{1,3}$ ]] && [[ $groupNum -ne 14 ]]);then
        bash /home/httpproxy/DevOps/Dps/function/Add_dps_to_group.sh $gId
    fi
    Process $code
    if [[ $? -eq 0 ]];then
        text[$code]=$code"组机器周期数异常,已自动修复。"
    fi
done

report_text="以下组分配已调整"
if [[ ${#text[@]} -ne 0 ]];then
    for txt in ${text[@]}
    do
        report_text="$report_text\n$txt"
    done
    send_msg $report_text
fi