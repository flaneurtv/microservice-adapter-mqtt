package env_test

import (
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/env"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestCorrectConfig(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"SERVICE_NAME":               "MyService",
		"SERVICE_PROCESSOR":          serviceProcessorFile.Name() + " run",
		"NAMESPACE":                  "root",
		"NAMESPACE_LISTENER":         "master",
		"MQTT_LISTENER_URL":          "tcp://mqtt.com:111",
		"MQTT_PUBLISHER_URL":         "tcp://mqtt.com:222",
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SUBSCRIPTIONS":              subscriptionsFile.Name(),
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, "MyService", cfg.ServiceName())
	assert.NotEmpty(t, cfg.ServiceUUID())
	assert.NotEmpty(t, cfg.ServiceHost())
	assert.Equal(t, serviceProcessorFile.Name()+" run", cfg.ServiceCmdLine())
	assert.Equal(t, "master", cfg.NamespaceListener())
	assert.Equal(t, "root", cfg.NamespacePublisher())
	assert.Equal(t, "tcp://mqtt.com:111", cfg.ListenerURL())
	assert.Equal(t, "tcp://mqtt.com:222", cfg.PublisherURL())
	assert.Equal(t, core.Credentials{UserName: "user111", Password: "password111"}, cfg.ListenerCredentials())
	assert.Equal(t, core.Credentials{UserName: "user222", Password: "password222"}, cfg.PublisherCredentials())
	assert.Equal(t, []string{"master/test", "master/clean", "master/good"}, cfg.Subscriptions())
	assert.Equal(t, "error", cfg.LogLevelConsole())
	assert.Equal(t, "error", cfg.LogLevelRemote())
}

func TestDefaultNamespace(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SUBSCRIPTIONS":              subscriptionsFile.Name(),
		"SERVICE_PROCESSOR":          serviceProcessorFile.Name(),
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, "default", cfg.NamespaceListener())
	assert.Equal(t, "default", cfg.NamespacePublisher())
	assert.Equal(t, []string{"default/test", "default/clean", "default/good"}, cfg.Subscriptions())
}

func TestEmptyListenerCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SERVICE_PROCESSOR":          serviceProcessorFile.Name(),
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
}

func TestEmptyPublisherCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS": listenerCredentialsFile.Name(),
		"SERVICE_PROCESSOR":         serviceProcessorFile.Name(),
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
}

func TestMissingPublisherCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name() + uuid.NewV4().String(),
		"SERVICE_PROCESSOR":          serviceProcessorFile.Name(),
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
}

func TestMissingListenerCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS": listenerCredentialsFile.Name() + uuid.NewV4().String(),
		"SERVICE_PROCESSOR":         serviceProcessorFile.Name(),
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
}

func TestMissingSubscriptions(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SUBSCRIPTIONS":              fmt.Sprintf("/dummy/subscriptions_%d", time.Now().UnixNano()),
		"SERVICE_PROCESSOR":          serviceProcessorFile.Name(),
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
}

func TestEmptySubscriptions(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SERVICE_PROCESSOR":          serviceProcessorFile.Name(),
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cfg.Subscriptions()))
}

func TestCorruptedCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"`)
	publisherCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SERVICE_PROCESSOR":          serviceProcessorFile.Name(),
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.NotNil(t, err)
}

func TestLogLevel(t *testing.T) {
	clearEnv()
	defer clearEnv()

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": serviceProcessorFile.Name(),
		"LOG_LEVEL":         "warning",
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, "warning", cfg.LogLevelConsole())
	assert.Equal(t, "warning", cfg.LogLevelRemote())
}

func TestLogLevelConsole(t *testing.T) {
	clearEnv()
	defer clearEnv()

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": serviceProcessorFile.Name(),
		"LOG_LEVEL":         "warning",
		"LOG_LEVEL_CONSOLE": "info",
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, "info", cfg.LogLevelConsole())
	assert.Equal(t, "warning", cfg.LogLevelRemote())
}

func TestLogLevelRemote(t *testing.T) {
	clearEnv()
	defer clearEnv()

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": serviceProcessorFile.Name(),
		"LOG_LEVEL":         "warning",
		"LOG_LEVEL_MQTT":    "info",
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, "info", cfg.LogLevelRemote())
	assert.Equal(t, "warning", cfg.LogLevelConsole())
}

func TestLogLevelConsoleRemote(t *testing.T) {
	clearEnv()
	defer clearEnv()

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": serviceProcessorFile.Name(),
		"LOG_LEVEL_CONSOLE": "debug",
		"LOG_LEVEL_MQTT":    "info",
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, "debug", cfg.LogLevelConsole())
	assert.Equal(t, "info", cfg.LogLevelRemote())
}

func TestServiceProcessorEmpty(t *testing.T) {
	clearEnv()
	defer clearEnv()

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": "",
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SERVICE_PROCESSOR can't be empty")
}

func TestServiceProcessorMissing(t *testing.T) {
	clearEnv()
	defer clearEnv()

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": fmt.Sprintf("/dummy/processor_%d", time.Now().UnixNano()),
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestServiceProcessorDirectory(t *testing.T) {
	clearEnv()
	defer clearEnv()

	serviceProcessorFile, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(serviceProcessorFile)

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": serviceProcessorFile,
	})

	_, err := env.NewAdapterConfig(&mockLogger{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "should reference to a file")
}

func TestDefaultValues(t *testing.T) {
	clearEnv()
	defer clearEnv()

	serviceProcessorFile, _ := ioutil.TempFile("", "")
	defer os.Remove(serviceProcessorFile.Name())

	setEnv(map[string]string{
		"SERVICE_PROCESSOR": serviceProcessorFile.Name(),
	})

	cfg, err := env.NewAdapterConfig(&mockLogger{})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(cfg.Subscriptions()))
	assert.Equal(t, core.Credentials{}, cfg.ListenerCredentials())
	assert.Equal(t, core.Credentials{}, cfg.PublisherCredentials())
	assert.Equal(t, "tcp://mqtt:1883", cfg.ListenerURL())
	assert.Equal(t, "tcp://mqtt:1883", cfg.PublisherURL())
}

func setEnv(env map[string]string) {
	for key, value := range env {
		os.Setenv(key, value)
	}
}

func clearEnv() {
	os.Unsetenv("SERVICE_NAME")
	os.Unsetenv("SERVICE_PROCESSOR")
	os.Unsetenv("NAMESPACE")
	os.Unsetenv("NAMESPACE_LISTENER")
	os.Unsetenv("NAMESPACE_PUBLISHER")
	os.Unsetenv("MQTT_LISTENER_URL")
	os.Unsetenv("MQTT_PUBLISHER_URL")
	os.Unsetenv("MQTT_LISTENER_CREDENTIALS")
	os.Unsetenv("MQTT_PUBLISHER_CREDENTIALS")
	os.Unsetenv("SUBSCRIPTIONS")
	os.Unsetenv("DEBUG")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_LEVEL_CONSOLE")
	os.Unsetenv("LOG_LEVEL_MQTT")
}

type mockLogger struct {
	messages []mockLoggerMessage
}

func (*mockLogger) SetClient(client core.MessageBusClient, namespace, serviceName, serviceUUID, serviceHost string) {
}

func (*mockLogger) SetLevels(levelConsole, levelRemote core.LogLevel) {
}

func (*mockLogger) SetCreatedAtGetter(getCreatedAt func() time.Time) {
}

func (log *mockLogger) Log(level core.LogLevel, message string) {
	log.messages = append(log.messages, mockLoggerMessage{level: level, message: message})
}

func (log *mockLogger) clear() {
	log.messages = nil
}

type mockLoggerMessage struct {
	level   core.LogLevel
	message string
}
