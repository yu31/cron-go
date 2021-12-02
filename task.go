package gcron

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/yu31/timewheel"

	"github.com/yu31/gcron/pkg/expr"
)

type Callback func(ctx context.Context, key string, value interface{}) error

// Schedule used in scheduler.
type Schedule interface {
	timewheel.Schedule
	Context() context.Context // return the context.
	Err() error               // return non-nil when run error.
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

	// Ctx
	Ctx context.Context

	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback Callback

	// the task key that caller set.
	key string
	// the schedule of parse by crontab express.
	schedule expr.Schedule
	// running indicates whether the task func is running. 1 => true, 0 => false.
	running int32
	// callback error.
	err error
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

	job.err = job.Callback(job.Ctx, job.key, job.Value)

	if !job.Concurrency {
		atomic.StoreInt32(&job.running, 0)
	}
}

func (job *Express) Context() context.Context {
	return job.Ctx
}

func (job *Express) Err() error {
	return job.err
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

	// Ctx
	Ctx context.Context

	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback Callback

	// the task key that caller set.
	key string
	// running indicates whether the task func is running. 1 => true, 0 => false.
	running int32
	// callback error.
	err error
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
	job.err = job.Callback(job.Ctx, job.key, job.Value)
	if !job.Concurrency {
		atomic.StoreInt32(&job.running, 0)
	}
}

func (job *Interval) Context() context.Context {
	return job.Ctx
}

func (job *Interval) Err() error {
	return job.err
}

// Once used to perform the task at a specified time.
type Once struct {
	// Time is the task execute time.
	Time time.Time

	// Ctx
	Ctx context.Context

	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback Callback

	// the task key that caller set.
	key string
	// done indicates whether the action has been performed.
	done int32
	// callback error.
	err error
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
	job.err = job.Callback(job.Ctx, job.key, job.Value)
}

func (job *Once) Context() context.Context {
	return job.Ctx
}

func (job *Once) Err() error {
	return job.err
}
