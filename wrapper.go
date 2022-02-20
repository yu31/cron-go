package gcron

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type JobWrapper func(job Job) Job

// JobChain is a sequence of JobWrapper that decorates submitted jobs with
type JobChain []JobWrapper

// Apply decorates the given exprSchedule with all JobWrapper in the jobChain.
//
// This:
//     WithJobWrapper(m1, m2, m3).Apply(Job)
// is equivalent to:
//     m1(m2(m3(Job)))
func (jobChain JobChain) Apply(job Job) Job {
	n := len(jobChain) - 1
	for i := range jobChain {
		job = jobChain[n-i](job)
	}
	return job
}

// WrapJobWaitGroup implements a JobWrapper to track the completion of job with sync.WaitGroup.
func WrapJobWaitGroup(wg *sync.WaitGroup) JobWrapper {
	return func(job Job) Job {
		return JobFunc(func() error {
			wg.Add(1)
			err := job.Run()
			wg.Done()
			return err
		})
	}
}

// WrapJobRecover implements a JobWrapper to recover when Job run panics.
func WrapJobRecover() JobWrapper {
	return func(job Job) Job {
		return JobFunc(func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					var ok bool
					err, ok = r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					err = fmt.Errorf("gcron: job run panic with error: %v", err)

					const size = 64 << 10
					buf := make([]byte, size)
					i := runtime.Stack(buf, true)
					buf = buf[:i]
					// FIXME: use logger.
					println(fmt.Sprintf("%v\n%s", err, string(buf)))
				}
			}()
			err = job.Run()
			return
		})
	}
}

// WrapJobRetry implements a JobWrapper to retry Run when any error.
// The limit <= 0 means no limit.
// The interval not allowed to be 0.
func WrapJobRetry(ctx context.Context, limit int64, interval time.Duration) JobWrapper {
	if limit <= 0 {
		limit = math.MaxInt64
	}
	if interval == 0 {
		panic("gcron: WrapJobRetry: the interval must be greater than 0")
	}
	return func(job Job) Job {
		return JobFunc(func() (err error) {
			if err = job.Run(); err == nil {
				return
			}
			if limit <= 0 {
				return
			}

			ticker := time.NewTicker(interval)
		LOOP:
			for i := int64(0); i < limit; i++ {
				select {
				case <-ticker.C:
					if err = job.Run(); err == nil {
						break LOOP
					}
				case <-ctx.Done():
					break LOOP
				}
			}
			ticker.Stop()
			return
		})
	}
}

// WrapJobSkipIfRunning implements a JobWrapper to skip an invocation of the Job if
// previous invocation is still running.
func WrapJobSkipIfRunning() JobWrapper {
	return func(job Job) Job {
		// running indicates whether the job func is running. 1 => true, 0 => false.
		var running int32

		return JobFunc(func() error {
			if !atomic.CompareAndSwapInt32(&running, 0, 1) {
				return nil
			}
			err := job.Run()
			atomic.StoreInt32(&running, 0)
			return err
		})
	}
}

// WrapJobBlockIfRunning implements a JobWrapper to block an invocation of the job util
// the previous one is completed.
func WrapJobBlockIfRunning() JobWrapper {
	return func(job Job) Job {
		var mu sync.Mutex

		return JobFunc(func() error {
			mu.Lock()
			err := job.Run()
			mu.Unlock()
			return err
		})
	}
}
