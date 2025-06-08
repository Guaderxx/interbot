package alog

type ALogger interface {
	Trace(msg string, v ...any)
	Debug(msg string, v ...any)
	Info(msg string, v ...any)
	Warn(msg string, v ...any)
	Error(msg string, v ...any)
	Fatal(msg string, v ...any)
	Panic(msg string, v ...any)

	With(args ...any) ALogger
	WithGroup(name string) ALogger

	// Accept the Gorm
}

type Options struct {
	Formatter string
	Level     string
	AddSource bool `mapstructure:"add_source"`
}
