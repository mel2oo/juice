package http

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/switch-li/juice/pkg/logger"
	dlog "github.com/switch-li/juice/pkg/logger/default"
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

func TLS(cafile, certfile, keyfile string) ServerOption {
	return func(s *Server) {
		s.tls = true
		s.caFile = cafile
		s.certFile = certfile
		s.keyFile = keyfile
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

type Server struct {
	*http.Server
	network  string
	address  string
	tls      bool
	caFile   string
	certFile string
	keyFile  string
	timeout  time.Duration
	log      logger.Logger
	exit     chan chan error
}

func NewServer(mux *Mux, opts ...ServerOption) *Server {
	srv := &Server{
		Server: &http.Server{
			Handler: mux,
		},
		network: "tcp",
		address: ":",
		timeout: time.Second * 5,
		log:     dlog.DefaultLogger,
		exit:    make(chan chan error),
	}

	for _, o := range opts {
		o(srv)
	}

	if srv.tls {
		pool := x509.NewCertPool()
		caCrt, err := ioutil.ReadFile(srv.caFile)
		if err != nil {
			return nil
		}
		pool.AppendCertsFromPEM(caCrt)

		srv.TLSConfig = &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
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
