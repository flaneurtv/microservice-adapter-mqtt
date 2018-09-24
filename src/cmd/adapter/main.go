package main

import (
	"fmt"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/env"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/logger"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/mqtt"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core/process"
)

func main() {
	log := logger.NewLogger()

	cfg, err := env.NewConfig()
	if err != nil {
		log.Panic("can't create config", err)
	}

	listenerClientID := fmt.Sprintf("%s_%s_%s_listener", cfg.ServiceName(), cfg.ServiceHost(), cfg.ServiceUUID())
	listener := mqtt.NewMQTTClient(cfg.ListenerURL(), listenerClientID, cfg.ListenerCredentials())

	var publisher core.MessageBusClient
	if cfg.ListenerURL() != cfg.PublisherURL() || cfg.ListenerCredentials() != cfg.PublisherCredentials() {
		publisherClientID := fmt.Sprintf("%s_%s_%s_publisher", cfg.ServiceName(), cfg.ServiceHost(), cfg.ServiceUUID())
		publisher = mqtt.NewMQTTClient(cfg.PublisherURL(), publisherClientID, cfg.PublisherCredentials())
	} else {
		publisher = listener
	}

	service := process.NewService(cfg.ServiceName(), cfg.ServiceUUID(), cfg.ServiceHost(), cfg.Namespace(), cfg.ServiceCmdLine(), log)

	adapter := core.NewAdapter(listener, publisher, cfg.Subscriptions(), service, log)
	done, err := adapter.Start()
	if err != nil {
		log.Panic("can't start adapter", err)
	}

	<-done
}
