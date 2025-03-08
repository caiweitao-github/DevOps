#!/bin/bash
file_name="/home/weitaocai/ansible/dps_inventory"
> $file_name
st="dps"
echo "dps:" > $file_name
echo "  hosts:" >> $file_name
mysql=''

get_code_info(){
    ip=$($mysql -N -e 'select login_ip from dps where code = "'$i'";')
    port=$($mysql -N -e 'select login_port from dps where code = "'$i'";')
    echo -e "    $i:\n      ansible_ssh_host: $ip\n      ansible_ssh_port: $port" >> $file_name
}

get_dps_group_info(){
    code=$($mysql -N -e 'select dps.code from dps,dps_group_relation,dpsgroup where dps_group_relation.group_id = dpsgroup.id and dps_group_relation.dps_id = dps.id and dpsgroup.code = "'$i'";')
    for dps in $code;do
        ip=$($mysql -N -e 'select login_ip from dps where code = "'$dps'";')
        port=$($mysql -N -e 'select login_port from dps where code = "'$dps'";')
        echo -e "    $dps:\n      ansible_ssh_host: $ip\n      ansible_ssh_port: $port" >> $file_name
    done
}

mapfile -t dps_code < code

for i in "${dps_code[@]}"; do
    if [[ "${i#"$st"}" != "$i" ]]; then
        get_code_info "$i"
    else
        get_dps_group_info "$i"
    fi
done
