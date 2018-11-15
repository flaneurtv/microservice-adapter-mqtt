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
		a.logger.Log(LogLevelDebug, "MQTT connection: listener and publisher are equal")
	}

	var inputMessages <-chan string
	if len(a.subscriptions) > 0 {
		inputMessages, err = a.listener.Subscribe(a.subscriptions)
		if err != nil {
			return nil, fmt.Errorf("can't subscribe: %s", err)
		} else {
			a.logger.Log(LogLevelInfo, fmt.Sprintf("Topics subscribed: %s", strings.Join(a.subscriptions, ", ")))
		}
	}

	outputMessages, errorMessages, err := a.service.Start(inputMessages)
	if err != nil {
		return nil, fmt.Errorf("can't start a service: %s", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

	LOOP:
		for {
			select {
			case msg, ok := <-outputMessages:
				if !ok {
					break LOOP
				}

				if gjson.Valid(msg) {
					topic := gjson.Get(msg, "topic").String()
					if topic != "" {
						err := a.publisher.Publish(topic, msg)
						if err != nil {
							a.logger.Log(LogLevelError, fmt.Sprintf("can't publish: %s", err))
						} else {
							a.logger.Log(LogLevelDebug, fmt.Sprintf("published: %s", msg))
						}
					} else {
						a.logger.Log(LogLevelError, fmt.Sprintf("missing topic: %s", msg))
					}
				} else {
					a.logger.Log(LogLevelError, fmt.Sprintf("invalid json: %s", msg))
				}
			case msg, ok := <-errorMessages:
				if !ok {
					break LOOP
				}

				logLevel := LogLevelError
				message := msg
				if gjson.Valid(msg) {
					message = gjson.Get(msg, "log_message").String()
					if message == "" {
						message = msg
					} else {
						logLevel, _ = ParseLogLevel(gjson.Get(msg, "log_level").String())
					}
				}
				a.logger.Log(logLevel, message)
			}
		}
	}()

	return done, nil
}
