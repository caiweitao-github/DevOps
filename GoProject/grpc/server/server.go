package main

import (
	"context"
	"grpc/service"
	"log"
	"net"

	"google.golang.org/grpc"
)

type HelloService struct {
	// UnimplementedHelloServiceServer这个结构体是必须要内嵌进来的
	// 也就是说我们定义的这个结构体对象必须继承UnimplementedHelloServiceServer。
	// 嵌入之后，我们就已经实现了GRPC这个服务的接口，但是实现之后我们什么都没做，没有写自己的业务逻辑，
	// 我们要重写实现的这个接口里的函数，这样才能提供一个真正的rpc的能力。
	service.UnimplementedHelloServiceServer
}

var _ service.HelloServiceServer = (*HelloService)(nil)

// Hello 重写实现的接口里的Hello函数
func (p *HelloService) Hello(ctx context.Context, req *service.Request) (*service.Response, error) {
	resp := &service.Response{}
	resp.Value = "hello:" + req.Value
	return resp, nil
}

func main() {
	// 首先通过grpc.NewServer() 构造一个grpc服务对象
	grpcServer := grpc.NewServer()
	// 然后通过grpc插件生成的RegisterHelloServiceServer函数注册我们实现的HelloService服务。
	service.RegisterHelloServiceServer(grpcServer, new(HelloService))

	listen, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen TCP err:", err)
	}
	//最后通过grpcServer.Serve(listen) 在一个监听端口上提供gRPC服务
	grpcServer.Serve(listen)
}
