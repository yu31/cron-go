package gcron

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yu31/timewheel"

	"github.com/yu31/gcron/pkg/expr"
)

type ScheduleFunc = timewheel.ScheduleFunc

// Schedule used in scheduler.
type Schedule interface {
	timewheel.Schedule
}

// UnixCron represents a periodic task with standard unix crontab expression.
type UnixCron struct {
	// Begin is the start time of the validity period of the job.
	// Zero means no limited.
	Begin time.Time

	// End is the end time of the validity period of the job.
	// Zero means no limited.
	End time.Time

	// Express is the crontab express specification.
	// Notice: It will panics if express is invalid.
	Express string

	once         sync.Once
	exprSchedule expr.Schedule // the exprSchedule of parse by crontab express.
}

// Next is called be timewheel.
func (job *UnixCron) Next(prev time.Time) time.Time {
	job.once.Do(func() {
		var err error
		job.exprSchedule, err = expr.Standard.Parse(job.Express)
		if err != nil {
			panic(fmt.Errorf("gcron: parse express error:%v", err))
		}
	})

	next := job.exprSchedule.Next(prev)
	// Not valid, advance the previous time to Begin.
	if !job.Begin.IsZero() && job.Begin.Sub(next) > 0 {
		next = job.Begin
	}

	// End of validity, return Zero.
	if !job.End.IsZero() && job.End.Sub(next) < 0 {
		return time.Time{}
	}
	return next
}

// Interval represents a periodic task with fixed interval
type Interval struct {
	// Begin is the start time of the validity period of the job.
	// Zero means no limited.
	Begin time.Time

	// End is the end time of the validity period of the job.
	// Zero means no limited.
	End time.Time

	// Interval is the time interval between each task.
	// The value cannot less than 10ms.
	Interval time.Duration
}

// Next is called be timewheel.
func (job *Interval) Next(prev time.Time) time.Time {
	next := prev.Add(job.Interval)

	// Not valid, advance the previous time to Begin.
	if !job.Begin.IsZero() && job.Begin.Sub(next) > 0 {
		next = job.Begin
	}

	// End of validity, return Zero.
	if !job.End.IsZero() && job.End.Sub(next) < 0 {
		return time.Time{}
	}
	return next
}

// Once used to perform the task at a specified time.
type Once struct {
	// Time is the task execute time.
	Time time.Time

	// done indicates whether the action has been performed.
	done int32
}

// Next is called be timewheel.
func (job *Once) Next(time.Time) time.Time {
	if atomic.CompareAndSwapInt32(&job.done, 0, 1) {
		return job.Time
	}
	return time.Time{}
}
