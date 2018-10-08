package core

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
)

type Adapter struct {
	listener      MessageBusClient
	publisher     MessageBusClient
	subscriptions []string
	service       Service
	logger        Logger
}

func NewAdapter(listener, publisher MessageBusClient, subscriptions []string, service Service, logger Logger) *Adapter {
	return &Adapter{
		listener:      listener,
		publisher:     publisher,
		subscriptions: subscriptions,
		service:       service,
		logger:        logger,
	}
}

func (a *Adapter) Start() (<-chan struct{}, error) {
	err := a.listener.Connect()
	if err != nil {
		return nil, fmt.Errorf("can't connect: %s", err)
	}

	if a.publisher != a.listener {
		err := a.publisher.Connect()
		if err != nil {
			return nil, fmt.Errorf("can't connect: %s", err)
		}
	} else {
		a.logger.Debug("MQTT connection: listener and publisher are equal")
	}

	inputMessages, err := a.listener.Subscribe(a.subscriptions)
	if err != nil {
		return nil, fmt.Errorf("can't subscribe: %s", err)
	} else {
		a.logger.Info(fmt.Sprintf("Topics subscribed: %s", strings.Join(a.subscriptions, ", ")))
	}

	outputMessages, err := a.service.Start(inputMessages)
	if err != nil {
		return nil, fmt.Errorf("can't start a service: %s", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		for msg := range outputMessages {
			topic := gjson.Get(msg, "topic").String()
			err := a.publisher.Publish(topic, msg)
			if err != nil {
				a.logger.Error("can't publish", err)
			} else {
				a.logger.Debug(fmt.Sprintf("published: %s", msg))
			}
		}
	}()

	return done, nil
}
