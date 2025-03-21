// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: kdl.proto

package service

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	KDLService_GetData_FullMethodName = "/kdl.KDLService/GetData"
)

// KDLServiceClient is the client API for KDLService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KDLServiceClient interface {
	GetData(ctx context.Context, in *Parameter, opts ...grpc.CallOption) (*ResultList, error)
}

type kDLServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewKDLServiceClient(cc grpc.ClientConnInterface) KDLServiceClient {
	return &kDLServiceClient{cc}
}

func (c *kDLServiceClient) GetData(ctx context.Context, in *Parameter, opts ...grpc.CallOption) (*ResultList, error) {
	out := new(ResultList)
	err := c.cc.Invoke(ctx, KDLService_GetData_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KDLServiceServer is the server API for KDLService service.
// All implementations must embed UnimplementedKDLServiceServer
// for forward compatibility
type KDLServiceServer interface {
	GetData(context.Context, *Parameter) (*ResultList, error)
	mustEmbedUnimplementedKDLServiceServer()
}

// UnimplementedKDLServiceServer must be embedded to have forward compatible implementations.
type UnimplementedKDLServiceServer struct {
}

func (UnimplementedKDLServiceServer) GetData(context.Context, *Parameter) (*ResultList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetData not implemented")
}
func (UnimplementedKDLServiceServer) mustEmbedUnimplementedKDLServiceServer() {}

// UnsafeKDLServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KDLServiceServer will
// result in compilation errors.
type UnsafeKDLServiceServer interface {
	mustEmbedUnimplementedKDLServiceServer()
}

func RegisterKDLServiceServer(s grpc.ServiceRegistrar, srv KDLServiceServer) {
	s.RegisterService(&KDLService_ServiceDesc, srv)
}

func _KDLService_GetData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Parameter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KDLServiceServer).GetData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KDLService_GetData_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KDLServiceServer).GetData(ctx, req.(*Parameter))
	}
	return interceptor(ctx, in, info, handler)
}

// KDLService_ServiceDesc is the grpc.ServiceDesc for KDLService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KDLService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kdl.KDLService",
	HandlerType: (*KDLServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetData",
			Handler:    _KDLService_GetData_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kdl.proto",
}
