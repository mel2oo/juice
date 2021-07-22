package http

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/switch-li/juice/pkg/logger"
	"github.com/switch-li/juice/pkg/logger/zap"
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

func TLSCert(certfile, keyfile string) ServerOption {
	return func(s *Server) {
		s.tls = true
		s.certFile = certfile
		s.keyFile = keyfile
	}
}

func TLSConfig(cnf *tls.Config) ServerOption {
	return func(s *Server) {
		s.TLSConfig = cnf
	}
}

// func WithTimeout(timeout time.Duration) ServerOption {
// 	return func(s *Server) {
// 		s.timeout = timeout
// 	}
// }

func Logger(log logger.Logger) ServerOption {
	return func(s *Server) {
		s.log = log
	}
}

type Server struct {
	*http.Server
	network  string
	address  string
	tls      bool
	certFile string
	keyFile  string
	// timeout  time.Duration
	log  logger.Logger
	exit chan chan error
}

func NewServer(mux *Mux, opts ...ServerOption) *Server {
	srv := &Server{
		Server: &http.Server{
			Handler: mux,
		},
		network: "tcp",
		address: ":",
		// timeout: time.Second * 5,
		log:  zap.DefaultLogger,
		exit: make(chan chan error),
	}

	for _, o := range opts {
		o(srv)
	}

	return srv
}

func (s *Server) Start() error {
	lis, err := net.Listen(s.network, s.address)
	if err != nil {
		s.log.Error(err)
		return err
	}

	s.log.Info("http server listen on:", lis.Addr().String())

	go func() {
		ch := <-s.exit
		ch <- lis.Close()
	}()

	if s.tls {
		return s.ServeTLS(lis, s.certFile, s.keyFile)
	}

	return s.Serve(lis)
}

func (s *Server) Stop() error {
	ch := make(chan error)
	s.exit <- ch
	return <-ch
}
