package zap

import (
	"io"
	"os"
	"sort"
	"time"

	"github.com/switch-li/juice/pkg/logger"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	*zap.Logger
	opts *logger.Options
}

func NewZapLogger(opts ...logger.Option) logger.Logger {
	options := logger.NewOptions()

	for _, o := range opts {
		o(options)
	}

	config := Config{}

	if options.Development {
		config = newDevelopment(options)
	} else {
		config = newProduction(options)
	}

	zopts := make([]zap.Option, 0, 1)

	logger := config.Build(zopts...)

	return &ZapLogger{
		logger,
		options,
	}
}

func newDevelopment(opts *logger.Options) Config {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var w io.Writer
	w = os.Stdout
	sink := zapcore.AddSync(w)

	return Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       opts.Development,
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		Encoder:           zapcore.NewConsoleEncoder(encoderConfig),
		WriteSyncer:       sink,
	}
}

func newProduction(opts *logger.Options) Config {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	//https://github.com/natefinch/lumberjack
	var w io.Writer
	w = os.Stdout
	if opts.OutputPath != "" {
		w = &lumberjack.Logger{
			Filename:   opts.OutputPath + opts.OutputName,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     14,   //days
			Compress:   true, // disabled by default
		}
	}
	sink := zapcore.AddSync(w)

	return Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       opts.Development,
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		Encoder:           zapcore.NewConsoleEncoder(encoderConfig), //NewJSONEncoder
		WriteSyncer:       sink,
	}
}

func (z *ZapLogger) ReplaceGlobals() {
	zap.ReplaceGlobals(z.Logger)
}

func (z *ZapLogger) Type() interface{} {
	return z.Logger
}

func (z *ZapLogger) String() string {
	return "zap"
}

func (z *ZapLogger) Log(level logger.Level, msg string) {
	switch level {
	case logger.DebugLevel:
		z.Debug(msg)
	case logger.InfoLevel:
		z.Info(msg)
	case logger.WarnLevel:
		z.Warn(msg)
	case logger.ErrorLevel:
		z.Error(msg)
	case logger.DPanicLevel:
		z.DPanic(msg)
	case logger.PanicLevel:
		z.Panic(msg)
	case logger.FatalLevel:
		z.Fatal(msg)
	}
}

func toZapLevel(l logger.Level) zapcore.Level {
	var zl zapcore.Level
	switch l {
	case logger.DebugLevel:
		zl = zapcore.DebugLevel
	case logger.InfoLevel:
		zl = zapcore.InfoLevel
	case logger.WarnLevel:
		zl = zapcore.WarnLevel
	case logger.ErrorLevel:
		zl = zapcore.ErrorLevel
	case logger.DPanicLevel:
		zl = zapcore.DPanicLevel
	case logger.PanicLevel:
		zl = zapcore.PanicLevel
	case logger.FatalLevel:
		zl = zapcore.FatalLevel
	}
	return zl
}

func (z *ZapLogger) Debug(v ...interface{}) {
	z.Logger.Sugar().Named(z.opts.Prefix).Debug(v)
}

func (z *ZapLogger) Info(v ...interface{}) {
	z.Logger.Sugar().Named(z.opts.Prefix).Info(v)
}

func (z *ZapLogger) Warn(v ...interface{}) {
	z.Logger.Sugar().Named(z.opts.Prefix).Warn(v)
}

func (z *ZapLogger) Error(v ...interface{}) {
	z.Logger.Sugar().Named(z.opts.Prefix).Error(v)
}

func (z *ZapLogger) DPanic(v ...interface{}) {
	z.Logger.Sugar().Named(z.opts.Prefix).DPanic(v)
}

func (z *ZapLogger) Panic(v ...interface{}) {
	z.Logger.Sugar().Named(z.opts.Prefix).Panic(v)
}

func (z *ZapLogger) Fatal(v ...interface{}) {
	z.Logger.Sugar().Named(z.opts.Prefix).Fatal(v)
}

type Config struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling Config.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config.
	Level zap.AtomicLevel `json:"level" yaml:"level"`
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"development" yaml:"development"`
	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `json:"disableCaller" yaml:"disableCaller"`
	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktrace bool `json:"disableStacktrace" yaml:"disableStacktrace"`
	// Sampling sets a sampling policy. A nil SamplingConfig disables sampling.
	Sampling *zap.SamplingConfig `json:"sampling" yaml:"sampling"`
	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoder zapcore.Encoder

	WriteSyncer zapcore.WriteSyncer

	InitialFields map[string]interface{} `json:"initialFields" yaml:"initialFields"`
}

func (cfg Config) Build(opts ...zap.Option) *zap.Logger {
	core := zapcore.NewCore(cfg.Encoder, cfg.WriteSyncer, cfg.Level)

	log := zap.New(core, cfg.buildOptions()...)
	if len(opts) > 0 {
		log = log.WithOptions(opts...)
	}

	return log
}

func (cfg Config) buildOptions() []zap.Option {
	opts := make([]zap.Option, 0, 4)

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
		opts = append(opts, zap.AddCallerSkip(1))
	}

	//stacktrace
	stackLevel := zap.ErrorLevel
	if cfg.Development {
		stackLevel = zap.WarnLevel
	}
	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
		}))
	}

	if len(cfg.InitialFields) > 0 {
		fs := make([]zap.Field, 0, len(cfg.InitialFields))
		keys := make([]string, 0, len(cfg.InitialFields))
		for k := range cfg.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, cfg.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	return opts
}
