package transport

import "context"

type Server interface {
	Start() error
	Stop() error
}

type Kind string

const (
	GRPC Kind = "GRPC"
	HTTP Kind = "HTTP"
)

type transportKey struct{}

func NewContext(ctx context.Context, tr Kind) context.Context {
	return context.WithValue(ctx, transportKey{}, tr)
}

func FromContext(ctx context.Context) (tr Kind, ok bool) {
	tr, ok = ctx.Value(transportKey{}).(Kind)
	return
}
