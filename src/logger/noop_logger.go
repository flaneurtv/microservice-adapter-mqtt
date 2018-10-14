package logger

import (
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"time"
)

type noopLogger struct {
}

func NewNoOpLogger() core.Logger {
	return &noopLogger{}
}

func (*noopLogger) SetLevel(level core.LogLevel) {
}

func (*noopLogger) SetClient(client core.MessageBusClient, namespace, serviceName, serviceUUID, serviceHost string) {
}

func (*noopLogger) SetCreatedAtGetter(getCreatedAt func() time.Time) {
}

func (*noopLogger) Log(level core.LogLevel, message string) {
}
