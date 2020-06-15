// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/umeshgeeta/goshared/util"
	"os"
	"testing"
	"time"
)

var es *ExecutionService

func TestMain(m *testing.M) {
	SetupRand()
	SetupTestTask()

	// We use default configuration, including for logging
	es = NewExecutionService("xxx", true)
	// this is expected to initialize logging as well
	util.SetConsoleLog(true)
	util.Log(fmt.Sprintf("Config: %v\n", es.
		ServiceCfgInUse))

	// start the execution service
	es.Start()

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestExecutorServiceFromDefaultCfg(t *testing.T) {
	//t.SkipNow()
	assert := assert.New(t)

	dsp := es.taskDispatcher
	assert.NotNil(dsp)

	ep := dsp.execPool
	assert.NotNil(ep)

	assert.NotNil(ep.asyncExecutors)
	acount := len(ep.asyncExecutors)
	assert.NotZero(acount)
	assert.Equal(acount, es.ServiceCfgInUse.ExexPool.AsyncTaskExecutorCount)

	assert.NotNil(ep.blockingExecutors)
	bcount := len(ep.blockingExecutors)
	assert.NotZero(bcount)
	assert.Equal(bcount, es.ServiceCfgInUse.ExexPool.BlockingTaskExecutorCount)

	task1 := NewBlockingTestTask(30, true)
	es.Submit(task1)
	// Note - if you make the task time small, say 15 i.e. 15 microseconds,
	// response of the task from executor comes faster than go routines
	// - one for housekeeping and another the caller for blocking tasks -
	// obtain Locks on mutex encapsulated by the 'waitingTask' in dispatcher.
	// The mutex is used by the conditional variable.
	task2 := NewBlockingTestTask(1000, false)
	es.Submit(task2)
	//task3 := NewTestTask(200)
	//es.Submit(task3)

	inQueue := es.taskDispatcher.execPool.HowManyInQueue()
	util.Log(fmt.Sprintf("Inqueue tasks: %d", inQueue))

	time.Sleep(3 * time.Second)

	waitTillNoJobsInExecution(es.Monitor)

	util.Log(string(es.GetData().Data))

}

func TestExecutorServiceBasicSuccessCase(t *testing.T) {
	//t.SkipNow()
	assert := assert.New(t)

	taskDuration := 100
	task11 := NewBlockingTestTask(taskDuration, true)
	start := time.Now()
	err, resp := es.Submit(task11)
	end := time.Now()
	assert.Nil(err) // no error expected in submission
	assert.Equal(resp.Status, TaskStatusCompletedSuccessfully)

	dur := (end.Nanosecond() - start.Nanosecond()) / 1000

	assert.Greater(dur, taskDuration)
}

func TestExecutionServiceSubmissionFailure(t *testing.T) {
	assert := assert.New(t)
	cfg := createCommonTestCfg(es)
	cfg.Dispatcher.WaitForChanAvail = false
	testEs := cfg.MakeExecServiceFromCfg()
	testEs.Start()

	task1 := NewBlockingTestTask(3000, false)
	testEs.Submit(task1)
	task2 := NewBlockingTestTask(300, false)
	testEs.Submit(task2)
	err, _ := testEs.Submit(task2)
	// we should get error here
	assert.Errorf(err, "")
}
