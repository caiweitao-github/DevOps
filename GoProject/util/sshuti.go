package util

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SshInit interface {
	InitSsh() (*ssh.ClientConfig, error)
}

type SshConn interface {
	Conn(cfg *ssh.ClientConfig, cmd []string) (string, error)
}

type SshUtil interface {
	SshInit
	SshConn
}

type SshInfo struct {
	Host, Port string
}

func (S SshInfo) InitSsh() (*ssh.ClientConfig, error) {
	privateKeyPath := "/root/.ssh/id_rsa_kps"
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         8 * time.Second,
	}

	return sshConfig, nil
}

func (S SshInfo) Conn(cfg *ssh.ClientConfig, cmd []string) (string, error) {
	addr := fmt.Sprintf("%s:%s", S.Host, S.Port)
	var client *ssh.Client
	var err error
	client, err = ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return "", err
	}
	defer client.Close()
	for _, v := range cmd {
		session, err := client.NewSession()
		if err != nil {
			return "", err
		}
		defer session.Close()
		res, _err := session.CombinedOutput(v)
		if _err != nil {
			return "", _err
		} else {
			return string(res), nil
		}
	}
	return "", nil
}

// 下面这段逻辑包含重试机制

// func (S SshInfo) Conn(cfg *ssh.ClientConfig, cmd []string) (string, error) {
// 	addr := fmt.Sprintf("%s:%s", S.Host, S.Port)
// 	var client *ssh.Client
// 	var err error

// 	for retry := 0; retry < 5; retry++ {
// 		client, err = ssh.Dial("tcp", addr, cfg)
// 		if err == nil {
// 			break
// 		}
// 		time.Sleep(2 * time.Second)
// 	}

// 	if err != nil {
// 		return "", err
// 	}
// 	defer client.Close()
// 	for _, v := range cmd {
// 		session, err := client.NewSession()
// 		if err != nil {
// 			return "", err
// 		}
// 		defer session.Close()
// 		res, _err := session.CombinedOutput(v)
// 		if _err != nil {
// 			return "", _err
// 		} else {
// 			return string(res), nil
// 		}
// 	}
// 	return "", nil
// }
