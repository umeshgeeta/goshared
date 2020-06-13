// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

// The built-in conditional variable in Go - sync.Cond - broadcasts or signal
// the waiting go routine exactly once. If the waiting routine misses that, it
// waits infinitely and essentially your program hangs. There is a long running
// argument within the Go Community whether the conditional variable to be
// abolished altogether or not: https://github.com/golang/go/issues/21165
// Many within the Go Community want to abolish it altogether while the few
// other experienced professionals want to keep it and improve it. The proposers
// want to go all hog on Go channels. No doubt Channels is the marquee feature
// of the Go Language. I do not know whether the cost of 'channels' is all
// justified for the ubiquitous usage of Channel in Go applications. Coming from
// C and Java Programming background, all I want is working sync.Cond. What I
// have seen though, if Signal or Broadcast message is missed; my program is
// toast. Now Go Community may want to claim that they only entertain flawless
// coders so that a programmer knows exactly when to expect the 'signal'. In
// my unit tests when tasks finish very early only, I have seen that the cost
// of acquiring Lock is significant; meaning the waiting Go routine misses
// the Signal before starting to Wait. Hence the humble attempt - try multiple
// Broadcast so the listeners get more than one opportunity to get out of
// Wait loop. I agree, it is a weak solution. But it is helping in my use
// case where I want 'channels usage strictly managed'.
//
// The usage pattern is at any given time, all waits will be considered for a
// single condition only. We are trying to solve the problem of at least one
// 'wait' coming after the Broadcast. So repeat call on the Broadcast are not
// allowed until the first Broadcast is complete which is determined based on
// how many receipts we get from waiters.
//
// In cases where we are not concerned about whether all waiters respond back
// but at least one 'waiter' at least responds back. The use case is we want any
// one to pick up the resume the work upon completing of the work. It is
// essentially Broadcast with argument 1, which is simplified as Signal() call.
//
// CondVar bundles required locker with it so as you do not need to do explicit
// Lock and Unlock invocation around Wait.
package util

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Wrapper around the built-in Cond struct. All additional variables are
// for internal consumption only.
type CondVar struct {
	sync.Mutex
	cond          *sync.Cond
	gapInterval   int // in microseconds
	durationLimit int // in microseconds
	howManyLeft   int
}

// Create CondVar where the first argument is the gap between two subsequent
// broadcasts in microseconds. The second argument indicates how long to keep
// broadcasting, duration again in microseconds.
func NewCondVar(gi int, dl int) *CondVar {
	//defer LogDebug("NewCondVar constructed")
	cv := new(CondVar)
	cv.cond = sync.NewCond(cv)
	cv.gapInterval = gi
	cv.durationLimit = dl
	return cv
}

// Wait for a condition
func (cv *CondVar) Wait() {
	cv.Lock()
	cv.cond.Wait()
	cv.howManyLeft--
	cv.Unlock()
}

// Broadcast with how many 'receipts' from waiters are expected.
func (cv *CondVar) Broadcast(r int) error {
	if cv.howManyLeft > 0 {
		return errors.New("earlier broadcast not complete")
	}
	cv.Lock()
	cv.howManyLeft = r
	cv.Unlock()
	go func() {
		broadcast(cv)
	}()
	return nil
}

// If any first listener responds saying it got the signal, we are ok here.
func (cv *CondVar) Signal() error {
	return cv.Broadcast(1)
}

func broadcast(cv *CondVar) {
	start := time.Now().Nanosecond()
	for cv.howManyLeft > 0 && (time.Now().Nanosecond() < start+(cv.durationLimit*1000)) {
		cv.cond.Broadcast()
		time.Sleep(time.Duration(cv.gapInterval) * time.Microsecond)
	}
	if cv.howManyLeft > 0 {
		Log(fmt.Sprintf("Could not get receipt from all broadcast listeners. Left out: %d", cv.howManyLeft))
	}
}
