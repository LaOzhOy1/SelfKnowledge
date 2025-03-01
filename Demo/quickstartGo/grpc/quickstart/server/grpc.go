package server

import (
	"google.golang.org/grpc"
	"net"
	"quickstart/config"
)
import proto "quickstart/proto"

//NewGRPCServer
/*
Client 端 protobuf 已经生成
Server 端需要用户进行自定义配置
此文件用于设置grpcServer与业务Service进行注册
*/
//NewNetListener
/*
设置监听socket信息
*/
func NewGRPCServer(srv proto.HelloServiceServer) *grpc.Server {
	// ServerOptions :
	//grpc.StatsHandler(),
	//grpc.UnaryInterceptor()
	server := grpc.NewServer()
	proto.RegisterHelloServiceServer(server, srv)
	// 注册监控服务
	return server
}

func NewNetListener(config *config.Config) (net.Listener, error) {
	return net.Listen("tcp", config.GrpcServerAddr)
}
