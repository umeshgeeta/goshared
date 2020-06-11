// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestExecutorFailSubmission(t *testing.T) {
	//t.SkipNow()
	execCfg := &ExecCfg{
		TaskQueueCapacity:   2,
		WaitForAvailability: false,
	}
	thread := NewExecutor(*execCfg)
	task := NewTestTask(RandomTestTaskExecTime())
	err := thread.Submit(task)
	assert := assert.New(t)
	assert.NotNil(err, "Expected error: %v\n", err)
}

func TestExecutorExecutionSuccess(t *testing.T) {
	execCfg := &ExecCfg{
		TaskQueueCapacity:   2,
		WaitForAvailability: false,
	}

	thread := NewExecutor(*execCfg)
	thread.Start()

	task := NewTestTask(10000)
	ch := make(chan Response)
	task.SetRespChan(ch)

	err := thread.Submit(task)

	var wg sync.WaitGroup
	wg.Add(2)

	if err == nil {
		go waitForResponse(ch, t, TaskStatusCompletedSuccessfully, "Tasks Status as expected.", &wg)
	} else {
		t.Errorf("Task submission error: %v\n", err)
	}

	task2 := NewTestTask(10)
	ch2 := make(chan Response)
	task2.SetRespChan(ch2)
	err2 := thread.Submit(task2)

	if err2 == nil {
		fmt.Printf("In queue: %d\n", thread.HowManyInQueue())
		go waitForResponse(ch2, t, TaskStatusCompletedSuccessfully, "Tasks Status as expected.", &wg)
	} else {
		t.Errorf("Task submission error: %v\n", err)
	}

	wg.Wait()
	fmt.Println("wait wg complete")
	thread.Stop()
	fmt.Println("thread stopped, everything is done")
}

func waitForResponse(ch chan Response, t *testing.T, ts int, msg string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("waitForResponse Start")
	rsp := <-ch
	assert := assert.New(t)
	assert.Equal(rsp.Status, ts, msg)
	fmt.Println("waitForResponse Done")
}
