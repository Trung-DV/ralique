package ratelimit

import "time"

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

// EnqueueJob blocks until get a slot becomes available to execute the job.
// execFn is the primary task that will be executed asynchronously after the slot is successfully acquired.
func (q SleepingChan) EnqueueJob(execFn func()) {
	<-q.EnqueueJobWithCallback(execFn, nil)
}

// EnqueueJobWithCallback returns a channel that can be used externally to determine when to wait for an available slot.
// execFn is the primary task that will be executed asynchronously after the slot is successfully acquired.
// callback is a hook used to notify when a job has finished, including sleeping if necessary.
func (q SleepingChan) EnqueueJobWithCallback(execFn func(), callback func()) <-chan struct{} {
	enqueued := make(chan struct{})
	go func() {
		q.slots <- struct{}{}
		defer func() { <-q.slots }()

		close(enqueued)

		start := time.Now()

		if callback != nil {
			defer callback()
		}
		execFn()

		sleepDur := time.Until(start.Add(q.windowTime))
		time.Sleep(sleepDur)
	}()
	return enqueued
}
