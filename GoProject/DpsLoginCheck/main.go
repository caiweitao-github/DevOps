package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

type dpsInfo struct {
	code      string
	loginIP   string
	loginPort string
}

var db *sql.DB

func initDB() error {
	conn := "root:@tcp(127.0.0.1:3306)/db?charset=utf8"
	var err error
	db, err = sql.Open("mysql", conn)
	if err != nil {
		return err
	}
	return nil
}

func getInfo() ([]dpsInfo, error) {
	sqlStr := "SELECT code, login_ip, login_port FROM dps WHERE status = 1 AND dps_type != 2"
	rows, err := db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var infoList []dpsInfo
	for rows.Next() {
		var info dpsInfo
		err := rows.Scan(&info.code, &info.loginIP, &info.loginPort)
		if err != nil {
			log.Printf("scan failed, err:%v\n", err)
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		infoList = append(infoList, info)
	}
	return infoList, nil
}

func publicKeyAuthFunc(keyPath string) (ssh.AuthMethod, error) {
	keyPath, err := homedir.Expand(keyPath)
	if err != nil {
		return nil, fmt.Errorf("find key's home dir failed: %v", err)
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("ssh key file read failed: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("ssh key signer failed: %v", err)
	}
	return ssh.PublicKeys(signer), nil
}

func sshCon(ctx context.Context, infoCh chan dpsInfo, cfg *ssh.ClientConfig) {
	for {
		select {
		case <-ctx.Done():
			return
		case info, ok := <-infoCh:
			if !ok {
				return
			}
			addr := fmt.Sprintf("%s:%s", info.loginIP, info.loginPort)
			sshClient, err := ssh.Dial("tcp", addr, cfg)
			if err != nil {
				log.Printf("%s: %v", info.code, err)
				return
			} else {
				sshClient.Close()
				return
			}
		}
	}
}

func main() {
	runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪，block
	runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪，mutex

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	username := "root"
	ctx, cancel := context.WithCancel(context.Background())
	sshKeyPath := "/root/.ssh/id_rsa_kps"
	infoCh := make(chan dpsInfo)
	doneCh := make(chan bool)
	err := initDB()
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	infoList, err := getInfo()
	if err != nil {
		log.Fatalf("Failed to get info from DB: %v", err)
	}

	config := &ssh.ClientConfig{
		Timeout:         time.Second * 10,
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	authMethod, err := publicKeyAuthFunc(sshKeyPath)
	if err != nil {
		log.Fatalf("Failed to create auth method: %v", err)
	}
	config.Auth = []ssh.AuthMethod{authMethod}

	for _, info := range infoList {
		go sshCon(ctx, infoCh, config)
		infoCh <- info
	}
	defer func() {
		cancel()
		doneCh <- true
	}()
	defer close(infoCh)
	<-doneCh
}

