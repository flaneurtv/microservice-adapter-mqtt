package mqtt

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"gitlab.com/flaneurtv/samm/core"
)

type mqttClient struct {
	client           mqtt.Client
	subscribedTopics [][]string
	inputMessages    []chan<- string
}

func NewMQTTClient(busURL, clientID string, credentials core.Credentials, logger core.Logger, onConnectionLost func(err error)) core.MessageBusClient {
	var client *mqttClient

	opts := mqtt.NewClientOptions()
	opts.AddBroker(busURL)
	opts.SetClientID(clientID)
	opts.Username = credentials.UserName
	opts.Password = credentials.Password
	opts.OnConnect = func(cl mqtt.Client) {
		logger.Log(core.LogLevelInfo, fmt.Sprintf("MQTT client connected to %s", busURL))

		err := client.subscribe()
		if err != nil {
			logger.Log(core.LogLevelError, fmt.Sprintf("Can't re-subscribe: %s", err))
		}
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		logger.Log(core.LogLevelInfo, fmt.Sprintf("MQTT client lost connection to %s: %s", busURL, err))
		if onConnectionLost != nil {
			onConnectionLost(err)
		}
	}

	internalClient := mqtt.NewClient(opts)

	client = &mqttClient{
		client: internalClient,
	}

	return client
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
	m.inputMessages = append(m.inputMessages, messages)
	m.subscribedTopics = append(m.subscribedTopics, topics)

	err := m.subscribe()
	return messages, err
}

func (m *mqttClient) subscribe() error {
	if len(m.subscribedTopics) == 0 {
		return nil
	}

	for i, topics := range m.subscribedTopics {
		topicsMap := make(map[string]byte, len(topics))
		for _, topic := range topics {
			topicsMap[topic] = 0
		}

		m.client.Unsubscribe(topics...)

		token := m.client.SubscribeMultiple(topicsMap, func(cl mqtt.Client, msg mqtt.Message) {
			m.inputMessages[i] <- string(msg.Payload())
		})
		token.Wait()
		err := token.Error()
		if err != nil {
			return err
		}
	}

	return nil
}
