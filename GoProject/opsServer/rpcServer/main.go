package main

import (
	"opsServer/service"
	"log"
	"net"
	"net/rpc"
)

var _ service.KDLService = (*service.KDL)(nil)

func main() {
	rpc.Register(new(service.KDL))
	listen, err := net.Listen(service.RpcConfig.Protocol, service.RpcConfig.Address)
	if err != nil {
		panic(err.Error())
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}
		go rpc.ServeConn(conn)
	}
}
