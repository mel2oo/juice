package logger

import (
	"fmt"
	"os"
	"path/filepath"
)

type Option func(*Options)

type Options struct {
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	Level             Level
	OutputName        string
	OutputPath        string
	Prefix            string
	MaxSize           int
	MaxBackups        int
	MaxAge            int
	Compress          bool
}

func NewOptions() *Options {
	s := filepath.Base(os.Args[0])
	// t := time.Now().Format("2006-01-02")
	return &Options{
		Development:       true,
		DisableCaller:     false,
		DisableStacktrace: false,
		Level:             DebugLevel,
		OutputPath:        "./",
		OutputName:        fmt.Sprintf("%s.log", s),
		Prefix:            fmt.Sprintf("[%s] ", s),
		MaxSize:           30,
		MaxBackups:        10,
		MaxAge:            7,
		Compress:          false,
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

func WithLevel(l Level) Option {
	return func(o *Options) {
		o.Level = l
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

func WithMaxSize(i int) Option {
	return func(o *Options) {
		o.MaxSize = i
	}
}

func WithMaxBackups(i int) Option {
	return func(o *Options) {
		o.MaxBackups = i
	}
}

func WithMaxAge(i int) Option {
	return func(o *Options) {
		o.MaxAge = i
	}
}

func WithEnableCompress() Option {
	return func(o *Options) {
		o.Compress = true
	}
}
