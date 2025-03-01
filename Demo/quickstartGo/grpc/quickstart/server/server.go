package server

import (
	"google.golang.org/grpc"
	"net"
)

//Server
/*
	服务器信息总览
*/
type Server struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}
