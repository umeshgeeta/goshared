// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/umeshgeeta/goshared/util"
	"testing"
)

func TestExecutorServiceFromDefaultCfg(t *testing.T) {
	assert := assert.New(t)

	es := NewExecutionService("xxx", true)
	util.Log(fmt.Sprintf("Config: %v\n", GlobalExecServiceCfg))

	dsp := es.taskDispatcher
	assert.NotNil(dsp)

	ep := dsp.execPool
	assert.NotNil(ep)

	assert.NotNil(ep.asyncExecutors)
	assert.NotZero(len(ep.asyncExecutors))

	assert.NotNil(ep.blockingExecutors)
	assert.NotZero(len(ep.blockingExecutors))
}
