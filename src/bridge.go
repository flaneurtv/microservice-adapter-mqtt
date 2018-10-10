package core

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"strings"
)

type Bridge struct {
	listener           MessageBusClient
	publisher          MessageBusClient
	namespaceListener  string
	namespacePublisher string
	subscriptions      []string
	logger             Logger
}

func NewBridge(listener, publisher MessageBusClient, namespaceListener, namespacePublisher string, subscriptions []string, logger Logger) *Bridge {
	return &Bridge{
		listener:           listener,
		publisher:          publisher,
		namespaceListener:  namespaceListener,
		namespacePublisher: namespacePublisher,
		subscriptions:      subscriptions,
		logger:             logger,
	}
}

func (b *Bridge) Start() (<-chan struct{}, error) {
	err := b.listener.Connect()
	if err != nil {
		return nil, fmt.Errorf("can't connect: %s", err)
	}

	if b.publisher != b.listener {
		err := b.publisher.Connect()
		if err != nil {
			return nil, fmt.Errorf("can't connect: %s", err)
		}
	} else {
		b.logger.Log(LogLevelDebug, "MQTT connection: listener and publisher are equal")
	}

	inputMessages, err := b.listener.Subscribe(b.subscriptions)
	if err != nil {
		return nil, fmt.Errorf("can't subscribe: %s", err)
	} else {
		b.logger.Log(LogLevelInfo, fmt.Sprintf("Topics subscribed: %s", strings.Join(b.subscriptions, ", ")))
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		for inpMsg := range inputMessages {
			if gjson.Valid(inpMsg) {
				inpTopic := gjson.Get(inpMsg, "topic").String()
				if inpTopic != "" {
					topic, msg := inpTopic, inpMsg
					if b.namespacePublisher != b.namespaceListener {
						topic = strings.Replace(inpTopic, b.namespaceListener+"/", b.namespacePublisher+"/", 1)
						msg, _ = sjson.Set(msg, "topic", topic)
					}

					err := b.publisher.Publish(topic, msg)
					if err != nil {
						b.logger.Log(LogLevelError, fmt.Sprintf("MQTT message for bridge dropped: %s, error=%s", msg, err))
					} else {
						b.logger.Log(LogLevelDebug, fmt.Sprintf("MQTT message relayed through bridge: %s => %s", inpTopic, topic))
					}
				} else {
					b.logger.Log(LogLevelError, fmt.Sprintf("missing topic: %s", inpMsg))
				}
			} else {
				b.logger.Log(LogLevelError, fmt.Sprintf("invalid json: %s", inpMsg))
			}
		}
	}()

	return done, nil
}
