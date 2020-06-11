// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"fmt"
	"math/rand"
	"time"
)

const TestTaskExecDurationLowerLimit = 8
const TestTaskExecDurationUpperLimit = 128
const TestTaskDurationRange = TestTaskExecDurationUpperLimit - TestTaskExecDurationLowerLimit

var taskIdCounter int

type TestTask struct {
	id           int
	blocking     bool
	execDuration int
	rc           chan Response
}

var GlobalRand *rand.Rand

func SetupRand() {
	// Create and seed the generator.
	// Typically a non-fixed seed should be used, such as time.Now().UnixNano().
	// Using a fixed seed will produce the same output on every run.
	GlobalRand = rand.New(rand.NewSource(99))
}

func GetRandomBoolean() bool {
	r := GlobalRand.Intn(2)
	if r == 0 {
		return false
	} else {
		return true
	}
}

func (tt *TestTask) GetId() int {
	return tt.id
}

func (tt *TestTask) Execute() Response {
	resp := NewResponse(tt.id)
	fmt.Printf("Start of execution for task id:	%d at time: %v\n", tt.id, time.Now().Nanosecond())
	time.Sleep(time.Duration(tt.execDuration) * time.Microsecond)
	fmt.Printf("  End of execution for task id:	%d at time:	%v\n", tt.id, time.Now().Nanosecond())
	// we regard the task is completed successfully, so we set the status
	resp.Status = TaskStatusCompletedSuccessfully
	return *resp
}

func (tt *TestTask) SetRespChan(rc chan Response) {
	tt.rc = rc
}

func (tt *TestTask) GetRespChan() chan Response {
	return tt.rc
}

func (tt *TestTask) IsBlocking() bool {
	return tt.blocking
}

func SetupTestTask() {
	taskIdCounter = 0
}

func NewTestTask(ed int) *TestTask {
	return NewBlockingTestTask(ed, GetRandomBoolean())
}

func NewBlockingTestTask(ed int, blocking bool) *TestTask {
	tt := new(TestTask)
	taskIdCounter = +nextTaskId()
	tt.id = taskIdCounter
	tt.blocking = blocking
	tt.execDuration = ed
	return tt
}

func nextTaskId() int {
	taskIdCounter += 1
	return taskIdCounter
}

func RandomTestTaskExecTime() int {
	return GlobalRand.Intn(TestTaskDurationRange)
}
