package grpc

import (
	"net"
	"time"

	"github.com/switch-li/juice/pkg/logger"
	dlog "github.com/switch-li/juice/pkg/logger/default"
	"github.com/switch-li/juice/transport/grpc/middleware"
	logging "github.com/switch-li/juice/transport/grpc/middleware/logging"
	"google.golang.org/grpc"
)

type ServerOption func(*Server)

func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func Logger(log logger.Logger) ServerOption {
	return func(s *Server) {
		s.log = log
	}
}

func Middleware(gm grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.middleware = gm
	}
}

type Server struct {
	*grpc.Server
	// lis        net.Listener
	network    string
	address    string
	timeout    time.Duration
	log        logger.Logger
	middleware grpc.UnaryServerInterceptor
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: ":",
		timeout: time.Second * 5,
		log:     dlog.DefaultLogger,
	}

	for _, o := range opts {
		o(srv)
	}

	srv.middleware = middleware.ChainUnaryServer(
		logging.UnaryServerInterceptor(srv.log),
	)

	srv.Server = grpc.NewServer(grpc.UnaryInterceptor(srv.middleware))

	return srv
}

func (s *Server) Start() error {
	lis, err := net.Listen(s.network, s.address)
	if err != nil {
		s.log.Error(err)
		return err
	}

	// s.lis = lis

	s.log.Info("grpc server listen on:", lis.Addr().String())

	return s.Serve(lis)
}

func (s *Server) Stop() error {
	s.GracefulStop()
	return nil
}
