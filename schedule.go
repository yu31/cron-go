package gcron

import (
	"sync/atomic"
	"time"

	"github.com/yu31/timewheel"

	"github.com/yu31/gcron/pkg/expr"
)

// Schedule used in scheduler.
type Schedule interface {
	timewheel.Schedule
}

// Express represents a periodic task with crontab express.
type Express struct {
	// Begin is the start time of the validity period of the job.
	// Zero means no limited.
	Begin time.Time

	// End is the end time of the validity period of the job.
	// Zero means no limited.
	End time.Time

	// Express is the crontab express specification.
	Express string

	// the exprSchedule of parse by crontab express.
	exprSchedule expr.Schedule
}

// Next is called be timewheel.
func (job *Express) Next(prev time.Time) time.Time {
	// End of validity, return Zero.
	if !job.End.IsZero() && prev.After(job.End) {
		return time.Time{}
	}
	// Not valid, advance the previous time to Begin.
	if !job.Begin.IsZero() && prev.Before(job.Begin) {
		prev = job.Begin
	}
	return job.exprSchedule.Next(prev)
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
	// End of validity, return Zero.
	if !job.End.IsZero() && prev.After(job.End) {
		return time.Time{}
	}
	// Not valid, advance the previous time to Begin.
	if !job.Begin.IsZero() && prev.Before(job.Begin) {
		prev = job.Begin
	}
	return prev.Add(job.Interval)
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
