// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	assert := assert.New(t)
	assert.False(IsLoggingConfigured())

	logFileName := "./log/test.log"

	InitializeLog(logFileName, 10, 2, 5, false)

	if _, err := os.Stat(logFileName); os.IsNotExist(err) {
		t.Errorf("Test Failed error: %v\n", err)
	}

	SetConsoleLog(true)
	Log("Started log")
	SetDeubgLog(true)
	LogDebug("Debug log should come up as well")
	Log("End log")

	assert.True(IsLoggingConfigured())
}
