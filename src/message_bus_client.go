package core

type MessageBusClient interface {
	Connect() error
	Subscribe(topics []string) (<-chan string, error)
	Publish(topic, message string) error
}
