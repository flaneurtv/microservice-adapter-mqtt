package logger

import "gitlab.com/flaneurtv/microservice-adapter-mqtt/core"

type noopLogger struct {
}

func NewNoOpLogger() core.Logger {
	return &noopLogger{}
}

func (*noopLogger) Debug(message string) {
}

func (*noopLogger) Info(message string) {
}

func (*noopLogger) Warn(message string) {
}

func (*noopLogger) Error(message string, err error) {
}

func (*noopLogger) Panic(message string, err error) {
}
