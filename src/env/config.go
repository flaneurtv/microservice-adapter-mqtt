package env

import (
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"io/ioutil"
	"os"
	"strings"
)

const (
	defaultNamespace                = "default"
	nullNamespace                   = "null"
	defaultListenerCredentialsPath  = "path/to/secrets/mqtt_listener.json"
	defaultPublisherCredentialsPath = "path/to/secrets/mqtt_publisher.json"
)

type config struct {
	serviceName    string
	serviceUUID    string
	serviceHost    string
	serviceCmdLine string

	namespace          string
	namespaceListener  string
	namespacePublisher string

	listenerURL          string
	listenerCredentials  core.Credentials
	publisherURL         string
	publisherCredentials core.Credentials

	subscriptions []string
}

func NewConfig() (core.Configuration, error) {
	serviceName := os.Getenv("SERVICE_NAME")
	serviceUUID := uuid.NewV4().String()
	serviceHost, _ := os.Hostname()
	serviceCmdLine := os.Getenv("SERVICE_PROCESSOR")

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = defaultNamespace
	}

	namespaceListener := os.Getenv("NAMESPACE_LISTENER")
	if namespaceListener == "" {
		namespaceListener = namespace
	}

	namespacePublisher := os.Getenv("NAMESPACE_PUBLISHER")
	if namespacePublisher == "" {
		namespacePublisher = namespace
	}

	listenerURL := os.Getenv("MQTT_LISTENER_URL")
	publisherURL := os.Getenv("MQTT_PUBLISHER_URL")

	listenerCredentials, err := readCredentials(os.Getenv("MQTT_LISTENER_CREDENTIALS"), defaultListenerCredentialsPath)
	if err != nil {
		return nil, err
	}

	publisherCredentials, err := readCredentials(os.Getenv("MQTT_PUBLISHER_CREDENTIALS"), defaultPublisherCredentialsPath)
	if err != nil {
		return nil, err
	}

	subscriptionsPath := os.Getenv("SUBSCRIPTIONS")
	subscriptions, err := readSubscriptions(subscriptionsPath, namespaceListener)
	if err != nil {
		return nil, err
	}

	return &config{
		serviceName:          serviceName,
		serviceUUID:          serviceUUID,
		serviceHost:          serviceHost,
		serviceCmdLine:       serviceCmdLine,
		namespace:            namespace,
		namespaceListener:    namespaceListener,
		namespacePublisher:   namespacePublisher,
		listenerURL:          listenerURL,
		listenerCredentials:  listenerCredentials,
		publisherURL:         publisherURL,
		publisherCredentials: publisherCredentials,
		subscriptions:        subscriptions,
	}, nil
}

func (cfg *config) ServiceName() string {
	return cfg.serviceName
}

func (cfg *config) ServiceUUID() string {
	return cfg.serviceUUID
}

func (cfg *config) ServiceHost() string {
	return cfg.serviceHost
}

func (cfg *config) ServiceCmdLine() string {
	return cfg.serviceCmdLine
}

func (cfg *config) Namespace() string {
	return cfg.namespace
}

func (cfg *config) NamespaceListener() string {
	return cfg.namespaceListener
}

func (cfg *config) NamespacePublisher() string {
	return cfg.namespacePublisher
}

func (cfg *config) ListenerURL() string {
	return cfg.listenerURL
}

func (cfg *config) ListenerCredentials() core.Credentials {
	return cfg.listenerCredentials
}

func (cfg *config) PublisherURL() string {
	return cfg.publisherURL
}

func (cfg *config) PublisherCredentials() core.Credentials {
	return cfg.publisherCredentials
}

func (cfg *config) Subscriptions() []string {
	return cfg.subscriptions
}

func readSubscriptions(subscriptionsPath, namespace string) ([]string, error) {
	content, err := ioutil.ReadFile(subscriptionsPath)
	if err != nil {
		return nil, fmt.Errorf("can't read subscriptions: %s", err)
	}

	lines := strings.Split(string(content), "\n")
	subscriptions := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			topic := line
			if namespace != nullNamespace {
				topic = fmt.Sprintf("%s/%s", namespace, topic)
			}
			subscriptions = append(subscriptions, topic)
		}
	}

	return subscriptions, nil
}

func readCredentials(credentialsPath, defaultCredentialsPath string) (core.Credentials, error) {
	var credentials core.Credentials

	path := credentialsPath
	if path == "" {
		path = defaultCredentialsPath
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && credentialsPath == "" {
			return credentials, nil
		}
		return credentials, fmt.Errorf("can't read credentials: %s", err)
	}

	err = json.Unmarshal(content, &credentials)
	if err != nil {
		return credentials, fmt.Errorf("can't parse credentials: %s", err)
	}

	return credentials, nil
}
