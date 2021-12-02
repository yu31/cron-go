package gcron

import (
	"context"
	"math"
	"sync"
	"time"
)

// Chain is a sequence of ScheduleWrapper that decorates submitted jobs with
// cross-cutting behaviors like Retry.
type Chain []ScheduleWrapper

// Apply decorates the given schedule with all ScheduleWrapper in the chain.
//
// This:
//     WithWrapper(m1, m2, m3).Apply(Schedule)
// is equivalent to:
//     m1(m2(m3(Schedule)))
func (chain Chain) Apply(s Schedule) Schedule {
	n := len(chain) - 1
	for i := range chain {
		s = chain[n-i](s)
	}
	return s
}

// chainRecover panics in wrapped tasks and log them with the provided logger.
func chainRecover() ScheduleWrapper {
	return func(s Schedule) Schedule {
		return &wrapperRecover{Schedule: s}
	}
}

func chainWaitGroup(wg *sync.WaitGroup) ScheduleWrapper {
	return func(s Schedule) Schedule {
		return &wrapperWaitGroup{Schedule: s, wg: wg}
	}
}

// ChainRetry implements a wrapper to retry Run when error.
// If limit < 0 means no limit.
// The interval can not less than 100ms.
func ChainRetry(ctx context.Context, limit int64, interval time.Duration) ScheduleWrapper {
	if limit < 0 {
		limit = math.MaxInt64
	}
	if interval < time.Millisecond*100 {
		panic("gcron:ChainRetry: the interval can not less then 100ms")
	}
	return func(s Schedule) Schedule {
		return &wrapperRetry{
			Schedule: s,
			ctx:      ctx,
			limit:    limit,
			interval: interval,
		}
	}
}
