package logger

type Logger interface {
	String() string
	Type() interface{}

	Log(level Level, v string)

	Debug(v ...interface{})
	Debugf(format string, args ...interface{})
	Info(v ...interface{})
	Infof(format string, args ...interface{})
	Warn(v ...interface{})
	Warnf(format string, args ...interface{})
	Error(v ...interface{})
	Errorf(format string, args ...interface{})
	DPanic(v ...interface{})
	Panic(v ...interface{})
	Fatal(v ...interface{})
}
