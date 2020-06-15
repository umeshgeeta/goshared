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

func createExecServiceWithTestCommonCfg(es *ExecutionService) *ExecutionService {
	cfg := createCommonTestCfg(es)
	return cfg.MakeExecServiceFromCfg()
}

func createCommonTestCfg(es *ExecutionService) *ExecServiceCfg {
	cfg := es.CloneCfg()
	cfg.Dispatcher.ChannelCount = 1
	cfg.Dispatcher.ChannelCapacity = 1
	cfg.ExexPool.AsyncTaskExecutorCount = 1
	cfg.ExexPool.BlockingTaskExecutorCount = 1
	cfg.Executor.TaskQueueCapacity = 1
	return cfg
}
