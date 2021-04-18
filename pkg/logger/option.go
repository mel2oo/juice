package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Option func(*Options)

type Options struct {
	Development   bool
	DisableCaller bool
	OutputPath    string
	Prefix        string
}

func NewOptions() *Options {
	s := filepath.Base(os.Args[0])
	t := time.Now().Format("2006-01-02")
	return &Options{
		Development:   true,
		DisableCaller: false,
		OutputPath:    fmt.Sprintf("%s_%s.log", s, t),
		Prefix:        fmt.Sprintf("[%s] ", s),
	}
}

func WithDevelopment(b bool) Option {
	return func(o *Options) {
		o.Development = b
	}
}

func DisableCaller() Option {
	return func(o *Options) {
		o.DisableCaller = true
	}
}

func OutputPath(s string) Option {
	return func(o *Options) {
		o.OutputPath = s
	}
}

func Prefix(s string) Option {
	return func(o *Options) {
		o.Prefix = s + " "
	}
}
