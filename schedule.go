package gcron

import (
	"context"
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

	// Concurrency indicates whether allow multi task execute concurrently.
	Concurrency bool

	// Express is the crontab express specification.
	Express string

	// Context
	Context context.Context

	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback func(ctx context.Context, value interface{})

	schedule expr.Schedule
	// running indicates whether the task func is running. 1 => true, 0 => false.
	running int32
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
	return job.schedule.Next(prev)
}

// Run is called by timewheel.
func (job *Express) Run() {
	// To avoids run repeat.
	if !job.Concurrency && !atomic.CompareAndSwapInt32(&job.running, 0, 1) {
		return
	}
	job.Callback(job.Context, job.Value)
	if !job.Concurrency {
		atomic.StoreInt32(&job.running, 0)
	}
}

// Interval represents a periodic task with fixed interval
type Interval struct {
	// Begin is the start time of the validity period of the job.
	// Zero means no limited.
	Begin time.Time

	// End is the end time of the validity period of the job.
	// Zero means no limited.
	End time.Time

	// Concurrency indicates whether allow multi task execute concurrently.
	Concurrency bool

	// Interval is the time interval between each task.
	Interval time.Duration

	// Context
	Context context.Context

	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback func(ctx context.Context, value interface{})

	// running indicates whether the task func is running. 1 => true, 0 => false.
	running int32
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

// Run is called by timewheel.
func (job *Interval) Run() {
	// To avoids run repeat.
	if !job.Concurrency && !atomic.CompareAndSwapInt32(&job.running, 0, 1) {
		return
	}
	job.Callback(job.Context, job.Value)
	if !job.Concurrency {
		atomic.StoreInt32(&job.running, 0)
	}
}

// Once used to perform the task at a specified time.
type Once struct {
	Time time.Time

	// Context
	Context context.Context

	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback func(ctx context.Context, value interface{})

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

// Run is called by timewheel.
func (job *Once) Run() {
	job.Callback(job.Context, job.Value)
}
