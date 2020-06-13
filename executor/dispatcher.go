// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"errors"
	"fmt"
	"github.com/umeshgeeta/goshared/util"
)

// Dispatcher type which hold reference to executor pool, channels used for
// getting back task execution results and go routines waiting on task results.
type Dispatcher struct {
	execPool     *ExecutorPool
	respChans    *responseChannels
	chanCount    int
	waitForChan  bool
	waitingTasks map[int]waitingTask
	JobStats     *TaskStats
}

type DispatcherCfg struct {

	// Number of channels used to receive back task execution results
	ChannelCount int `json:"channel_count"`

	// Channel buffer size
	ChannelCapacity int `json:"channel_capacity"`

	// Whether caller should wait for response channel availability while
	// submitting a task
	WaitForChanAvail bool `json:"wait_for_chan_avail"`
}

// create a dispatcher with the given number of Response channel counts
// max channel count should be equal to number of executors in the executor pool
// if tasks are generally very short running, channels can be less
// we each dedicated channel for each task execution
//
// task submission on the dispatcher happens in the calling 'go routine / thread'
// (Alternate design could be waiting tasks are in map, we use one single fixed
// channel on which responses for all tasks are published and on the receiving
// side based on task id of the response object, we match waiting tasks in the
// map (key task id) and release the response.)
//
// cc: how many channels to create to listen back task result
//
// cp: capacity - buffer size - for each channel
//
// ep: executor pool
//
// wfc: whether to block the submission for availability a channel to hear back
// the task result. It does not apply for async tasks.
func NewDispatcher(cfg DispatcherCfg, ep *ExecutorPool) *Dispatcher {
	var disp Dispatcher
	disp.waitingTasks = make(map[int]waitingTask)
	disp.respChans = newRC(cfg.ChannelCount, cfg.ChannelCapacity, cfg.WaitForChanAvail, &disp.waitingTasks)
	disp.execPool = ep
	disp.waitForChan = cfg.WaitForChanAvail
	disp.chanCount = cfg.ChannelCount
	disp.JobStats = newTaskStats()
	return &disp
}

func (disp *Dispatcher) Start() {
	disp.execPool.Start()
	disp.respChans.start()
}

func (disp *Dispatcher) Submit(tsk Task) (error, *Response) {
	var err error = nil
	var resp *Response = nil
	if tsk != nil {
		err, resp = disp.submitTask(tsk)
		if err == nil {
			// there is no error in submitting the job, we start counting
			disp.JobStats.taskSubmitted(tsk.IsBlocking())
		}
	} else {
		err = errors.New("invalid task")
	}
	return err, resp
}

type waitingTask struct {
	cond             *util.CondVar
	responseReceived bool
	taskResponse     Response
	blocking         bool // whether task for which we will be waiting, is it blocking or not
}

func addNewWaitingTask(disp *Dispatcher, chanIndex int, tsk Task) *waitingTask {
	r := new(waitingTask)
	// track whether the task is blocking or not
	r.blocking = tsk.IsBlocking()
	r.cond = util.NewCondVar(10, 100)
	// update the internal map
	disp.waitingTasks[tsk.GetId()] = *r
	util.LogDebug(fmt.Sprintf("WaitingTask %v for task (id=%d) created", r, tsk.GetId()))
	// We start a go routine which will be waiting on this condition.
	// It is guaranteed that the go routine spawned will not go into infinite
	// loop because, the task is yet to be submitted. In other words, we do all
	// the house keeping work like setting up the listener for the response
	// before any response upon execution can be ever created.
	go func(wt *waitingTask) {
		for !wt.responseReceived {
			util.LogDebug(fmt.Sprintf("Starting wait. waitingTask: %v", wt))
			wt.cond.Wait()
			// read back from the map
			var updatedWt waitingTask = disp.waitingTasks[tsk.GetId()]
			wt = &updatedWt
			util.LogDebug(fmt.Sprintf("wait done! waitingTask: %v", wt))
		}
		util.LogDebug(fmt.Sprintf("Waiting done for cond %v", wt.cond))
		// get hold of the response....
		tr := wt.taskResponse
		// next remove the map entry
		delete(disp.waitingTasks, tr.TaskId)
		// and finally we need to mark channel as available
		disp.respChans.markAvailable(chanIndex)
		// as well as count the job done
		disp.JobStats.taskDone(wt.blocking)
	}(r)
	return r
}

func (disp *Dispatcher) submitTask(tsk Task) (error, *Response) {
	var err error = nil
	var resp *Response = nil
	// we have to get a channel on which we will wait for the response
	i, ai := disp.respChans.nextAvailChanIndex()
	if ai != nil {
		// we got a valid channel to use here
		// set in the task so executor can use
		tsk.SetRespChan(ai)
		// before submit task, create a listener to receive any response
		nwt := addNewWaitingTask(disp, i, tsk)
		// try submitting the task for the execution, we are waiting in nwt
		err = disp.execPool.Submit(tsk)
		// If no error, we have been able to submit successfully
		// and go routine is started to undertake house keeping
		// when the result comes back. We do not have anything here
		// do so we should return.
		if err != nil {
			// We got an error while submitting the job
			// so there will not be any response coming on this channel
			// we need to release the channel and destroy waiting task record
			// The way we achieve that is by setting a special error response
			// so that the house keeping go routine which is waiting will
			// exit and normal steps of house keeping will be executed.
			nwt.taskResponse = *FailedToSubmitResponse(tsk.GetId())
		} else {
			// else the task was submitted successfully with another go routine
			// waiting to undertake house keeping when the execution response
			// appears on the listening channel

			// However if this is a blocking task, we need to wait here
			// for the response from execution as well.
			if tsk.IsBlocking() {
				for !nwt.responseReceived {
					nwt.cond.Wait()
					// get the updated copy
					var updatedWt waitingTask = disp.waitingTasks[tsk.GetId()]
					nwt = &updatedWt
				}
				resp = &nwt.taskResponse
			}
		}
	} else {
		err = errors.New("cannot submit, no channel available")
	}
	return err, resp
}

func (disp *Dispatcher) Stop() {
	disp.execPool.Stop()
	disp.respChans.stop()
}
