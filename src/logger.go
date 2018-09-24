package core

type Logger interface {
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string, err error)
	Panic(message string, err error)
}
