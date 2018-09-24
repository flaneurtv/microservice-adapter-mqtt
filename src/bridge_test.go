package core_test

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/logger"
	"sync"
	"testing"
	"time"
)

func TestBridge(t *testing.T) {
	bus := NewMockBus()
	client1 := NewMockClient(bus)
	client2 := NewMockClient(bus)

	listener := NewMockClient(bus)
	publisher := NewMockClient(bus)

	bridge := core.NewBridge(listener, publisher, "tick", "tack", []string{"tick/first"}, logger.NewNoOpLogger())
	done, err := bridge.Start()
	assert.Nil(t, err)

	output, err := client2.Subscribe([]string{"tack/first"})
	assert.Nil(t, err)

	var messages []string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for msg := range output {
			messages = append(messages, msg)
		}
		wg.Done()
	}()

	client1.Publish("tick/first", `{"topic": "tick/first", "payload": "a"}`)
	client1.Publish("tick/first", `{"topic": "tick/first", "payload": "b"}`)
	client1.Publish("test/first", `{"topic": "tick/first", "payload": "c"}`)
	client1.Publish("tick/first", `{"topic": "tick/first", "payload": "d"}`)
	client1.Publish("test/first", `{"topic": "tick/first", "payload": "e"}`)
	client1.Publish("test/first", `{"topic": "tick/first", "payload": "stop"}`)

	time.Sleep(time.Millisecond * 500)
	bus.close()
	<-done
	wg.Wait()

	assert.Equal(t, 3, len(messages))
	assert.Equal(t, `{"topic": "tack/first", "payload": "a"}`, messages[0])
	assert.Equal(t, `{"topic": "tack/first", "payload": "b"}`, messages[1])
	assert.Equal(t, `{"topic": "tack/first", "payload": "d"}`, messages[2])
}

func TestBridgeConnectError(t *testing.T) {
	bus := NewMockBus()
	client := NewMockClient(bus)
	client.forceConnectError = true
	bridge := core.NewBridge(client, client, "tick", "tack", []string{"tick/first"}, logger.NewNoOpLogger())
	_, err := bridge.Start()
	assert.NotNil(t, err)
	assert.Equal(t, "can't connect: connect error", err.Error())
}

func TestBridgeConnectPublisherError(t *testing.T) {
	bus := NewMockBus()
	listener := NewMockClient(bus)
	publisher := NewMockClient(bus)
	publisher.forceConnectError = true
	bridge := core.NewBridge(listener, publisher, "tick", "tack", []string{"tick/first"}, logger.NewNoOpLogger())
	_, err := bridge.Start()
	assert.NotNil(t, err)
	assert.Equal(t, "can't connect: connect error", err.Error())
}

func TestBridgeSubscribeError(t *testing.T) {
	bus := NewMockBus()
	client := NewMockClient(bus)
	client.forceSubscribeError = true
	bridge := core.NewBridge(client, client, "tick", "tack", []string{"tick/first"}, logger.NewNoOpLogger())
	_, err := bridge.Start()
	assert.NotNil(t, err)
	assert.Equal(t, "can't subscribe: subscribe error", err.Error())
}
