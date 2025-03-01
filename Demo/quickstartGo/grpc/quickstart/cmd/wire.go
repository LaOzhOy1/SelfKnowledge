//go:build wireinject
// +build wireinject

package cmd

import (
	"github.com/google/wire"
	"quickstart/config"
	grpc_server "quickstart/controller/grpc"
	proto "quickstart/proto"
	"quickstart/server"
)

var applicationContext = wire.NewSet(
	wire.Struct(new(proto.UnimplementedHelloServiceServer), "*"),
	config.NewConfig,
	wire.Struct(new(grpc_server.HelloService), "*"),
	wire.Bind(new(proto.HelloServiceServer), new(*grpc_server.HelloService)),
	server.NewNetListener,
	server.NewGRPCServer,

	wire.Struct(new(server.Server), "*"),
)

func InitializeServer() (*server.Server, error) {
	panic(wire.Build(applicationContext))
}
