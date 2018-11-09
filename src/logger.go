package core

import (
	"strings"
	"time"
)

type LogLevel string

const (
	LogLevelDebug     LogLevel = "debug"
	LogLevelInfo      LogLevel = "info"
	LogLevelNotice    LogLevel = "notice"
	LogLevelWarning   LogLevel = "warning"
	LogLevelError     LogLevel = "error"
	LogLevelCritical  LogLevel = "critical"
	LogLevelAlert     LogLevel = "alert"
	LogLevelEmergency LogLevel = "emergency"
)

var levels = []LogLevel{
	LogLevelDebug, LogLevelInfo, LogLevelNotice, LogLevelWarning,
	LogLevelError, LogLevelCritical, LogLevelAlert, LogLevelEmergency,
}

type Logger interface {
	SetLevels(levelConsole, leverRemote LogLevel)
	SetClient(client MessageBusClient, namespace, serviceName, serviceUUID, serviceHost string)
	SetCreatedAtGetter(getCreatedAt func() time.Time)
	Log(level LogLevel, message string)
}

func (level LogLevel) IsWeaker(other LogLevel) bool {
	if other == "" {
		return false
	}

	for _, lvl := range levels {
		if lvl == other {
			return false
		}
		if lvl == level {
			return true
		}
	}
	return false
}

func ParseLogLevel(level string) (LogLevel, bool) {
	level = strings.ToLower(level)
	for _, lvl := range levels {
		if level == string(lvl) {
			return lvl, true
		}
	}
	return LogLevelDebug, false
}
