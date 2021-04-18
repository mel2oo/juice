package main

import (
	"context"

	"github.com/switch-li/juice"
	"github.com/switch-li/juice/examples/greeter"
	lzap "github.com/switch-li/juice/pkg/logger/zap"
	"github.com/switch-li/juice/pkg/shutdown"
	"github.com/switch-li/juice/transport/grpc"
	"github.com/switch-li/juice/transport/grpc/middleware"
	grcp_logging "github.com/switch-li/juice/transport/grpc/middleware/logging"
)

func main() {
	app := juice.NewApp(
		juice.Server(
			// NewHTTPServer(),
			NewGRPCServer(),
		),
	)

	app.Run()

	shutdown.NewHook().Close(
		func() {
			app.Stop()
		},
	)
}

// func NewHTTPServer() *http.Server {
// 	log := lzap.NewZapLogger(
// 	// zlog.WithDevelopment(false)
// 	)
// 	srv := http.NewServer(
// 		transport.Address(":8880"),
// 		transport.Logger(log),
// 		transport.HTTPMiddleware(
// 			http_logger.Use(log),
// 		),
// 	)

// 	srv.Handle("/hello", http2.HandlerFunc(func(w http2.ResponseWriter, r *http2.Request) {
// 		fmt.Println("hello")
// 		fmt.Fprint(w, "hello")
// 	}))

// 	return srv
// }

type GreeterService struct {
	greeter.UnimplementedGreeterServer
}

func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

func (s *GreeterService) SayHello(ctx context.Context, request *greeter.HelloRequest) (*greeter.HelloResponse, error) {
	return &greeter.HelloResponse{Message: "hello"}, nil
}

func NewGRPCServer() *grpc.Server {
	log := lzap.NewZapLogger()

	srv := grpc.NewServer(
		grpc.Address(":8890"),
		grpc.Logger(log),
		grpc.Middleware(
			middleware.ChainUnaryServer(
				grcp_logging.UnaryServerInterceptor(log),
			),
		),
	)

	greeter.RegisterGreeterServer(srv, NewGreeterService())

	return srv
}
