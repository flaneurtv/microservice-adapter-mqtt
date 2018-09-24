package core_test

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/logger"
	"strings"
	"testing"
)

func TestAdapter(t *testing.T) {
	bus := NewMockBus()
	client1 := NewMockClient(bus)
	var subscriptions1 []string
	listener2 := NewMockClient(bus)
	publisher2 := NewMockClient(bus)
	subscriptions2 := []string{"tick"}
	output1 := make(chan string)
	service1 := NewMockServiceProducer(output1)
	counter := 0
	service2 := NewMockService(func(msg string) string {
		counter++
		return fmt.Sprintf(`{"topic": "tick-response", "payload": %d}`, counter)
	})

	adapter1 := core.NewAdapter(client1, client1, subscriptions1, service1, logger.NewNoOpLogger())
	done1, err := adapter1.Start()
	assert.Nil(t, err)

	adapter2 := core.NewAdapter(listener2, publisher2, subscriptions2, service2, logger.NewNoOpLogger())
	done2, err := adapter2.Start()
	assert.Nil(t, err)

	output1 <- `{"topic": "tick", "payload": "a"}`
	output1 <- `{"topic": "tick", "payload": "b"}`
	output1 <- `{"topic": "test", "payload": "c"}`
	output1 <- `{"topic": "tick", "payload": "d"}`
	output1 <- `{"topic": "test", "payload": "e"}`
	output1 <- `{"topic": "tick", "payload": "stop"}`
	close(output1)

	<-done1
	<-done2

	assert.Equal(t, 3, len(service2.inputMessages))
	assert.Equal(t, `{"topic": "tick", "payload": "a"}`, service2.inputMessages[0])
	assert.Equal(t, `{"topic": "tick", "payload": "b"}`, service2.inputMessages[1])
	assert.Equal(t, `{"topic": "tick", "payload": "d"}`, service2.inputMessages[2])

	assert.Equal(t, 3, len(service2.outputMessages))
	assert.Equal(t, `{"topic": "tick-response", "payload": 1}`, service2.outputMessages[0])
	assert.Equal(t, `{"topic": "tick-response", "payload": 2}`, service2.outputMessages[1])
	assert.Equal(t, `{"topic": "tick-response", "payload": 3}`, service2.outputMessages[2])
}

func TestAdapterConnectError(t *testing.T) {
	bus := NewMockBus()
	client1 := NewMockClient(bus)
	client1.forceConnectError = true
	var subscriptions1 []string
	listener2 := NewMockClient(bus)
	publisher2 := NewMockClient(bus)
	publisher2.forceConnectError = true
	subscriptions2 := []string{"tick"}
	service1 := NewMockServiceProducer(make(chan string))
	service2 := NewMockService(func(msg string) string {
		return msg
	})

	adapter1 := core.NewAdapter(client1, client1, subscriptions1, service1, logger.NewNoOpLogger())
	_, err := adapter1.Start()
	assert.NotNil(t, err)
	assert.Equal(t, "can't connect: connect error", err.Error())

	adapter2 := core.NewAdapter(listener2, publisher2, subscriptions2, service2, logger.NewNoOpLogger())
	_, err = adapter2.Start()
	assert.NotNil(t, err)
	assert.Equal(t, "can't connect: connect error", err.Error())
}

func TestAdapterSubscribeError(t *testing.T) {
	bus := NewMockBus()
	client1 := NewMockClient(bus)
	client1.forceSubscribeError = true
	service1 := NewMockServiceProducer(make(chan string))

	adapter1 := core.NewAdapter(client1, client1, nil, service1, logger.NewNoOpLogger())
	_, err := adapter1.Start()
	assert.NotNil(t, err)
	assert.Equal(t, "can't subscribe: subscribe error", err.Error())
}

func TestAdapterStartError(t *testing.T) {
	bus := NewMockBus()
	client1 := NewMockClient(bus)
	service1 := NewMockServiceProducer(make(chan string))
	service1.forceStartError = true

	adapter1 := core.NewAdapter(client1, client1, nil, service1, logger.NewNoOpLogger())
	_, err := adapter1.Start()
	assert.NotNil(t, err)
	assert.Equal(t, "can't start a service: start error", err.Error())
}

type mockBus struct {
	subscribers map[string]chan<- string
}

func NewMockBus() *mockBus {
	return &mockBus{subscribers: make(map[string]chan<- string)}
}

func (b *mockBus) Subscribe(topics []string, messages chan<- string) {
	key := "|" + strings.Join(topics, "|") + "|"
	b.subscribers[key] = messages
}

func (b *mockBus) Publish(topic, message string) {
	key := "|" + topic + "|"
	for topics, messages := range b.subscribers {
		if strings.Contains(topics, key) {
			messages <- message
		}
	}
}

func (b *mockBus) close() {
	for _, messages := range b.subscribers {
		close(messages)
	}
}

type mockClient struct {
	bus                 *mockBus
	forceConnectError   bool
	forceSubscribeError bool
}

func NewMockClient(bus *mockBus) *mockClient {
	return &mockClient{bus: bus}
}

func (c *mockClient) Connect() error {
	if c.forceConnectError {
		return errors.New("connect error")
	}

	return nil
}

func (c *mockClient) Subscribe(topics []string) (<-chan string, error) {
	if c.forceSubscribeError {
		return nil, errors.New("subscribe error")
	}

	messages := make(chan string)
	c.bus.Subscribe(topics, messages)
	return messages, nil
}

func (c *mockClient) Publish(topic, message string) error {
	c.bus.Publish(topic, message)
	return nil
}

type mockService struct {
	getOutputMessage func(msg string) string
	inputMessages    []string
	outputMessages   []string
}

func NewMockService(getOutputMessage func(msg string) string) *mockService {
	return &mockService{getOutputMessage: getOutputMessage}
}

func (sp *mockService) Start(input <-chan string) (output <-chan string, err error) {
	out := make(chan string)
	go func() {
		defer close(out)
		for msg := range input {
			payload := gjson.Get(msg, "payload").String()
			if payload == "stop" {
				break
			}
			outMsg := sp.getOutputMessage(msg)
			out <- outMsg

			sp.inputMessages = append(sp.inputMessages, msg)
			sp.outputMessages = append(sp.outputMessages, outMsg)
		}
	}()
	return out, nil
}

type mockServiceProducer struct {
	output          <-chan string
	forceStartError bool
}

func NewMockServiceProducer(output <-chan string) *mockServiceProducer {
	return &mockServiceProducer{output: output}
}

func (sp *mockServiceProducer) Start(input <-chan string) (output <-chan string, err error) {
	if sp.forceStartError {
		return nil, errors.New("start error")
	}

	return sp.output, nil
}
