package mqtt

import (
	"github.com/eclipse/paho.mqtt.golang"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
)

type mqttClient struct {
	client mqtt.Client
}

func NewMQTTClient(busURL, clientID string, credentials core.Credentials) core.MessageBusClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(busURL)
	opts.SetClientID(clientID)
	opts.Username = credentials.UserName
	opts.Password = credentials.Password

	client := mqtt.NewClient(opts)

	return &mqttClient{
		client: client,
	}
}

func (m *mqttClient) Connect() error {
	token := m.client.Connect()
	token.Wait()
	return token.Error()
}

func (m *mqttClient) Publish(topic, message string) error {
	token := m.client.Publish(topic, 0, false, message)
	token.Wait()
	return token.Error()
}

func (m *mqttClient) Subscribe(topics []string) (<-chan string, error) {
	messages := make(chan string)
	topicsMap := make(map[string]byte, len(topics))
	for _, topic := range topics {
		topicsMap[topic] = 0
	}
	token := m.client.SubscribeMultiple(topicsMap, func(cl mqtt.Client, msg mqtt.Message) {
		messages <- string(msg.Payload())
	})
	token.Wait()
	return messages, token.Error()
}
