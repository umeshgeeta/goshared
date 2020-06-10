// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

// Package executor contains basic implementation of an executor framework.
// The client facing interface is provided by the ExecutionService which
// creates ExecutorPool and Dispatcher. Task interface is defined so client
// would implement these methods on client specific Type to undertake specific
// work. When a task is submitted to the ExecutionService it submits it to the
// Dispatcher which assigns a channel on which task execution result / response
// is received. Along with attaching a channel to the task, it creates a
// separate routine to listen to response arrival. If the submitted task is
// blocking the calling routine is actually waiting for the response so
// Dispatcher returns the response upon receiving as well as undertakes the
// house keeping work of recycling the channel for another task. If the
// submitted task is asynchronous, caller is returned upon task submission
// even though there is a routine waiting for the response to undertake the
// house keeping.
//
// ExecutorPool maintains a set of executors for async tasks and another one for
// blocking tasks.
package executor

import (
	"errors"
	"fmt"
	"sync"
)

// We start with core Executor contract as an interface. As expected it has
// common methods like Start, Stop and Submit to receive a task. User can also
// specify whether we wait for availability of an internal buffer to accept the
// incoming task.
type Executor interface {
	Start()

	Submit(t Task) error

	HowManyInQueue() int

	WaitForAvailability(wfa bool)

	Stop()
}

// Executor configuration parameters
type ExecCfg struct {

	// How many maximum number tasks accepted by the executor when it is
	// already executing a task. These tasks will form the queue.
	TaskQueueCapacity int `json:"task_queue_capacity"`

	// If true, despite the full task queue capacity, caller invoking
	// Submit method will wait i.e. will be blocked. By default we keep it
	// false. So once the task queue is full, subsequent attempts to add a task
	// will fail as long as the queue if filled.
	WaitForAvailability bool `json:"wait_for_availability"`
}

// We model thread struct as a standard executor. It is a frugal attempt to
// model Java thread Object. The run method on this struct, a private method,
// so outside modules cannot call it directly; is basically an infinite loop
// of either waiting for a task or executing until the flag is turned off by
// the Stop method. All submitted tasks are funneled through a channel so
// 'waiting' for a task happens naturally.
type thread struct {
	continueRun         bool
	taskQueue           chan Task // incoming tasks on a channel
	queueCapacity       int
	waitForAvailability bool
	mux                 sync.Mutex
}

// Start the thread. We expect that callers would not call Start after having
// already called Stop. The pattern assumed is create new thread, start and
// stop. Multiple Starts and Stops are not supported at present.
func (t *thread) Start() {
	// First, create the channel which will queue the incoming tasks
	t.taskQueue = make(chan Task, t.queueCapacity)
	// Set the flag so as we continue to process the incoming tasks
	t.continueRun = true
	// do all that task execution in a different thread and
	// do not occupy the calling thread
	go t.run()
}

func (t *thread) run() {
	for t.continueRun {
		tsk := <-t.taskQueue
		if tsk != nil {
			rspChan := tsk.GetRespChan()
			if rspChan != nil {
				resp := tsk.Execute()
				// set the task is in response since we do not know
				// whether the task implementation may or many have set
				resp.TaskId = tsk.GetId()
				rspChan <- resp
			} else {
				// Every task is expected to have a channel, at least for tjr house keeping.
				// So regard this as an error condition. Since we do not have
				// response channel, no point in making the response object with
				// errors filled. For now, we simply log the error.
				fmt.Printf("Tasks %d has no channel to report back response.\n", tsk.GetId())
			}
		} else {
			fmt.Printf("Read nil task on the channel, channel probably closed.")
		}
	}
	fmt.Println("Exiting run")
}

func (t *thread) Submit(tsk Task) error {
	if !t.continueRun {
		return errors.New("executor is not started")
	}
	var err error = nil
	if t.waitForAvailability {
		// channel blocks naturally until the capacity is made available
		t.taskQueue <- tsk
	} else {
		t.mux.Lock()
		if t.HowManyInQueue() < t.queueCapacity {
			t.taskQueue <- tsk
		} else {
			err = errors.New("cannot submit, executor already has accepted maximum number of tasks")
		}
		t.mux.Unlock()
	}
	fmt.Printf("Submitted task %d successfully\n", tsk.GetId())
	return err
}

func (t *thread) Stop() {
	t.continueRun = false
	// also close the queue channel so no more tasks are accepted
	close(t.taskQueue)
}

func (t *thread) HowManyInQueue() int {
	return len(t.taskQueue)
}

func (t *thread) WaitForAvailability(wfa bool) {
	t.waitForAvailability = wfa
}

func NewExecutor(cfg ExecCfg) Executor {
	// We start a thread with 'continueRun' as false so that the caller needs to explicitly
	// invoke Start on the thread where queue capacity is built and any other initialization.
	t := new(thread)
	t.waitForAvailability = cfg.WaitForAvailability
	t.mux = sync.Mutex{}
	t.queueCapacity = cfg.TaskQueueCapacity
	return t
}

func (t *thread) IsRunning() bool {
	return t.continueRun
}
