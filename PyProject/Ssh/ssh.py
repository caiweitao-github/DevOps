import paramiko

# logger = get_logger("TpsSSH", "log/tps_svn_version.log")

class TpsSSH():
    def __init__(self):
        self.ssh = None
        self.username = ''
        self.password = ''
        self.ip = None
        self.port = 22
        
    def login_tps(self, ip):
        self.ssh = paramiko.SSHClient()
        self.ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())  # 跳过连接确认
        try:
            self.ssh.connect(ip, self.port, self.username, self.password)
            return 1
        except Exception as e:
            return 0
        
    def exec(self, svn, cmd, dir_path):
        path = '/tmp/%s' %(dir_path)
        stdin, stdout, stderr = self.ssh.exec_command("cd %s && svn up -r %s && %s" %(path, svn, cmd))
        stdout_info = stdout.read().decode('utf8')
        



    def close(self):
        if self.ssh:
            try:
                self.ssh.close()
            except BaseException as e:
                print(e)
                # logger.error("ssh close ->%s" % e)
                
params = {'tps_node_list': ['114.118.1.40'], 'tps_type': 'tps', 'svn_version': '33117', 'svn_parameter': 'echo "test" >> /tmp/svn.log'}
     
tps_ssh = TpsSSH()
ip_list = params.get('tps_node_list')
for ip in ip_list:
    login_result = tps_ssh.login_tps(ip)
    if login_result:
        tps_ssh.exec(svn=params.get('svn_version'), cmd=params.get('svn_parameter'), dir_path=params.get('tps_type'))
