package main

import (
	"context"
	"fmt"
	"grpcTest/service"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Dial err: ", err)
	}
	defer conn.Close()

	client := service.NewKDLServiceClient(conn)

	req := &service.Parameter{ServerIP: "221.131.165.73", Num: 2}
	reply, err := client.GetData(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply.GetResults())
	for _, v := range reply.GetResults() {
		fmt.Println(v)
	}
}
