package core

type Configuration interface {
	ServiceName() string
	ServiceUUID() string
	ServiceHost() string
	ServiceCmdLine() string

	NamespaceListener() string
	NamespacePublisher() string

	ListenerURL() string
	ListenerCredentials() Credentials
	PublisherURL() string
	PublisherCredentials() Credentials

	Subscriptions() []string

	LogLevel() string
}

type Credentials struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}