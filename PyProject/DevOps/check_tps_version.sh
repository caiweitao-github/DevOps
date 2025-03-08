#!/bin/bash
# 检测tps项目svn版本并升级
# 检测tps版本
function check_tps_version() {
    # 获取最新tps版本
    current_tps_version=$(ssh root@106.12.253.107 -p8022 "/root/tps_Linux_amd64/tps --version | grep 'version'" | awk -F ': v' '{print $2}')
    # 获取镜像tps版本
    mirror_tps_version=$(/root/tps_Linux_amd64/tps --version | grep 'version' | awk -F ': v' '{print $2}')
    # 如果当前tps版本不是最新tps版本，则升级tps版本
    if [[ $current_tps_version != $latest_tps_version ]]; then
        wget -P /root/tps_Linux_amd64 https://kdl-tps.oss-cn-beijing.aliyuncs.com/tpsinstall/tps.tar.gz;tar -xf /root/tps_Linux_amd64/tps.tar.gz -C /root/tps_Linux_amd64/
        systemctl restart proxy_tps;systemctl restart proxy_tps_b
    fi
}

# 检测tps_nqs版本
function check_nqs_version() {
    # 获取当前tps_nqs版本
    current_nqs_version=$(ssh root@106.12.253.107 -p8022 "cd /root/tps_nqs;svn info | grep 'Revision'" | awk -F ': ' '{print $2}')
    # 升级镜像tps_nqs版本
    cd /root/tps_nqs
    echo yes | svn up --username 'tps' --password 'Gizatps.1908' $tps_nqs_version
    bash /root/tps_nqs/restart_auth.sh
    bash /root/tps_nqs/restart_webapi.sh
    bash /root/tps_nqs/restart_webapi_b.sh
}
# 免密步骤
apt-get -y install sshpass  >> /dev/null
ssh-keygen -t rsa -P "" -f ~/.ssh/id_rsa  >> /dev/null
sshpass -p'Gizatps@2020' ssh-copy-id -f -i ~/.ssh/id_rsa.pub "-o StrictHostKeyChecking=no" 106.12.253.107 -p8022 >> /dev/null
if [ $? -ne 0 ];then
    exit 1
fi
check_tps_version
check_nqs_version