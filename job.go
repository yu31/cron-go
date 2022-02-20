package gcron

import (
	"context"

	"github.com/yu31/timewheel"
)

var (
	_ Job = (*Task)(nil)
)

type JobFunc = timewheel.JobFunc

type Job interface {
	timewheel.Job
}

// Task represents a Job implementation.
type Task struct {
	// Value is the arguments used with callback function.
	Value interface{}

	// Callback called when job expired.
	Callback func(ctx context.Context, value interface{}) error
}

func (task *Task) Run(ctx context.Context) error {
	return task.Callback(ctx, task.Value)
}
