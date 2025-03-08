package main

import (
	"context"
	"database/sql"
	"fmt"
	"grpcTest/service"
	"net"
	"util"

	"google.golang.org/grpc"
)

var Db *sql.DB

type KDL struct {
	service.UnimplementedKDLServiceServer
}

var _ service.KDLServiceServer = (*KDL)(nil)

func main() {
	grpcServer := grpc.NewServer()
	service.RegisterKDLServiceServer(grpcServer, new(KDL))

	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
	}
	grpcServer.Serve(listen)
}

func init() {
	var err error
	Db, err = util.KdlnodeDB()
	if err != nil {
		panic(err)
	}
}

func (g *KDL) GetData(ctx context.Context, parameter *service.Parameter) (*service.ResultList, error) {
	sqlStr := `SELECT transfer_server.code,node.code,source,ip,port FROM node,transfer_server
    WHERE RAND() <= 1 and transfer_server.status in (1,3) and transfer_server.login_ip = node.ip
    and transfer_server.login_ip = ? and source in (select source from node where ip = ? group by source) and node.status = 1 ORDER BY RAND()
    LIMIT ?`

	rows, err := Db.Query(sqlStr, parameter.ServerIP, parameter.ServerIP, parameter.Num)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resultList service.ResultList
	resultList.Results = make([]*service.Result, 0) // Initialize the Results field

	for rows.Next() {
		var r service.Result
		err := rows.Scan(&r.ServerCode, &r.NodeCode, &r.Source, &r.NodeIP, &r.NodePort)
		if err != nil {
			return nil, err
		}
		resultList.Results = append(resultList.Results, &r)
	}

	return &resultList, nil
}
