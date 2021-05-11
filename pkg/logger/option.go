package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Option func(*Options)

type Options struct {
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	OutputName        string
	OutputPath        string
	Prefix            string
}

func NewOptions() *Options {
	s := filepath.Base(os.Args[0])
	t := time.Now().Format("2006-01-02")
	return &Options{
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputName:        "./",
		OutputPath:        fmt.Sprintf("%s_%s.log", s, t),
		Prefix:            fmt.Sprintf("[%s] ", s),
	}
}

func WithDevelopment() Option {
	return func(o *Options) {
		o.Development = true
	}
}

func WithDisableCaller() Option {
	return func(o *Options) {
		o.DisableCaller = true
	}
}

func WithDisableStacktrace() Option {
	return func(o *Options) {
		o.DisableStacktrace = true
	}
}

func WithOutputName(s string) Option {
	return func(o *Options) {
		o.OutputName = s
	}
}

func WithOutputPath(s string) Option {
	return func(o *Options) {
		o.OutputPath = s
	}
}

func WithPrefix(s string) Option {
	return func(o *Options) {
		o.Prefix = s + " "
	}
}
