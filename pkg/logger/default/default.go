package dlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/switch-li/juice/pkg/logger"
)

var DefaultLogger logger.Logger = NewDefaultLogger(
	logger.WithDevelopment(),
	logger.WithDisableCaller(),
)

type defaultLogger struct {
	log  *log.Logger
	opts *logger.Options
}

func NewDefaultLogger(opts ...logger.Option) logger.Logger {
	options := logger.NewOptions()

	for _, o := range opts {
		o(options)
	}

	if options.Development {
		return newDevelopment(options)
	} else {
		return newProduction(options)
	}
}

func newDevelopment(opts *logger.Options) logger.Logger {
	w := os.Stdout
	return &defaultLogger{
		log.New(w, opts.Prefix, log.LstdFlags),
		opts,
	}
}

func newProduction(opts *logger.Options) logger.Logger {

	logFile, err := os.OpenFile(opts.OutputPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}

	return &defaultLogger{
		log.New(logFile, opts.Prefix, log.Ldate),
		opts,
	}
}

func (d *defaultLogger) Type() interface{} {
	return d
}

func (d *defaultLogger) Log(level logger.Level, msg string) {
	l := level.String()
	if d.opts.DisableCaller {
		d.log.Printf("%s %v\n", l, msg)
	} else {
		if _, file, line, ok := runtime.Caller(2); ok {
			caller := fmt.Sprintf("%s:%d", filepath.Base(file), line)
			d.log.Printf("%s %s %v\n", l, caller, msg)
		}
	}
}

func (d *defaultLogger) Logf(level logger.Level, format string, v ...interface{}) {
	l := level.String()
	m := fmt.Sprintf(format, v...)
	if d.opts.DisableCaller {
		d.log.Printf("%s %v\n", l, m)
	} else {
		if _, file, line, ok := runtime.Caller(2); ok {
			caller := fmt.Sprintf("%s:%d", filepath.Base(file), line)
			d.log.Printf("%s %s %v\n", l, caller, m)
		}
	}
}

func (d *defaultLogger) Debug(v ...interface{}) {
	d.Log(logger.DebugLevel, fmt.Sprint(v...))
}

func (d *defaultLogger) Info(v ...interface{}) {
	d.Log(logger.InfoLevel, fmt.Sprint(v...))
}

func (d *defaultLogger) Warn(v ...interface{}) {
	d.Log(logger.WarnLevel, fmt.Sprint(v...))
}

func (d *defaultLogger) Error(v ...interface{}) {
	d.Log(logger.ErrorLevel, fmt.Sprint(v...))
}

func (d *defaultLogger) DPanic(v ...interface{}) {
	d.Log(logger.DPanicLevel, fmt.Sprint(v...))
}

func (d *defaultLogger) Panic(v ...interface{}) {
	d.Log(logger.PanicLevel, fmt.Sprint(v...))
}

func (d *defaultLogger) Fatal(v ...interface{}) {
	d.Log(logger.FatalLevel, fmt.Sprint(v...))
}

func (d *defaultLogger) String() string {
	return "default"
}
