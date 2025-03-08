package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"util"
)

var (
	logFile  = "/data/kdl/log/devops/code_repeat_check.log"
	logInfo  = util.LogConf(logFile, "[INFO] ")
	logError = util.LogConf(logFile, "[ERROR] ")
)

var db *sql.DB
var nodeops *sql.DB

var port = []string{""}

func main() {
	logInfo.Println("run...")
	start := time.Now()
	getRepeatDps()
	end := time.Now()
	logInfo.Printf("Time Use: %v", end.Sub(start))
}

func init() {
	var err error
	db, err = util.ConnDb("", "", "")
	if err != nil {
		logError.Fatalf("connect to db db failed: %v", err)
	}

	nodeops, err = util.ConnDb("", "", "")
	if err != nil {
		logError.Fatalf("connect to nodeops db failed: %v", err)
	}
}

func getExceptCode() []string {
	exceptCode := make([]string, 0)
	sql := "select code from dps where dps_type = 2 and status in (1,3)"
	rows, err := db.Query(sql)
	if err != nil {
		logError.Fatalf("query except code failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		err := rows.Scan(&code)
		if err != nil {
			logError.Fatalf("scan except code failed: %v", err)
		}
		exceptCode = append(exceptCode, code)
	}
	return exceptCode
}

func getDpsLoging(code string) (ip, port string) {
	sql := "select login_ip,login_port from dps where code = ?"
	err := db.QueryRow(sql, code).Scan(&ip, &port)
	if err != nil {
		logError.Fatalf("query dps login info failed: %v", err)
	}
	return
}

func ssh(ip, port string, cmd []string) (ppp string, err error) {
	var Cmd util.SshUtil = util.SshInfo{Host: ip, Port: port}
	SshCfg, err := Cmd.InitSsh()
	if err != nil {
		logError.Fatalf("init ssh failed: %v", err)
	}
	ppp, err = Cmd.Conn(SshCfg, cmd)
	return
}

func getRepeatDps() {
	var errip string
	exCode := getExceptCode()
	dpsinfo := make([]string, 0)
	sql := "select dps_code,ip,count(ip) from dps_changeip_history where change_time > DATE_SUB(NOW(), INTERVAL 200 second) group by dps_code,ip having count(ip)>2 limit 2"
	rows, err := nodeops.Query(sql)
	if err != nil {
		logError.Fatalf("query repeat dps failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var code, ip, num string
		err := rows.Scan(&code, &ip, &num)
		if err != nil {
			logError.Fatalf("scan repeat dps failed: %v", err)
		}
		for _, v := range exCode {
			if v == code {
				logInfo.Printf("%s 为海外机器.", code)
				return
			}
		}
		dpsinfo = append(dpsinfo, code, ip, num)
	}
	if len(dpsinfo) == 6 {
		cmd := []string{"source /root/.bash_profile;shutdown -h -t 5"}
		if dpsinfo[0] == dpsinfo[3] {
			loginip, loginport := getDpsLoging(dpsinfo[0])
			ppp, err := ssh(loginip, loginport, []string{"ip a|grep 'ppp0'|tail -n 1|awk '{print $2}'"})
			if err != nil {
				util.SendMess([]string{fmt.Sprintf("%s (ssh root@%s -p%s)存在IP重复上报情况, 但机器无法登录, 请手动检查.", dpsinfo[0], loginip, loginport)})
				return
			}
			if strings.TrimSpace(ppp) == strings.TrimSpace(dpsinfo[1]) {
				errip = dpsinfo[4]
			} else if strings.TrimSpace(ppp) == strings.TrimSpace(dpsinfo[4]) {
				errip = dpsinfo[1]
			}
		}
		if errip != "" {
			for _, p := range port {
				_, err := ssh(errip, p, cmd)
				if err != nil {
					logError.Printf("shutdown %s failed: %v", errip, err)
				} else {
					util.SendMess([]string{fmt.Sprintf("%s 存在code重复上报情况重复IP为%s, 已自动处理.", dpsinfo[0], errip)})
					return
				}
			}
			util.SendMess([]string{fmt.Sprintf("%s 存在code重复上报情况重复IP为%s, 关机失败, 请手动检查.", dpsinfo[0], errip)})
		}
	}
}
