package main

import (
	"context"

	pb "github.com/mel2oo/juice/examples/greeter"
	"github.com/mel2oo/juice/pkg/logger"
	"github.com/mel2oo/juice/pkg/logger/zap"
	"github.com/mel2oo/juice/transport/grpc"
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
	log := zap.NewZapLogger(
		logger.WithDevelopment(),
	)

	srv := grpc.NewServer(
		grpc.Address(":8890"),
		grpc.Logger(log),
	)

	pb.RegisterGreeterServer(srv, NewGreeterService())

	srv.Start()
}
