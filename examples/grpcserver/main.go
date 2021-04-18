package main

import (
	"context"

	pb "github.com/switch-li/juice/examples/greeter"
	"github.com/switch-li/juice/transport/grpc"
)

type GreeterService struct {
	pb.UnimplementedGreeterServer
}

func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

func (g *GreeterService) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Message: "hello " + request.GetName()}, nil
}

func main() {

	srv := grpc.NewServer(
		grpc.Address(":8890"),
	)

	pb.RegisterGreeterServer(srv, NewGreeterService())

	srv.Start()
}
