// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package util

import (
	"os"
	"testing"
)

type AppCfg struct {
	AppName     string
	LogSettings LoggingCfg
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestUtilCfgLog(t *testing.T) {
	// TBD - delete the log file if it exist (manually done before running the test today).
	SetLogSettings("util_test.json") // ""
	if _, err := os.Stat(GlobalLogSettings.LogFileName); os.IsNotExist(err) {
		t.Errorf("Test Failed error: %v\n", err)
	}
}
