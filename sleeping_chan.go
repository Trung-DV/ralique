package ratelimit

import (
	"time"
)

type SleepingChan struct {
	slots chan struct{}

	windowTime time.Duration
}

// NewSleepingChan returns new SleepingChan with maximum slot
func NewSleepingChan(maxItem int, windowTime time.Duration) *SleepingChan {
	return &SleepingChan{
		windowTime: windowTime,
		slots:      make(chan struct{}, maxItem),
	}
}

// EnqueueJob waits to get a slot before doing job
// execFn is the main task need to be done after acquire slot success
func (q SleepingChan) EnqueueJob(execFn func()) {
	q.EnqueueJobWithCallback(execFn, nil)
}

// EnqueueJobWithCallback waits to get a slot before doing job
// execFn is the main task need to be done after acquire slot success
// callback is a hook to notify Job have finished
func (q SleepingChan) EnqueueJobWithCallback(execFn func(), callback func()) {
	q.slots <- struct{}{}
	defer func() {
		<-q.slots
	}()

	start := time.Now()

	if callback != nil {
		defer callback()
	}
	execFn()

	sleepDur := time.Until(start.Add(q.windowTime))
	time.Sleep(sleepDur)
}
