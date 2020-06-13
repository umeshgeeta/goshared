// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package executor

import (
	"fmt"
	"github.com/umeshgeeta/goshared/util"
	"sync"
)

// Strictly integrally used structure and methods to manage a fixed set of
// channels to be used for getting back task responses.
type responseChannels struct {

	// The choice of data structure is a simple array. Did evaluate sync.Pool
	// structure, but the comment thread at:
	// https://github.com/golang/go/issues/23199
	// scared me. It is true this thread is talking about byte buffers whereas
	// in this case it will be a pool of channels, concrete instances without
	// any changing sizes. However, I saw to difficulties:
	// - I would have to implement channel count enforcement in New and
	// - could not get handle on how to utilize all channels in maximum way.
	// Remember, we allow flexible channel count so that on a single channel
	// more than one task responses can be waited for. When you do the first
	// Get channel call on the Pool, we can say start listening to 4 task
	// responses (say capacity is 4). For the 5th task we go Get another
	// channel from the Pool. But as results on the first channel starts to
	// come, that channel will be available for more tasks. This would mean
	// any time we want a channel for a new task, we have to walk through
	// already pulled out channels to see if any space capacity and so.
	// Overall lot of logic will still have to be build outside of Pool and
	// hence decided to stick with this simple data structure and related logic.
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
	chanAvail                *util.CondVar
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
	rc.chanAvail = util.NewCondVar(10, 100)
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
					// it is house keeping routine + the original caller
					wt.cond.Broadcast(2)
				} else {
					// it is house keeping routine only
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
		// it is actually an error condition, there is a defect lying here
		// let us log for now to find more
		util.Log("unexpected - marking a channel available which is already free")
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
	avlIndex = rc.firstAvailable
	if avlIndex > -1 {
		result = rc.pickFromAvailChannels(avlIndex)
	} else {
		if rc.waitForChannel {
			avlIndex, result = rc.waitForAChannel(avlIndex)
		}
		// else we did not find a channel and we are not going to wait; so should return
	}
	return avlIndex, result
}

func (rc *responseChannels) waitForAChannel(avl int) (int, chan Response) {
	var avlIndex = avl
	for avlIndex == -1 {
		rc.chanAvail.Wait()
		avlIndex = rc.firstAvailable
	}
	var result chan Response = nil
	result = rc.pickFromAvailChannels(avlIndex)
	return avlIndex, result
}

func (rc *responseChannels) pickFromAvailChannels(avlIndex int) chan Response {
	// caller has the lock, so we do not worry about it
	var result chan Response = nil
	rc.mux.Lock()
	defer rc.mux.Unlock()
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
				min = nt
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
