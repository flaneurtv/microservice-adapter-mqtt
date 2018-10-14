package core_test

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/flaneurtv/microservice-adapter-mqtt/core"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	level, ok := core.ParseLogLevel("error")

	assert.True(t, ok)
	assert.Equal(t, core.LogLevelError, level)

	level, ok = core.ParseLogLevel("err")

	assert.False(t, ok)
	assert.Equal(t, core.LogLevelDebug, level)

	level, ok = core.ParseLogLevel("DEBUG")

	assert.True(t, ok)
	assert.Equal(t, core.LogLevelDebug, level)

	level, ok = core.ParseLogLevel("inFO")

	assert.True(t, ok)
	assert.Equal(t, core.LogLevelInfo, level)
}

func TestLogLevelWeakness(t *testing.T) {
	assert.False(t, core.LogLevelError.IsWeaker(core.LogLevelError))
	assert.True(t, core.LogLevelError.IsWeaker(core.LogLevelCritical))
	assert.True(t, core.LogLevelDebug.IsWeaker(core.LogLevelError))
	assert.False(t, core.LogLevelInfo.IsWeaker(core.LogLevelDebug))

	assert.True(t, core.LogLevelWarning.IsWeaker(core.LogLevelEmergency))
	assert.True(t, core.LogLevelWarning.IsWeaker(core.LogLevelAlert))
	assert.True(t, core.LogLevelWarning.IsWeaker(core.LogLevelCritical))
	assert.True(t, core.LogLevelWarning.IsWeaker(core.LogLevelError))
	assert.False(t, core.LogLevelWarning.IsWeaker(core.LogLevelWarning))
	assert.False(t, core.LogLevelWarning.IsWeaker(core.LogLevelNotice))
	assert.False(t, core.LogLevelWarning.IsWeaker(core.LogLevelInfo))
	assert.False(t, core.LogLevelWarning.IsWeaker(core.LogLevelDebug))
}
