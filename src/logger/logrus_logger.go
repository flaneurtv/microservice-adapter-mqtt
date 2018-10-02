package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
)

type logrusLogger struct {
	log *logrus.Logger
}

func NewLogger() core.Logger {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.99", FullTimestamp: true}

	return &logrusLogger{
		log: log,
	}
}

func (logger *logrusLogger) SetLevel(level string) {
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		parsedLevel = logrus.ErrorLevel
	}
	logger.log.SetLevel(parsedLevel)
}

func (logger *logrusLogger) Debug(message string) {
	logger.log.Debugln(message)
}

func (logger *logrusLogger) Info(message string) {
	logger.log.Infoln(message)
}

func (logger *logrusLogger) Warn(message string) {
	logger.log.Warnln(message)
}

func (logger *logrusLogger) Error(message string, err error) {
	logger.log.Errorln(fmt.Sprintf("%s: %s", message, err))
}

func (logger *logrusLogger) Panic(message string, err error) {
	logger.log.Panic(fmt.Sprintf("%s: %s", message, err))

}
