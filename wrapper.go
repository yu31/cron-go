package gcron

import (
	"context"
	"fmt"
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
		return JobFunc(func(ctx context.Context) error {
			wg.Add(1)
			err := job.Run(ctx)
			wg.Done()
			return err
		})
	}
}

// WrapJobRecover implements a JobWrapper to recover when Job run panics.
func WrapJobRecover() JobWrapper {
	return func(job Job) Job {
		return JobFunc(func(ctx context.Context) (err error) {
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
			err = job.Run(ctx)
			return
		})
	}
}

// WrapJobRetry implements a JobWrapper to retry Run when any error.
// The limit < 0 means no limited.
// The interval not allowed must be greater than 0.
func WrapJobRetry(ctxRetry context.Context, limit int64, interval time.Duration) JobWrapper {
	if limit != 0 && interval <= 0 {
		panic("gcron: WrapJobRetry: the interval must be greater than 0")
	}

	return func(job Job) Job {
		return JobFunc(func(ctx context.Context) (err error) {
			if err = job.Run(ctx); err == nil {
				return
			}
			if limit == 0 {
				return
			}

			ticker := time.NewTicker(interval)
			i := int64(0)
		LOOP:
			for {
				select {
				case <-ticker.C:
					if err = job.Run(ctx); err == nil {
						break LOOP
					}
				case <-ctxRetry.Done():
					break LOOP
				case <-ctx.Done():
					break LOOP
				}
				if limit < 0 {
					// The `limit` < 0 means no retry limited.
					continue LOOP
				}
				i++
				if i >= limit {
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

		return JobFunc(func(ctx context.Context) error {
			if !atomic.CompareAndSwapInt32(&running, 0, 1) {
				return nil
			}
			err := job.Run(ctx)
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

		return JobFunc(func(ctx context.Context) error {
			mu.Lock()
			err := job.Run(ctx)
			mu.Unlock()
			return err
		})
	}
}
