#!/usr/bin/env python
# -*- encoding: utf-8 -*-
    
import paramiko
import log

class TpsSSH():
    def __init__(self):
        self.ssh = None
        self.key = None
        self.username = 'root'
        self.password = 'Gizatps@2020'
        self.ip = None
        self.port = 8022
        
    def login_tps(self, ip):
        self.ssh = paramiko.SSHClient()
        self.key = paramiko.RSAKey.from_private_key_file('/root/.ssh/id_rsa_kps')
        self.ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        try:
            self.ssh.connect(ip, self.port, self.username, pkey=self.key)
            log.info("connent  %s ..." % ip)
            return True
        except Exception as e:
            log.error("ssh close ->%s" % e)
        
    def exec(self, cmd):
        try:
            stdin, stdout, stderr = self.ssh.exec_command(cmd)
            stdout_info = stdout.read().decode('utf8')
            stdout_err = stderr.read().decode('utf8')
            if stdout_info:
                log.info("exec_command: %s" % stdout_info)
            if stdout_err:
                log.error("exec_command: %s" % stdout_err)
        except Exception as e:
            log.error(e)

    def close(self):
        if self.ssh:
            try:
                self.ssh.close()
            except BaseException as e:
                log.error("ssh close ->%s" % e)