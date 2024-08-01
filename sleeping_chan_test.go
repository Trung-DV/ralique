package ratelimit_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trung-dv/ratelimit"
)

func TestTicketPool(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		maxSlot      int
		windowTime   string
		concurNumber int
		workingTime  string
		expectDur    string
	}{
		{1, "0s", 1000, "0s", "1s"},
		{1, "1s", 3, "0s", "3s"},
		{1, "1s", 3, "2s", "6s"},
		{1, "1s", 3, "3s", "9s"},
		{1, "3s", 3, "0s", "9s"},
		{1, "1s", 5, "0s", "5s"},
		{2, "2s", 3, "0s", "4s"},
		{2, "1s", 5, "2s", "6s"},
		{2, "1s", 5, "0s", "3s"},
		{2, "3s", 5, "0s", "9s"},
		{2, "0s", 10, "0s", "1s"},
		{3, "2s", 10, "0s", "8s"},
		{10, "0s", 20, "0s", "1s"},
		{10, "1s", 60, "0s", "6s"},
		{100, "1s", 200, "0s", "2s"},
		{1000, "1s", 2000, "0s", "2s"},
		{1000, "1s", 3000, "0s", "3s"},
	} {
		t.Run(fmt.Sprintf("max_%d_items,window_%s,%d_concurr_take_%s,need_total_%s", tc.maxSlot, tc.windowTime, tc.concurNumber, tc.workingTime, tc.expectDur), func(t *testing.T) {
			t.Parallel()

			windowTime, err := time.ParseDuration(tc.windowTime)
			require.NoError(t, err)
			workingTime, err := time.ParseDuration(tc.workingTime)
			require.NoError(t, err)
			testDuration, err := time.ParseDuration(tc.expectDur)
			require.NoError(t, err)

			start := time.Now()
			rt := ratelimit.NewSleepingChan(tc.maxSlot, windowTime)
			wg := sync.WaitGroup{}
			for i := 0; i < tc.concurNumber; i++ {
				wg.Add(1)
				<-rt.EnqueueJobWithCallback(
					func() {
						time.Sleep(workingTime)
						t.Log("executed", time.Now())
					},
					wg.Done,
				)
			}

			wg.Wait()
			require.WithinDuration(t, start.Add(testDuration), time.Now(), time.Second)
		})
	}
}

func TestUsage(t *testing.T) {
	t.Parallel()
	rt := ratelimit.NewSleepingChan(2, time.Second)
	go rt.EnqueueJob(func() {
		t.Log(time.Now().String(), "first task done at")
	})
	go rt.EnqueueJob(func() {
		t.Log(time.Now().String(), "second task done at")
	})
	go rt.EnqueueJob(func() {
		t.Log(time.Now().String(), "third task done at")
	})

	time.Sleep(500 * time.Millisecond)
	go rt.EnqueueJob(func() {
		t.Log(time.Now().String(), "after sleep task done at")
	})
	time.Sleep(2 * time.Second)
}
