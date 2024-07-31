# Sliding Window Rate Limit Implementations

## Sleeping Channel:
Sleeping Channel is a technique combine of Semaphore and rest sleeping to make sure in any window time, there is no more than max numbers of job in handling.

Usage:

```go
import "github.com/trung-dv/ratelimit"
func main() {
	wg := sync.WaitGroup{}
	rt := ratelimit.NewSleepingChan(2, time.Second)
	println("")
	start := time.Now()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go rt.EnqueueJobWithCallback(
			func() {
				time.Sleep(500 * time.Millisecond)
			},
			wg.Done,
		)
	}
	wg.Done()
	println("total", time.Since(start))
}

```


You can add the pieces of code like a utils.

```go
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
```