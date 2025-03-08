# DevOps

```
此代码仓库为日常运维中使用的脚本文件, 包括异常节点重启, 异常信息告警, 相关节点信息入库以及信息查询等快捷运维脚本
已废弃的脚本暂未统计代码说明及使用方式
```

# 代码说明

```
1. tp_ipnumber.py 
    主要作用: 检测第三方节点实时可用ip, 小于10%则说明有问题飞书通知到运维告警群, 一小时内只告警一次
    运行方式: 计划任务 ---> */2 * * * * python -u /home/httpproxy/bin/tp_ipnumber.py >/dev/null 2>&1
    运行机器: main10

2.tps_host_check.py
    主要作用: 查询最近1天切换的隧道域名对应的tid是否还请求之前分配的tps_code, 不一致则飞书告警
    运行方式: 计划任务 ---> 20 9 * * 1-6 python -u /home/httpproxy/bin/tps_host_check.py 2>&1
    运行机器: main13

3.webstat_check.sh
    主要作用: 根据查询Nginx相关日志, 检测登录webstat是否是白名单IP/武汉地区IP, 如果不是则触发告警
    运行方式: 常驻进程
    运行机器: opschecker1

4.xx_check.sh
    主要作用: 每小时检测一次小熊ip的余额，小于100000飞书告警
    运行方式: 计划任务 ---> 0 * * * * bash /root/bin/xx_check.sh
    运行机器: n9e_server

5.network.sh
    主要作用: 监控网卡实时出入带宽
    运行方式: 手动执行
    使用方式: bash network.sh 网卡名称

6.getmyip_check.sh 
    主要作用: 检测最近一分钟内是否有刷/api/getmyip接口的ip，并封禁
    运行方式: 计划任务 ---> */1 * * * * bash /root/bin/getmyip_check.sh >> /root/log/getmyip_check.log 2>&1
    运行机器: svipnginx

7.django_check.sh
    主要作用: 检测最近一分钟的webhp的django.log日志，对ERROR的日志告警到飞书Django日志告警
    运行方式: 计划任务 ---> */1 * * * * /bin/bash /home/httpproxy/bin/django_check.sh
    运行机器: main4

8.depth_check.sh
    主要作用: 检测每个requests的depth报错数量是否大于1w，大于则告警到飞书
    运行方式: 计划任务 ---> 0 * * * * bash /home/httpproxy/bin/depth_check.sh
    运行机器: k8s1, k8s2

9.code502_click.sh
    主要作用: 检测最近一分钟内是否有status code 502的报错日志信息, 有则推送到飞书
    运行方式: 计划任务 ---> */1 * * * * bash /root/bin/code502_click.sh
    运行机器: tfs中转服务器所有机器

10.ck_daily_del_old_data.sh
    主要作用: 定期删除ck里的老数据，释放磁盘空间
    运行方式: 计划任务 ---> 0 1 * * * sh /home/httpproxy/devops/ck_daily_del_old_data.sh >> $LOG_DIR/devops/ck_daily_del_old_data.log 2>&1
    运行机器: main13

11.check_sproxy.sh
    主要作用: 检查redis哈希表中status占比低于50%，就kill掉柚雷的sproxyserver进程
    运行方式: 计划任务 ---> crontab */10 * * * * bash /root/bin/check_sproxy.sh
    运行机器: 所有柚雷的服务器

12.transfer/transfer_check.py
    主要作用: 中转服务带宽检测脚本，持续超带宽会电话通知运维人员
    运行方式: 计划任务 ---> */5 * * * * python -u /home/httpproxy/DevOps/transfer/transfer_check.py
    运行机器: main13

13.mysql/mysql_ha.py
    主要作用: MySQL数据库主从复制检测, 检测到主从复制失败则会飞书预警
    运行方式: 计划任务 ---> */10 * * * * python /home/httpproxy/DevOps/mysql/mysql_ha.py
    运行机器: main13

14.Other/cos_check.py
    主要作用: 检测对象存储使用容量
    运行方式: 计划任务 --> 30 9 * * * python -u /home/httpproxy/DevOps/Other/cos_check.py
    运行机器: main13

15.Kps/kps_changeip_email.py
    主要作用: 手动更换KPS IP后邮件通知用户
    运行方式: 在main13上手动执行
    使用方式: python -u kps_changeip_email_new.py old_ip:new_ip 

16.Kps/kps_renew_notify.py
    主要作用: KPS独享型续费通知
    运行方式: 计划任务 ---> 1 9-23 * * * python -u /home/httpproxy/DevOps/Kps/kps_renew_notify.py > /dev/null 2>&1
    运行机器: main13

17.Dps/check_dps_city_code.py
    主要作用: 定期检测dps表中province_code,city_code为空的正常dps,并将它们按照location进行匹配后填写
    运行方式: 计划任务 ---> 1 8-21 * * * python -u /home/httpproxy/DevOps/Dps/check_dps_city_code.py >> /dev/null 2>&1
    运行机器: main13

18.Dps/Dps_check_info.py
    主要作用: 修复数据库中位置,运营商,登录IP和端口为空的异常数据,如若修复失败飞书推送至运维群内
    运行方式: 计划任务 ---> 1 8-21 * * * python -u /home/httpproxy/DevOps/Dps/Dps_check_info.py >> /dev/null 2>&1
    运行机器: main13

19.Dps/get_abnormal_machine.sh
    主要作用: 通过查询数据库获取前一天dps异常数据个数
    运行方式: 计划任务 ---> 1 9 * * * bash /home/httpproxy/DevOps/Dps/get_abnormal_machine.sh >> /dev/null 2>&1
    运行机器: main13

20.transfer_server_bindwidth_check.sh 
    主要作用: 通过监听dstat.log 文件获取实时带宽是否超过规定阈值, 若超过规定阈值则查询iftop实时占用带宽的客户端IP(过滤中转IP及服务器IP)的订单信息进行飞书预警通知
    运行方式: 常驻进程 ---> nohup /bin/bash 文件名 >> /dev/null 2>&1 &
    运行机器: 所有中转服务器,包括JDE/TFS/YLES等
    特别注意: 脚本需要根据实际的日志记录字段或带宽最大值及阈值比例进行动态更新

21.OpenResty_install.sh
    主要功能: 主要是完成OpenResty的安装, 次要的功能就是根据输入的code匹配安装对应的Nginx服务器
    运行方式: 手动执行
    使用方式: bash OpenResty_install.sh www1  // 就会自动安装wwwnginx服务器,从而进行Nginx配置文件的修改

22.check_tps_version.sh
    主要功能: 在使用已经创建好TPS的镜像后开机自动升级tps和tps_nqs版本并重启相关服务
    运行方式: 开机自启
    使用方式: 在创建tps镜像时, 将此脚本放入/etc/rc.local中后保存退出

23.speedtest.sh
    主要功能: 测试vps或其他服务器的上下行带宽, 采用多节点进行测试
    运行方式: 手动执行
    使用方式: bash speedtest.sh
    特别注意: 此脚本测试可能存在失败的情况, 这个时候就需要使用其他测试工具了. 如: curl 命令

24.n9e_tps_clear_log.sh
    主要功能: 使用夜莺监控的故障自愈功能对tps或其他服务器进行磁盘清理
    运行方式: 由夜莺监控触发, 有告警自动触发
    特别注意: 由于是夜莺触发, 缺少人工干预, 所以在编写告警规则及编写脚本时需要多次debug, 避免删除掉重要文件

25.get_ime_ec_request_domain.py
    主要功能: 统计IME/EC设备请求的域名,jip_node_index以及出口IP统计
    运行方式: 计划任务 ---> 10 6 * * * python -u /tmp/haohe/get_ime_ec_request_domain.py >> /dev/null 2>&1
    运行机器: main13

26.delete_oss_file.py
    主要功能: 删除30天前的oss文件, 避免资源消耗对象存储桶的容量
    运行方式: 计划任务 --->  30 6 * * * python -u /tmp/haohe/delete_oss_file.py >> /dev/null 2>&1 
    运行机器: main13

27.get_jip_bandwidth.py
    主要功能: 使用pycurl 模块对代理IP进行带宽测试入库, 方便数据统计
    运行方式: 手动运行
    特别注意: 可以根据此代码对其他产品的上下行进行测试入库

28.Kps/error_kpsorder_check.sh
    主要功能: 检测KPS分配异常的情况, 如 同一用户多个订单分配到相同的KPS的情况等. 如出现这种情况飞书预警
    运行方式: 计划任务 ---> 1 9 * * * bash /home/httpproxy/DevOps/Kps/error_kpsorder_check.sh >> /dev/null 2>&1
    运行机器: main13
29.jipscheduler_check.sh
    主要功能: 检测jip_server如果在数据库中已经启用，但是5分钟内还没在scheduler注册，则飞书告警
    运行方式: 计划任务 ---> */5 * * * * bash /home/httpproxy/bin/jipscheduler_check.sh >> /dev/null 2>&1
    运行机器: main13
```

# 其他代码说明:

```
运行机器: main13 
    1.技术支持数据日报
        30 9 * * 1 python -u /home/httpproxy/bin/push_custom_support_data.py
    2.Nginx过期订单检测
        30 10 * * * bash /home/httpproxy/bin/expire_order_check.sh
    3.最近一小时请求检查
        5 */1 * * * /home/httpproxy/bin/requests_check
    4.sfps续费通知
        1 9-23 * * * python -u /home/httpproxy/bin/sfps_renew_notify.py
    5.IP池统计
        */1 * * * *  /home/httpproxy/bin/IpPooL
    6.IP池数据日报
        10 9 * * * python -u /home/httpproxy/bin/ip_pool_push_data.py
    7.第三方节点请求域名统计
        10 6 * * * python -u /tmp/haohe/get_ime_ec_request_domain.py >> /dev/null 2>&1
        30 6 * * * python -u /tmp/haohe/delete_oss_file.py >> /dev/null 2>&1 
    8.kps异常分配通知
        * */1 * * * python -u /home/httpproxy/bin/kps_check.py
    9.定时将不存在于IP池的node加入默认池
        * */1 * * * bash /home/httpproxy/DevOps/mysql/kdlnode_insert.sh >> /dev/null 2>&1
    10.删除第三方节点换IP历史数据
        0 4 * * *   python -u /home/httpproxy/DevOps/mysql/delete_3RD_changeip_history.py >> /data/kdl/log/devops/delete_changeip_history.log 2>&1
    11.检查dps异常换IP
        3 9 * * * bash /home/httpproxy/devops/dps_cheackip_err.sh >> /dev/null 2&>1
    12.检查Tps getnode提取为空
        */1 * * * * python -u /home/httpproxy/devops/monitoring_tps_getnode.py >> /dev/null 2&>1
    13.检查jip_server是否注册
        */5 * * * * bash /home/httpproxy/bin/jipscheduler_check.sh >> /dev/null 2>&1
```

# 2024-09-27 新增项目 ../API/
以下内容为各大供应商可用API进行操作的项目, 介绍如下:
../API/volcano_project: 火山引擎接口, 功能有: 续费, 开通, 删除 等功能
../API/ucloud_project: ucloud接口, 主要功能有: 取消/开通自动续费, 查询余额, 续费实例IP等功能
../API/vps_project: 263vps/乐讯达网络接口, 主要功能: 查询官网后台与供应商后台进行对比, 避免误下架等操作
../API/tencent_project/Tencent_CDN_AnalysisData_Project: 腾讯接口, 主要功能: 获取三个网站的访问流量, 超过1G进行封禁处理, 每天9天进行飞书通知前一天的流量详情
../API/tencent_project/Tencent_Fps_Network_Project: 腾讯接口, 主要功能: 查询fps服务器的流量使用情况, 飞书通知
