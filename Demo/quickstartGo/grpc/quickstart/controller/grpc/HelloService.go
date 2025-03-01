package grpc

import (
	"context"
	"fmt"
)
import proto "quickstart/proto"

type HelloService struct {
	proto.UnimplementedHelloServiceServer
	// services
	//			-> repository
}

func (this *HelloService) SayHello(context context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	fmt.Printf("receive a request pid is %d\n", req.Pid)
	return &proto.HelloResponse{Result: 1}, nil
}
