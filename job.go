package gcron

import (
	"context"

	"github.com/yu31/timewheel"
)

var _ Job = (*Task)(nil)

type Job interface {
	timewheel.Job

	WithContext(ctx context.Context) // Set the context.
	Context() context.Context        // Return the context of job.
	TaskKey() string                 // Return the task key of job.
	TaskValue() interface{}          // Return the task value of job.
}

// Task represents a Job implementation.
type Task struct {
	// Ctx
	Ctx context.Context

	// the task key that invoker set.
	Key string

	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback func(ctx context.Context, key string, value interface{}) error
}

func (task *Task) Run() error {
	return task.Callback(task.Ctx, task.Key, task.Value)
}

func (task *Task) WithContext(ctx context.Context) {
	task.Ctx = ctx
}

func (task *Task) Context() context.Context {
	return task.Ctx
}

func (task *Task) TaskKey() string {
	return task.Key
}

func (task *Task) TaskValue() interface{} {
	return task.Value
}
