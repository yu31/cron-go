package gcron

import (
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
	Callback func(value interface{}) error
}

func (task *Task) Run() error {
	return task.Callback(task.Value)
}
