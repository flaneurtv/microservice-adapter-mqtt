package env

import (
	"encoding/json"
	"errors"
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
	defaultListenerCredentialsPath  = "/run/secrets/mqtt_listener.json"
	defaultPublisherCredentialsPath = "/run/secrets/mqtt_publisher.json"
	defaultListenerURL              = "tcp://mqtt:1883"
	defaultPublisherURL             = "tcp://mqtt:1883"
	defaultServiceCmdLine           = "/srv/processor"
	defaultSubscriptionsFile        = "/srv/subscriptions.txt"
)

type config struct {
	serviceName    string
	serviceUUID    string
	serviceHost    string
	serviceCmdLine string

	namespaceListener  string
	namespacePublisher string

	listenerURL          string
	listenerCredentials  core.Credentials
	publisherURL         string
	publisherCredentials core.Credentials

	subscriptions []string

	logLevelConsole string
	logLevelRemote  string
}

func NewAdapterConfig(logger core.Logger) (core.Configuration, error) {
	return newConfig(logger, true)
}

func NewBridgeConfig(logger core.Logger) (core.Configuration, error) {
	return newConfig(logger, false)
}

func newConfig(logger core.Logger, withServiceProcessor bool) (core.Configuration, error) {
	serviceName := os.Getenv("SERVICE_NAME")
	serviceUUID := uuid.NewV4().String()
	serviceHost, _ := os.Hostname()

	var serviceCmdLine string
	if withServiceProcessor {
		var err error
		serviceCmdLine, err = getServiceCmdLine(logger)
		if err != nil {
			return nil, err
		}
	}

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
	if listenerURL == "" {
		logger.Log(core.LogLevelInfo, fmt.Sprintf("MQTT_LISTENER_URL not set, trying default url '%s'", defaultListenerURL))
		listenerURL = defaultListenerURL
	}

	publisherURL := os.Getenv("MQTT_PUBLISHER_URL")
	if publisherURL == "" {
		logger.Log(core.LogLevelInfo, fmt.Sprintf("MQTT_PUBLISHER_URL not set, trying default url '%s'", defaultPublisherURL))
		publisherURL = defaultPublisherURL
	}

	listenerCredentials, err := readCredentials("Listener", "MQTT_LISTENER_CREDENTIALS", defaultListenerCredentialsPath, logger)
	if err != nil {
		return nil, err
	}

	publisherCredentials, err := readCredentials("Publisher", "MQTT_PUBLISHER_CREDENTIALS", defaultPublisherCredentialsPath, logger)
	if err != nil {
		return nil, err
	}

	subscriptions, err := readSubscriptions(namespaceListener, logger)
	if err != nil {
		return nil, err
	}

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "error"
	}

	logLevelConsole := strings.ToLower(os.Getenv("LOG_LEVEL_CONSOLE"))
	if logLevelConsole == "" {
		logLevelConsole = logLevel
	}

	logLevelRemote := strings.ToLower(os.Getenv("LOG_LEVEL_MQTT"))
	if logLevelRemote == "" {
		logLevelRemote = logLevel
	}

	return &config{
		serviceName:          serviceName,
		serviceUUID:          serviceUUID,
		serviceHost:          serviceHost,
		serviceCmdLine:       serviceCmdLine,
		namespaceListener:    namespaceListener,
		namespacePublisher:   namespacePublisher,
		listenerURL:          listenerURL,
		listenerCredentials:  listenerCredentials,
		publisherURL:         publisherURL,
		publisherCredentials: publisherCredentials,
		subscriptions:        subscriptions,
		logLevelConsole:      logLevelConsole,
		logLevelRemote:       logLevelRemote,
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

func (cfg *config) LogLevelConsole() string {
	return cfg.logLevelConsole
}

func (cfg *config) LogLevelRemote() string {
	return cfg.logLevelRemote
}

func readSubscriptions(namespace string, logger core.Logger) ([]string, error) {
	subscriptionsPath, ok := os.LookupEnv("SUBSCRIPTIONS")
	if !ok {
		logger.Log(core.LogLevelWarning, fmt.Sprintf("SUBSCRIPTIONS not set, trying default location '%s'", defaultSubscriptionsFile))
		subscriptionsPath = defaultSubscriptionsFile
	} else if strings.TrimSpace(subscriptionsPath) == "" {
		logger.Log(core.LogLevelInfo, "SUBSCRIPTIONS set to nil, starting without subscriptions")
		return nil, nil
	}

	content, err := ioutil.ReadFile(subscriptionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Log(core.LogLevelWarning, fmt.Sprintf("Subscriptions file not found at '%s', starting without subscriptions: %s", subscriptionsPath, err))
			return nil, nil
		}
		return nil, fmt.Errorf("can't read subscriptions: %s", err)
	}

	logger.Log(core.LogLevelInfo, fmt.Sprintf("Subscriptions file found at '%s'", subscriptionsPath))

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

	if len(subscriptions) == 0 {
		logger.Log(core.LogLevelWarning, "Subscriptions file empty, starting without subscriptions")
	}

	return subscriptions, nil
}

func readCredentials(credentialsTitle, credentialsEnvVar, defaultCredentialsPath string, logger core.Logger) (core.Credentials, error) {
	var credentials core.Credentials

	credentialsPath, ok := os.LookupEnv(credentialsEnvVar)
	if !ok {
		logger.Log(core.LogLevelWarning, fmt.Sprintf("%s not set, trying default location %s", credentialsEnvVar, defaultCredentialsPath))
		credentialsPath = defaultCredentialsPath
	} else if strings.TrimSpace(credentialsPath) == "" {
		return credentials, fmt.Errorf("%s can't be empty", credentialsEnvVar)
	}

	content, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Log(core.LogLevelWarning, fmt.Sprintf("%s credentials file '%s' doesn't exist - trying to connect with empty credentials", credentialsTitle, credentialsPath))
			return credentials, nil
		}
		return credentials, fmt.Errorf("can't read credentials: %s", err)
	}

	logger.Log(core.LogLevelInfo, fmt.Sprintf("%s credentials found at '%s'", credentialsTitle, credentialsPath))

	err = json.Unmarshal(content, &credentials)
	if err != nil {
		return credentials, fmt.Errorf("can't parse credentials: %s", err)
	}

	return credentials, nil
}

func getServiceCmdLine(logger core.Logger) (string, error) {
	serviceCmdLine, ok := os.LookupEnv("SERVICE_PROCESSOR")
	if !ok {
		logger.Log(core.LogLevelWarning, fmt.Sprintf("SERVICE_PROCESSOR not set, trying %s", defaultServiceCmdLine))
		serviceCmdLine = defaultServiceCmdLine
	} else if strings.TrimSpace(serviceCmdLine) == "" {
		return "", errors.New("SERVICE_PROCESSOR can't be empty")
	}

	parts := strings.Fields(serviceCmdLine)
	info, err := os.Stat(parts[0])
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", errors.New("SERVICE_PROCESSOR should reference to a file")
	}
	return serviceCmdLine, nil
}
