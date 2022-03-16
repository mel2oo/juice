package juice

import (
	"context"
	"os"

	"github.com/mel2oo/juice/transport"
)

type Option func(o *options)

type options struct {
	ctx     context.Context
	sigs    []os.Signal
	servers []transport.Server
}

func Signal(sigs ...os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}

func Context(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func Server(srv ...transport.Server) Option {
	return func(o *options) {
		o.servers = srv
	}
}
