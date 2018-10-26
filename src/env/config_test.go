package env_test

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/env"
	"io/ioutil"
	"os"
	"testing"
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

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"SERVICE_NAME":               "MyService",
		"SERVICE_PROCESSOR":          "./my/service run",
		"NAMESPACE":                  "root",
		"NAMESPACE_LISTENER":         "master",
		"MQTT_LISTENER_URL":          "tcp://mqtt.com:111",
		"MQTT_PUBLISHER_URL":         "tcp://mqtt.com:222",
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SUBSCRIPTIONS":              subscriptionsFile.Name(),
	})

	cfg, err := env.NewConfig()
	assert.Nil(t, err)
	assert.Equal(t, "MyService", cfg.ServiceName())
	assert.NotEmpty(t, cfg.ServiceUUID())
	assert.NotEmpty(t, cfg.ServiceHost())
	assert.Equal(t, "./my/service run", cfg.ServiceCmdLine())
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
	})

	cfg, err := env.NewConfig()
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

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SUBSCRIPTIONS":              subscriptionsFile.Name(),
	})

	_, err := env.NewConfig()
	assert.Nil(t, err)
}

func TestEmptyPublisherCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS": listenerCredentialsFile.Name(),
		"SUBSCRIPTIONS":             subscriptionsFile.Name(),
	})

	_, err := env.NewConfig()
	assert.Nil(t, err)
}

func TestMissingPublisherCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name() + uuid.NewV4().String(),
		"SUBSCRIPTIONS":              subscriptionsFile.Name(),
	})

	_, err := env.NewConfig()
	assert.NotNil(t, err)
}

func TestMissingListenerCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS": listenerCredentialsFile.Name() + uuid.NewV4().String(),
		"SUBSCRIPTIONS":             subscriptionsFile.Name(),
	})

	_, err := env.NewConfig()
	assert.NotNil(t, err)
}

func TestMissingSubscriptions(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"}`)
	publisherCredentialsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
	})

	_, err := env.NewConfig()
	assert.NotNil(t, err)
}

func TestCorruptedCredentials(t *testing.T) {
	clearEnv()
	defer clearEnv()

	listenerCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(listenerCredentialsFile.Name())

	publisherCredentialsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(publisherCredentialsFile.Name())

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	listenerCredentialsFile.WriteString(`{"username": "user111", "password": "password111"}`)
	listenerCredentialsFile.Close()
	publisherCredentialsFile.WriteString(`{"username": "user222", "password": "password222"`)
	publisherCredentialsFile.Close()
	subscriptionsFile.WriteString("test\nclean\ngood\n")
	subscriptionsFile.Close()

	setEnv(map[string]string{
		"MQTT_LISTENER_CREDENTIALS":  listenerCredentialsFile.Name(),
		"MQTT_PUBLISHER_CREDENTIALS": publisherCredentialsFile.Name(),
		"SUBSCRIPTIONS":              subscriptionsFile.Name(),
	})

	_, err := env.NewConfig()
	assert.NotNil(t, err)
}

func TestLogLevel(t *testing.T) {
	clearEnv()
	defer clearEnv()

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	setEnv(map[string]string{
		"SUBSCRIPTIONS": subscriptionsFile.Name(),
		"LOG_LEVEL":     "warning",
	})

	cfg, err := env.NewConfig()
	assert.Nil(t, err)
	assert.Equal(t, "warning", cfg.LogLevelConsole())
	assert.Equal(t, "warning", cfg.LogLevelRemote())
}

func TestLogLevelConsole(t *testing.T) {
	clearEnv()
	defer clearEnv()

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	setEnv(map[string]string{
		"SUBSCRIPTIONS":     subscriptionsFile.Name(),
		"LOG_LEVEL":         "warning",
		"LOG_LEVEL_CONSOLE": "info",
	})

	cfg, err := env.NewConfig()
	assert.Nil(t, err)
	assert.Equal(t, "info", cfg.LogLevelConsole())
	assert.Equal(t, "warning", cfg.LogLevelRemote())
}

func TestLogLevelRemote(t *testing.T) {
	clearEnv()
	defer clearEnv()

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	setEnv(map[string]string{
		"SUBSCRIPTIONS":  subscriptionsFile.Name(),
		"LOG_LEVEL":      "warning",
		"LOG_LEVEL_MQTT": "info",
	})

	cfg, err := env.NewConfig()
	assert.Nil(t, err)
	assert.Equal(t, "info", cfg.LogLevelRemote())
	assert.Equal(t, "warning", cfg.LogLevelConsole())
}

func TestLogLevelConsoleRemote(t *testing.T) {
	clearEnv()
	defer clearEnv()

	subscriptionsFile, _ := ioutil.TempFile("", "")
	defer os.Remove(subscriptionsFile.Name())

	setEnv(map[string]string{
		"SUBSCRIPTIONS":     subscriptionsFile.Name(),
		"LOG_LEVEL_CONSOLE": "debug",
		"LOG_LEVEL_MQTT":    "info",
	})

	cfg, err := env.NewConfig()
	assert.Nil(t, err)
	assert.Equal(t, "debug", cfg.LogLevelConsole())
	assert.Equal(t, "info", cfg.LogLevelRemote())
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
