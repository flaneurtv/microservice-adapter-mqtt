package logger_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/logger"
	"strings"
	"testing"
	"time"
)

func TestMQTTLogger(t *testing.T) {
	logOutput := bytes.NewBuffer(nil)
	logError := bytes.NewBuffer(nil)
	log := logger.NewMQTTLogger(logOutput, logError)
	client := &mockClient{}
	log.SetClient(client, "root", "first", "id1", "host.com")
	log.SetLevels(core.LogLevelNotice, core.LogLevelWarning)
	log.SetCreatedAtGetter(func() time.Time {
		return time.Date(2018, 10, 9, 10, 11, 12, 345345345, time.Now().Location())
	})

	log.Log(core.LogLevelError, "error A")
	log.Log(core.LogLevelInfo, "info B")
	log.Log(core.LogLevelNotice, "notice C")
	log.Log(core.LogLevelCritical, "critical D")
	log.Log(core.LogLevelWarning, "warning E")

	assert.Equal(t, 3, len(client.messages))

	assert.Equal(t, `root/log/first/id1/error`, client.messages[0].topic)
	assert.Equal(t, `{"topic":"root/log/first/id1/error","service_name":"first","service_uuid":"id1","service_host":"host.com","created_at":"2018-10-09T10:11:12.345Z","payload":{"log_entry":{"log_level":"error","log_message":"error A"}}}`, client.messages[0].message)

	assert.Equal(t, `root/log/first/id1/critical`, client.messages[1].topic)
	assert.Equal(t, `{"topic":"root/log/first/id1/critical","service_name":"first","service_uuid":"id1","service_host":"host.com","created_at":"2018-10-09T10:11:12.345Z","payload":{"log_entry":{"log_level":"critical","log_message":"critical D"}}}`, client.messages[1].message)

	assert.Equal(t, `root/log/first/id1/warning`, client.messages[2].topic)
	assert.Equal(t, `{"topic":"root/log/first/id1/warning","service_name":"first","service_uuid":"id1","service_host":"host.com","created_at":"2018-10-09T10:11:12.345Z","payload":{"log_entry":{"log_level":"warning","log_message":"warning E"}}}`, client.messages[2].message)

	outputLines := strings.Split(logOutput.String(), "\n")
	assert.Equal(t, 3, len(outputLines))
	assert.Equal(t, "notice: notice C", outputLines[0])
	assert.Equal(t, "warning: warning E", outputLines[1])
	assert.Equal(t, "", outputLines[2])

	errorLines := strings.Split(logError.String(), "\n")
	assert.Equal(t, 3, len(errorLines))
	assert.Equal(t, "error: error A", errorLines[0])
	assert.Equal(t, "critical: critical D", errorLines[1])
	assert.Equal(t, "", errorLines[2])
}

type mockClient struct {
	messages []mqttMessage
}

func (c *mockClient) Connect() error {
	return nil
}

func (c *mockClient) Subscribe(topics []string) (<-chan string, error) {
	return nil, nil
}

func (c *mockClient) Publish(topic, message string) error {
	c.messages = append(c.messages, mqttMessage{topic: topic, message: message})
	return nil
}

type mqttMessage struct {
	topic   string
	message string
}
