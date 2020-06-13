// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const logFileName = "./log/test.log"

func TestLog(t *testing.T) {
	assert := assert.New(t)
	assert.False(IsLoggingConfigured())

	initilizeTestLog()

	if _, err := os.Stat(logFileName); os.IsNotExist(err) {
		t.Errorf("Test Failed error: %v\n", err)
	}

	Log("Started log")
	LogDebug("Debug log should come up as well")
	Log("End log")

	assert.True(IsLoggingConfigured())
}

func initilizeTestLog() {
	InitializeLog(logFileName, 10, 2, 5, false)
	SetConsoleLog(true)
	SetDebugLog(true)
}
