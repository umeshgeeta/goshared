// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"encoding/json"
	"fmt"
	"github.com/umeshgeeta/goshared/util"
	"sync"
	"time"
)

// It tracks various task statistics submitted and executed through a dispatcher.
type TaskStats struct {
	sync.Mutex
	UpSinceWhen            time.Time `json:"up_since_when"`
	TotalTasksSubmitted    int       `json:"total_tasks_submitted"`
	BlockingTasksSubmitted int       `json:"blocking_tasks_submitted"`
	AsyncTasksSubmitted    int       `json:"async_tasks_submitted"`
	TasksInExecution       int       `json:"tasks_in_execution"`
}

// Create a new task stats (on purpose with lesser scope, only executor
// package can construct).
func newTaskStats() *TaskStats {
	ts := TaskStats{}
	ts.UpSinceWhen = time.Now()
	return &ts
}

func (ts *TaskStats) taskSubmitted(blocking bool) {
	ts.Lock()
	if blocking {
		ts.BlockingTasksSubmitted++
	} else {
		ts.AsyncTasksSubmitted++
	}
	ts.TotalTasksSubmitted++
	ts.TasksInExecution++
	ts.Unlock()
}

func (ts *TaskStats) taskDone(blocking bool) {
	ts.Lock()
	ts.TasksInExecution--
	ts.Unlock()
}

func (ts *TaskStats) byteArray() []byte {
	var result []byte
	ts.Lock()
	defer ts.Unlock()
	result, err := json.Marshal(ts)
	if err != nil {
		util.Log(fmt.Sprintf("Error in marshalling TaskStats: %v", err))
		return nil
	}
	return result
}

func taskStats(ba []byte) *TaskStats {
	var result TaskStats
	err := json.Unmarshal(ba, &result)
	if err != nil {
		util.Log(fmt.Sprintf("Failed to create TaskStats %v", err))
		return nil
	}
	return &result
}
