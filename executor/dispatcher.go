// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"errors"
	"fmt"
	"github.com/umeshgeeta/goshared/util"
	"sync"
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
	r.cond = util.NewCondVar(40, 1000)
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

type responseChannels struct {
	responseChannels []chan Response

	// We track how many tasks are waiting on a 'ith' channel. Each channel has
	// the buffer capacity and it can buffer than many messages. When publisher
	// wants to push a message while the buffer is full, it gets blocked.
	// Meanwhile for another channel if there are not than many tasks waiting,
	// buffer is not full, we can use that channel for this task. So the way
	// we do this is by tracking how many tasks are waiting on that channel.
	// To be sure there is a difference between tasks waiting on a channel for
	// responses and actually how many elements are present in the buffer. The
	// buffer could be all empty which channel capacity number tasks would be
	// blocked on that channel for responses. So the key thing is to track how
	// many tasks are waiting on that channel regardless of 'len' of the channel
	// which actually givens messages sitting in the buffer.
	tasksWaitingOnChn        []int
	firstAvailable           int
	channelCount             int
	mux                      sync.Mutex
	chanAvail                *sync.Cond
	waitForChannel           bool
	continueRun              bool
	waitingTasksInDispatcher *map[int]waitingTask
}

func newRC(cc int, cp int, wfc bool, wtid *map[int]waitingTask) *responseChannels {
	var rc responseChannels
	rc.responseChannels = make([]chan Response, cc)
	for ch := range rc.responseChannels {
		rc.responseChannels[ch] = make(chan Response, cp)
	}
	rc.tasksWaitingOnChn = make([]int, cc)
	rc.channelCount = cc
	rc.waitForChannel = wfc
	rc.firstAvailable = 0
	rc.chanAvail = sync.NewCond(&rc.mux)
	rc.waitingTasksInDispatcher = wtid
	return &rc
}

func (rc *responseChannels) start() {
	rc.continueRun = true
	for ch := range rc.responseChannels {
		// We are starting 'listener go routines' for each of the channel
		// which run in the infinite loop until we stop the dispatcher.
		// The outer for loop is to start 'listener go routine' for each
		// channel channels provided. Infinite for loop is inside the go
		// routine where it keeps listening for various tasks over the lifetime.
		// In the infinite loop, every time we get a response, we dig
		// out the task id for which the response is received and then
		// we access the associated waiting task entry for that task id
		// from the dispatcher level map. Finally in the waiting task
		// entry we send the signal so as any routine which is waiting
		// on that entry will be notified. Typically for every submitted
		// task, there will be at least one go routine waiting - the
		// house keeping one. For blocked tasks, there will be an additional
		// go routine waiting; the main routine which invoked submit on
		// the dispatcher. For the Submit routine listener in case of
		// blocked routine, it would need the task result / response so
		// in this loop we set that once we get the response on the channel.
		go func(rci chan Response) {
			for rc.continueRun {
				var tr Response = <-rci
				var wt = (*rc.waitingTasksInDispatcher)[tr.TaskId]
				wt.taskResponse = tr
				wt.responseReceived = true
				// update back in the map
				(*rc.waitingTasksInDispatcher)[tr.TaskId] = wt
				if wt.blocking {
					wt.cond.Broadcast(2)
				} else {
					wt.cond.Broadcast(1)
				}
				util.LogDebug(fmt.Sprintf("Received response %v for taskId %d. Waiting task %v is signaled",
					wt.taskResponse, tr.TaskId, wt))
			}
		}(rc.responseChannels[ch])
	}
}

func (rc *responseChannels) stop() {
	rc.continueRun = false
	for ch := range rc.responseChannels {
		close(rc.responseChannels[ch])
	}
}

func (rc *responseChannels) markAvailable(ai int) {
	rc.mux.Lock()
	if rc.tasksWaitingOnChn[ai] > 0 {
		rc.tasksWaitingOnChn[ai]--
	} else {
		rc.tasksWaitingOnChn[ai] = 0
	}
	if rc.firstAvailable == -1 {
		rc.firstAvailable = ai
	}
	rc.mux.Unlock()
	// inform to other routines which are waiting
	rc.chanAvail.Signal()
}

func (rc *responseChannels) nextAvailChanIndex() (int, chan Response) {
	var result chan Response = nil
	var avlIndex = -1
	rc.mux.Lock()
	avlIndex = rc.firstAvailable
	if avlIndex > -1 {
		result = rc.pickFromAvailChannels(avlIndex)
	} else {
		if rc.waitForChannel {
			avlIndex, result = rc.waitForAChannel(avlIndex)
		}
		// else we did not find a channel and we are not going to wait; so should return
	}
	rc.mux.Unlock()
	return avlIndex, result
}

func (rc *responseChannels) waitForAChannel(avl int) (int, chan Response) {
	rc.mux.Lock()
	var avlIndex = avl
	for avlIndex == -1 {
		rc.chanAvail.Wait()
		avlIndex = rc.firstAvailable
	}
	var result chan Response = nil
	result = rc.pickFromAvailChannels(avlIndex)
	rc.mux.Unlock()
	return avlIndex, result
}

func (rc *responseChannels) pickFromAvailChannels(avlIndex int) chan Response {
	// caller has the lock, so we do not worry about it
	var result chan Response = nil
	result = rc.responseChannels[avlIndex]
	i := rc.next(avlIndex)
	min := rc.tasksWaitingOnChn[i]
	minIndex := i
	cap := cap(rc.responseChannels[0]) // all channels are of same capacity
	found := false
	for !found && i != rc.firstAvailable {
		nt := rc.tasksWaitingOnChn[i]
		if nt == 0 {
			found = true
		} else {
			if nt < cap && nt < min {
				minIndex = i
			}
		}
		i = rc.next(i)
	}
	if i == rc.firstAvailable && !found {
		// we scanned the entire array and
		// did not find any available channel
		rc.firstAvailable = -1
	} else {
		// got a valid one which is not busy,
		// we mark that as next available channel index
		rc.firstAvailable = minIndex
	}
	// in any case we need to increase the count of tasks waiting on the channel
	rc.tasksWaitingOnChn[avlIndex]++
	return result
}

func (rc *responseChannels) next(i int) int {
	j := i
	j++
	if j == rc.channelCount {
		j = 0
	}
	return j
}
