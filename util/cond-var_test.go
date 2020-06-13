// MIT License
// Author: Umesh Patil, Neosemantix, Inc.

package util

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func TestCondVar_Broadcast(t *testing.T) {

	// Because we have logging infrastructure in this package / directory;
	// we do not initialize it in Test Main in util_test file. We explicitly
	// initialize it here using ./../executor/static/default-cfg.json which
	// has a common logging settings.
	initilizeTestLog()

	assert := assert.New(t)
	sleepDuration := 1 // in microseconds
	ncv := NewCondVar(5, 500)

	wg.Add(2)
	waitInGoRoutine(ncv, sleepDuration)
	waitInGoRoutine(ncv, sleepDuration)
	ncv.Broadcast(2)
	// Because Broadcast has to obtain locks and spawn off another channel,
	// it takes time and during that period we assert that there are at least
	// 2 waiters on this cond var.
	assert.Equal(ncv.howManyLeft, 2)
	// Because go routine sleeps for sometime before starting the wait, waiter
	// count should be still non-zero and any call to broadcast again should fail.
	assert.Errorf(ncv.Broadcast(1), errMsgBdincomplete)
	// we wait
	wg.Wait()
	// all must have heard back from waiters
	assert.Equal(ncv.howManyLeft, 0)
}

func waitInGoRoutine(ncv *CondVar, sd int) {
	go func(n *CondVar, s int) {
		defer wg.Done()
		time.Sleep(time.Duration(s) * time.Microsecond)
		n.Wait()
		LogDebug("Exiting waitInGoRoutine")
	}(ncv, sd)
}
