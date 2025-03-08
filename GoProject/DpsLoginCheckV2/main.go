package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
)

type DpsInfo struct {
	code      string
	loginIP   string
	loginPort string
}

var Db *sql.DB

func InitDb() (err error) {
	dsn := "root:@tcp(127.0.0.1:3306)/db?charset=utf8"
	Db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	return nil
}

func GetExceptionDps() ([]DpsInfo, error) {
	sqlStr := "SELECT code,login_ip,login_port FROM dps WHERE status = 1 AND dps_type != 2 and provider not like '%自营%' and provider != '骐网科技'"
	DpsList := make([]DpsInfo, 0)
	rows, err := Db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var dps DpsInfo
		err := rows.Scan(&dps.code, &dps.loginIP, &dps.loginPort)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		DpsList = append(DpsList, dps)
	}
	return DpsList, nil
}

func InitSsh() (*ssh.ClientConfig, error) {
	privateKeyPath := "/root/.ssh/id_rsa_kps"
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Printf("Can't Read PrivateKeyFile: %s\n", err)
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		log.Printf("Can't Resolution PrivateKeyFile: %s\n", err)
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return sshConfig, nil
}

func SshUtil(host DpsInfo, wg *sync.WaitGroup, cfg *ssh.ClientConfig, ctx context.Context) {
	defer wg.Done()
	addr := fmt.Sprintf("%s:%s", host.loginIP, host.loginPort)

	select {
	case <-ctx.Done():
		log.Printf("%s time out!\n", host.code)
		return
	default:
		client, err := ssh.Dial("tcp", addr, cfg)
		if err != nil {
			log.Printf("failed connect %s: %s\n", host.code, err)
			return
		}
		defer client.Close()
	}
}

func main() {
	err := InitDb()
	if err != nil {
		fmt.Printf("init db failed, err: %v\n", err)
		return
	}
	wg := sync.WaitGroup{}
	infoList, err := GetExceptionDps()
	if err != nil {
		log.Fatalf("Failed to get info from DB: %v", err)
	}
	wg.Add(len(infoList))
	SshCfg, err := InitSsh()
	if err != nil {
		panic(err)
	}
	for _, i := range infoList {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		go SshUtil(i, &wg, SshCfg, ctx)
		defer cancel()
	}
	wg.Wait()
}
