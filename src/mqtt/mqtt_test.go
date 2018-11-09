package mqtt_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/surgemq/surgemq/auth"
	"github.com/surgemq/surgemq/service"
	"github.com/surgemq/surgemq/topics"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/logger"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/mqtt"
	"sync"
	"testing"
	"time"
)

func TestClients(t *testing.T) {
	mqttURL := "tcp://:15355"
	srv := startMockMQTTServer(t, mqttURL, "")
	defer closeMockMQTTServer(t, srv)

	client1 := mqtt.NewMQTTClient(mqttURL, "client1", core.Credentials{}, logger.NewNoOpLogger(), nil)
	err := client1.Connect()
	assert.Nil(t, err)

	client2 := mqtt.NewMQTTClient(mqttURL, "client2", core.Credentials{}, logger.NewNoOpLogger(), nil)
	err = client2.Connect()
	assert.Nil(t, err)

	client3 := mqtt.NewMQTTClient(mqttURL, "client3", core.Credentials{}, logger.NewNoOpLogger(), nil)
	err = client3.Connect()
	assert.Nil(t, err)

	messages2, err := client2.Subscribe([]string{"test", "work"})
	assert.Nil(t, err)

	messages3, err := client3.Subscribe([]string{"work", "job"})
	assert.Nil(t, err)

	go func() {
		client1.Publish("test", "123")
		client1.Publish("job", "456")
		client1.Publish("work", "789")
		client1.Publish("job", "012")
	}()

	msg21 := <-messages2
	msg22 := <-messages2

	msg31 := <-messages3
	msg32 := <-messages3
	msg33 := <-messages3

	assert.Equal(t, "123", msg21)
	assert.Equal(t, "789", msg22)

	assert.Equal(t, "456", msg31)
	assert.Equal(t, "789", msg32)
	assert.Equal(t, "012", msg33)
}

func TestCredentials(t *testing.T) {
	auth.Register("test_auth", &testAuthenticator{})
	defer auth.Unregister("test_auth")

	mqttURL := "tcp://:15355"
	srv := startMockMQTTServer(t, mqttURL, "test_auth")
	defer closeMockMQTTServer(t, srv)

	client := mqtt.NewMQTTClient(mqttURL, "client1", core.Credentials{}, logger.NewNoOpLogger(), nil)
	err := client.Connect()
	assert.NotNil(t, err)

	client = mqtt.NewMQTTClient(mqttURL, "client1", core.Credentials{UserName: "user123", Password: "password123"}, logger.NewNoOpLogger(), nil)
	err = client.Connect()
	assert.Nil(t, err)

	client = mqtt.NewMQTTClient(mqttURL, "client1", core.Credentials{UserName: "user555", Password: "password555"}, logger.NewNoOpLogger(), nil)
	err = client.Connect()
	assert.NotNil(t, err)
}

func TestLostConnection(t *testing.T) {
	mqttURL := "tcp://:15355"
	srv := startMockMQTTServer(t, mqttURL, "")

	var lost bool

	client1 := mqtt.NewMQTTClient(mqttURL, "client1", core.Credentials{}, logger.NewNoOpLogger(), func(err error) {
		lost = true
	})
	err := client1.Connect()
	assert.Nil(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		client1.Publish("test", "123")
		client1.Publish("job", "456")

		closeMockMQTTServer(t, srv)

		client1.Publish("work", "789")
		client1.Publish("job", "012")
	}()

	wg.Wait()
	time.Sleep(time.Millisecond * 1000)
	assert.True(t, lost)
}

func TestReconnect(t *testing.T) {
	mqttURL := "tcp://:15355"
	srv := startMockMQTTServer(t, mqttURL, "")
	defer func() {
		defer closeMockMQTTServer(t, srv)
	}()

	client1 := mqtt.NewMQTTClient(mqttURL, "client1", core.Credentials{}, logger.NewNoOpLogger(), nil)
	err := client1.Connect()
	assert.Nil(t, err)

	client2 := mqtt.NewMQTTClient(mqttURL, "client2", core.Credentials{}, logger.NewNoOpLogger(), nil)
	err = client2.Connect()
	assert.Nil(t, err)

	client3 := mqtt.NewMQTTClient(mqttURL, "client3", core.Credentials{}, logger.NewNoOpLogger(), nil)
	err = client3.Connect()
	assert.Nil(t, err)

	messages2, err := client2.Subscribe([]string{"test", "work"})
	assert.Nil(t, err)

	messages3, err := client3.Subscribe([]string{"work", "job"})
	assert.Nil(t, err)

	go func() {
		client1.Publish("test", "123")
		client1.Publish("job", "456")

		time.Sleep(time.Millisecond * 500)
		closeMockMQTTServer(t, srv)
		srv = startMockMQTTServer(t, mqttURL, "")
		time.Sleep(time.Millisecond * 1000)

		err = client1.Publish("work", "789")
		assert.Nil(t, err)
		err = client1.Publish("job", "012")
		assert.Nil(t, err)
		err = client1.Publish("work", "777")
		assert.Nil(t, err)
	}()

	msg21 := <-messages2
	msg22 := <-messages2
	msg23 := <-messages2

	msg31 := <-messages3
	msg32 := <-messages3
	msg33 := <-messages3
	msg34 := <-messages3

	assert.Equal(t, "123", msg21)
	assert.Equal(t, "789", msg22)
	assert.Equal(t, "777", msg23)

	assert.Equal(t, "456", msg31)
	assert.Equal(t, "789", msg32)
	assert.Equal(t, "012", msg33)
	assert.Equal(t, "777", msg34)
}

func startMockMQTTServer(t *testing.T, mqttURL, authenticator string) *service.Server {
	time.Sleep(500 * time.Millisecond)
	srv := &service.Server{}
	srv.Authenticator = authenticator
	go func() {
		err := srv.ListenAndServe(mqttURL)
		assert.Nil(t, err)
	}()
	time.Sleep(100 * time.Millisecond)
	return srv
}

func closeMockMQTTServer(t *testing.T, srv *service.Server) {
	err := srv.Close()
	assert.Nil(t, err)

	topics.Unregister("mem")
	topics.Register("mem", topics.NewMemProvider())
}

type testAuthenticator struct {
}

func (*testAuthenticator) Authenticate(id string, cred interface{}) error {
	if id == "user123" && cred.(string) == "password123" {
		return nil
	}
	return errors.New("bad auth")
}
