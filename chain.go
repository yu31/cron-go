package gcron

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
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
func (chain JobChain) Apply(job Job) Job {
	n := len(chain) - 1
	for i := range chain {
		job = chain[n-i](job)
	}
	return job
}

// chainRecover panics in wrapped tasks and log them with the provided logger.
func chainRecover() JobWrapper {
	return func(job Job) Job {
		return &wrapperRecover{Job: job}
	}
}

func chainWaitGroup(wg *sync.WaitGroup) JobWrapper {
	return func(job Job) Job {
		return &wrapperWaitGroup{Job: job, wg: wg}
	}
}

// ChainRetry implements a wrapper to retry Run when error.
// If limit < 0 means no limit.
// The interval can not less than 100ms.
func ChainRetry(ctx context.Context, limit int64, interval time.Duration) JobWrapper {
	if limit < 0 {
		limit = math.MaxInt64
	}
	if interval < time.Millisecond*100 {
		panic("gcron:ChainRetry: the interval can not less then 100ms")
	}
	return func(job Job) Job {
		return &wrapperRetry{
			Job:      job,
			ctx:      ctx,
			limit:    limit,
			interval: interval,
		}
	}
}

// ChainSkipIfRunning skips an invocation of the Job if a previous invocation is
// still running.
func ChainSkipIfRunning() JobWrapper {
	return func(job Job) Job {
		return &wrapperSkipIfRunning{
			Job:     job,
			running: 0,
		}
	}
}

// ChainDelayIfRunning serializes jobs, delaying subsequent runs until the
// previous one is complete.
func ChainDelayIfRunning() JobWrapper {
	return func(job Job) Job {
		return &wrapperDelayIfRunning{
			Job: job,
			mu:  sync.Mutex{},
		}
	}
}

// ChainOpentracing implements a wrapper to supported opentracing.
func ChainOpentracing(tracer opentracing.Tracer) JobWrapper {
	return func(job Job) Job {
		return &wrapperOpentracing{
			Job:    job,
			tracer: tracer,
		}
	}
}
