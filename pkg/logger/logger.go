package logger

type Logger interface {
	String() string
	Type() interface{}

	Log(level Level, v string)

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	DPanic(v ...interface{})
	Panic(v ...interface{})
	Fatal(v ...interface{})
}
