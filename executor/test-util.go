// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import "github.com/umeshgeeta/goshared/util"

func waitTillNoJobsInExecution(montior *util.Monitor) {
	done := false
	for !done {
		blob := <-montior.MonDataChan
		ts := taskStats(blob.Data)
		if ts.TasksInExecution == 0 {
			done = true
		}
	}
}
