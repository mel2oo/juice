package grpc_logging

import (
	"context"
	"time"

	"github.com/mel2oo/juice/pkg/logger"
	"github.com/mel2oo/juice/transport/grpc/middleware/logging/ctxzap"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
)

var (
	defaultOptions = &options{
		levelFunc:       DefaultCodeToLevel,
		shouldLog:       DefaultDeciderMethod,
		codeFunc:        DefaultErrorToCode,
		durationFunc:    DefaultDurationToField,
		messageFunc:     DefaultMessageProducer,
		timestampFormat: time.RFC3339,
	}
)

type options struct {
	levelFunc       CodeToLevel
	shouldLog       Decider
	codeFunc        ErrorToCode
	durationFunc    DurationToField
	messageFunc     MessageProducer
	timestampFormat string
}

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateClientOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultClientCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) logger.Level

// DurationToField function defines how to produce duration fields for logging
type DurationToField func(duration time.Duration) zapcore.Field

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log.
func WithDecider(f Decider) Option {
	return func(o *options) {
		o.shouldLog = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// WithDurationField customizes the function for mapping request durations to Zap fields.
func WithDurationField(f DurationToField) Option {
	return func(o *options) {
		o.durationFunc = f
	}
}

// WithMessageProducer customizes the function for message formation.
func WithMessageProducer(f MessageProducer) Option {
	return func(o *options) {
		o.messageFunc = f
	}
}

// WithTimestampFormat customizes the timestamps emitted in the log fields.
func WithTimestampFormat(format string) Option {
	return func(o *options) {
		o.timestampFormat = format
	}
}

// DefaultCodeToLevel is the default implementation of gRPC return codes and interceptor log level for server side.
func DefaultCodeToLevel(code codes.Code) logger.Level {
	switch code {
	case codes.OK:
		return logger.InfoLevel
	case codes.Canceled:
		return logger.InfoLevel
	case codes.Unknown:
		return logger.ErrorLevel
	case codes.InvalidArgument:
		return logger.InfoLevel
	case codes.DeadlineExceeded:
		return logger.WarnLevel
	case codes.NotFound:
		return logger.InfoLevel
	case codes.AlreadyExists:
		return logger.InfoLevel
	case codes.PermissionDenied:
		return logger.WarnLevel
	case codes.Unauthenticated:
		return logger.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return logger.WarnLevel
	case codes.FailedPrecondition:
		return logger.WarnLevel
	case codes.Aborted:
		return logger.WarnLevel
	case codes.OutOfRange:
		return logger.WarnLevel
	case codes.Unimplemented:
		return logger.ErrorLevel
	case codes.Internal:
		return logger.ErrorLevel
	case codes.Unavailable:
		return logger.WarnLevel
	case codes.DataLoss:
		return logger.ErrorLevel
	default:
		return logger.ErrorLevel
	}
}

// DefaultClientCodeToLevel is the default implementation of gRPC return codes to log levels for client side.
func DefaultClientCodeToLevel(code codes.Code) logger.Level {
	switch code {
	case codes.OK:
		return logger.DebugLevel
	case codes.Canceled:
		return logger.DebugLevel
	case codes.Unknown:
		return logger.InfoLevel
	case codes.InvalidArgument:
		return logger.DebugLevel
	case codes.DeadlineExceeded:
		return logger.InfoLevel
	case codes.NotFound:
		return logger.DebugLevel
	case codes.AlreadyExists:
		return logger.DebugLevel
	case codes.PermissionDenied:
		return logger.InfoLevel
	case codes.Unauthenticated:
		return logger.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return logger.DebugLevel
	case codes.FailedPrecondition:
		return logger.DebugLevel
	case codes.Aborted:
		return logger.DebugLevel
	case codes.OutOfRange:
		return logger.DebugLevel
	case codes.Unimplemented:
		return logger.WarnLevel
	case codes.Internal:
		return logger.WarnLevel
	case codes.Unavailable:
		return logger.WarnLevel
	case codes.DataLoss:
		return logger.WarnLevel
	default:
		return logger.InfoLevel
	}
}

// DefaultDurationToField is the default implementation of converting request duration to a Zap field.
var DefaultDurationToField = DurationToTimeMillisField

// DurationToTimeMillisField converts the duration to milliseconds and uses the key `grpc.time_ms`.
func DurationToTimeMillisField(duration time.Duration) zapcore.Field {
	return zap.Float32("grpc.time_ms", durationToMilliseconds(duration))
}

// DurationToDurationField uses a Duration field to log the request duration
// and leaves it up to Zap's encoder settings to determine how that is output.
func DurationToDurationField(duration time.Duration) zapcore.Field {
	return zap.Duration("grpc.duration", duration)
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}

// MessageProducer produces a user defined log message
type MessageProducer func(ctx context.Context, msg string, level logger.Level, code codes.Code, err error, duration zapcore.Field)

// DefaultMessageProducer writes the default message
func DefaultMessageProducer(ctx context.Context, msg string, level logger.Level, code codes.Code, err error, duration zapcore.Field) {
	// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
	// ctxzap.Extract(ctx).Check(level, msg).Write(
	// 	zap.Error(err),
	// 	zap.String("grpc.code", code.String()),
	// 	duration,
	// )

	ctxzap.Extract(ctx).Sugar().Infof("hello world -- cc")
}
