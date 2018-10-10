package logger

import (
	"fmt"
	"github.com/tidwall/sjson"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"io"
	"time"
)

type mqttLogger struct {
	output io.Writer
	error  io.Writer

	level       core.LogLevel
	client      core.MessageBusClient
	namespace   string
	serviceName string
	serviceUUID string
	serviceHost string

	getCreatedAt func() time.Time
}

func NewMQTTLogger(output, error io.Writer) core.Logger {
	return &mqttLogger{output: output, error: error}
}

func (logger *mqttLogger) SetLevel(level core.LogLevel) {
	logger.level = level
}

func (logger *mqttLogger) SetClient(client core.MessageBusClient, namespace, serviceName, serviceUUID, serviceHost string) {
	logger.client = client
	logger.namespace = namespace
	logger.serviceName = serviceName
	logger.serviceUUID = serviceUUID
	logger.serviceHost = serviceHost
}

func (logger *mqttLogger) SetCreatedAtGetter(getCreatedAt func() time.Time) {
	logger.getCreatedAt = getCreatedAt
}

func (logger *mqttLogger) Log(level core.LogLevel, message string) {
	if level.IsWeaker(logger.level) {
		return
	}

	var out io.Writer
	if level.IsWeaker(core.LogLevelError) {
		out = logger.output
	} else {
		out = logger.error
	}
	_, _ = fmt.Fprintf(out, fmt.Sprintf("%s: %s\n", level, message))

	if logger.client != nil {
		topic, jsonMessage := logger.generateDebugMessage(level, message)
		err := logger.client.Publish(topic, jsonMessage)
		if err != nil {
			_, _ = fmt.Fprintf(out, fmt.Sprintf("error: can't publish a log message: %s\n", jsonMessage))
		}
	}
}

func (logger *mqttLogger) generateDebugMessage(level core.LogLevel, message string) (topic string, jsonMessage string) {
	var createdAt time.Time
	if logger.getCreatedAt != nil {
		createdAt = logger.getCreatedAt()
	} else {
		createdAt = time.Now().UTC()
	}

	topic = fmt.Sprintf("%s/log/%s/%s/%s", logger.namespace, logger.serviceName, logger.serviceUUID, level)
	jsonMessage, _ = sjson.Set(jsonMessage, "payload.log_entry.log_message", message)
	jsonMessage, _ = sjson.Set(jsonMessage, "payload.log_entry.log_level", string(level))
	jsonMessage, _ = sjson.Set(jsonMessage, "created_at", createdAt.Format("2006-01-02T15:04:05.000Z"))
	jsonMessage, _ = sjson.Set(jsonMessage, "service_host", logger.serviceHost)
	jsonMessage, _ = sjson.Set(jsonMessage, "service_uuid", logger.serviceUUID)
	jsonMessage, _ = sjson.Set(jsonMessage, "service_name", logger.serviceName)
	jsonMessage, _ = sjson.Set(jsonMessage, "topic", topic)
	return topic, jsonMessage
}
