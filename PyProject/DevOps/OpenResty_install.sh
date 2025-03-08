#/bin/bash
# 基于Debain操作系统安装OpenResty
MAIN_CODE=$1
if [ -z $MAIN_CODE ] ; then
    echo "Miss machine code!"
    exit 1
fi

function install_nginx_conf() {
    # 备份原有的nginx.conf
    if [[ -f /usr/local/openresty/nginx/conf/${nginx_conf_prefix}nginx_nginx.conf ]];then
        mv /usr/local/openresty/nginx/conf/nginx.conf /usr/local/openresty/nginx/conf/nginx.conf.bak
        mv /usr/local/openresty/nginx/conf/${nginx_conf_prefix}nginx_nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
    fi
    # 备份原有的http.conf
    if [[ -f  /usr/local/openresty/nginx/conf/${nginx_conf_prefix}nginx_http.conf ]];then
        mv /usr/local/openresty/nginx/conf/http.conf /usr/local/openresty/nginx/conf/http.conf.bak
        mv /usr/local/openresty/nginx/conf/${nginx_conf_prefix}nginx_http.conf /usr/local/openresty/nginx/conf/http.conf
    fi

    if [[ $nginx_conf_prefix == dps ]];then
        mkdir -p /usr/local/openresty/nginx/Lua/
        \cp -r /root/frontserver/setup/Lua/* /usr/local/openresty/nginx/Lua/
        \cp -r /root/frontserver/setup/nginx/503_api.html /data/www/error_page/
    elif  [[ $nginx_conf_prefix == www ]];then
        \cp -r /root/frontserver/setup/nginx/rebots.txt /data/www/
        echo "ERROR(-101): 请求频率超限" >> /data/www/error_page/503_api.html
        echo "-10" >> /data/www/error_page/503.html
    elif [[ $nginx_conf_prefix == kgtools ]];then
        mkdir -p /data/www/error_page
        echo "ERROR(-101): 请求频率超限" >> /data/www/error_page/503_api.html
        echo "-10" >> /data/www/error_page/503.html
    fi
}

# 确认系统版本
lsb_release -a

#修改主机名称
printf "$MAIN_CODE" > /etc/hostname
hostnamectl set-hostname "$MAIN_CODE"  

apt update
apt -y install wget
apt -y install subversion
apt -y install vim

#拉取代码
echo yes |svn co --username 'httpproxy' --password 'nTF*nmG@K!' https://svn.gizaworks.com/repos/trunk/db/frontserver frontserver

#开启计划任务日志
echo "cron.*     /var/log/cron.log"  >> /etc/rsyslog.d/50-default.conf && service rsyslog restart

#添加用户
useradd -m henry -s /bin/bash
useradd -m kdl -s /bin/bash
echo 'kdl:kdlvps2016' | chpasswd
echo "root:Gizavps2014" | chpasswd

#kdl添加附属组
usermod -G adm kdl

#设置8022端口
sed -i '/Port 22/i\Port 8022' /etc/ssh/sshd_config
#修改sshd配置文件
echo "ClientAliveInterval 60" >> /etc/ssh/sshd_config
printf "\nAllowUsers root henry\nAllowUsers kdl@118.89.225.185 kdl@10.*\n" >> /etc/ssh/sshd_config
service ssh restart

#创建日志目录
mkdir -p /root/log/

#配置vim
cp /root/frontserver/setup/vimrc ~/.vimrc

#ulimit设为65535,Ubuntu必须指定单个用户进行配置,*是不会生效的，会走系统默认
printf "root soft nofile 65535\nroot hard nofile 65535\n" >> /etc/security/limits.conf
printf "root hard nproc 65535\nroot soft nproc 65535\n" >> /etc/security/limits.conf
echo "ulimit -SHn 65535" >> /etc/profile
echo "ulimit -SHu 65535" >> /etc/profile
ulimit -SHn 65535
ulimit -SHu 65535

#修改内核参数
printf "vm.overcommit_memory = 1\nnet.ipv4.tcp_syncookies = 1\nnet.ipv4.tcp_tw_reuse = 1\nnet.ipv4.tcp_tw_recycle = 1\nnet.ipv4.tcp_fin_timeout = 5\n" >> /etc/sysctl.conf
/sbin/sysctl -p

#更换时区
export TZ=\'Asia/Shanghai\' >> ~/.bash_profile
export TZ='Asia/Shanghai'
    
#尽可能使用物理内存，不用swap
sysctl vm.swappiness=0
echo "vm.swappiness=0" |tee -a /etc/sysctl.conf
    
#安装redis
apt -y install redis-server
mkdir /root/redis/
\cp /root/frontserver/setup/redis.conf /etc/redis/
printf "redis-server /etc/redis/redis.conf\n" >> /etc/rc.local
redis-server /etc/redis/redis.conf

apt -y install bash-completion
apt -y install htop
apt -y install iftop
apt -y install python-pip
apt -y install python-redis
apt -y install git
pip install msgpack
pip install requests
pip install redis==2.10.5

#配置
apt-get -y install --no-install-recommends wget gnupg ca-certificates
wget -O - https://openresty.org/package/pubkey.gpg | apt-key add -
release_verison=`grep -Po 'VERSION="[0-9]+ \(\K[^)]+' /etc/os-release`
echo "deb http://openresty.org/package/debian $release_verison openresty" \
    | tee /etc/apt/sources.list.d/openresty.list
apt-get update || (echo "update fail!" && exit 255)
apt-get -y install openresty  #生成 DHE 参数
#为了避免使用 OpenSSL 默认的 1024bit DHE 参数，我们需要生成一份更强的参数文件
cd /etc/ssl/certs;openssl dhparam -out dhparam.pem 2048;

#创建日志文件
echo "cron.*     /var/log/cron.log"  >> /etc/rsyslog.d/50-default.conf && service rsyslog restart

cp /root/frontserver/setup/logrotate.d/nginx /etc/logrotate.d/
cp /root/frontserver/setup/nginx/503* /data/www/error_page
mkdir -p /data/{www/{proxy_temp_dir,proxy_cache_dir,error_page},log/nginx}
mkdir -p /root/{log,redis}     #负载监控
apt -y install dstat
cp /root/frontserver/setup/dstat.cron /etc/cron.d/dstat
chmod a+x /etc/cron.d/dstat

cp /root/frontserver/setup/version_control.cron /etc/cron.d/version_control
chmod a+x /etc/cron.d/version_control

\cp -r /root/frontserver/setup/nginx/* /usr/local/openresty/nginx/conf/

ln -s /usr/local/openresty/nginx/conf /etc/nginx
ln -s /usr/local/openresty/nginx/sbin/nginx  /usr/bin/nginx
systemctl stop openresty.service


#nginx日志监控
cat > /etc/logrotate.d/nginx << EOF
/data/log/nginx/*.log {
        daily
        dateext
        dateformat .%Y-%m-%d
        rotate 30
        compress
        delaycompress
        notifempty
        create 644 root root
        sharedscripts
        postrotate
                [ ! -f /var/run/nginx.pid ] || kill -USR1 \`cat /var/run/nginx.pid\`
        endscript
}
EOF
 
#复制配置文件到当前目录
cp -r /root/frontserver/bin/ /root/

python -u /root/bin/webapi.py >> /root/log/webapi.log 2>&1 &
python -u /root/bin/ip_access_monitor_www.py  >> /root/log/ip_access_monitor_www.log 2>&1 &
python -u /root/bin/ip_access_monitor_for_free.py  >> /root/log/ip_access_monitor_for_free.log 2>&1 &

#kdl用户免密码登录
mkdir -p /home/kdl/.ssh/
chmod 700 /home/kdl/.ssh/
cp /root/frontserver/setup/authorized_keys /home/kdl/.ssh/
chmod 600 /home/kdl/.ssh/authorized_keys
chown -R kdl:kdl /home/kdl/.ssh/

echo  "source ~/frontserver/setup/alias.sh" >> ~/.bash_profile
    
nginx_conf_prefix=$(echo $MAIN_CODE | awk -F 'nginx' '{print $1}')
if [[ $nginx_conf_prefix == dps || $nginx_conf_prefix == master || $nginx_conf_prefix == www || $nginx_conf_prefix == svip || $nginx_conf_prefix == kgtools ]];then
    install_nginx_conf
fi

cpu_cores=`cat /proc/cpuinfo| grep "cpu cores"| uniq | awk -F":" '{print $2}'`
sed -i "s/^worker_processes.*/worker_processes $cpu_cores;/g" /usr/local/openresty/nginx/conf/nginx.conf

if [[ -f /var/run/nginx.pid ]];then
    nginx -t
    nginx -s reload
    echo "安装完成"
else
    nginx
    exit 1
fi