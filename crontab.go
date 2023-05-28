package cron

import (
	"context"
	"sync"
	"time"

	"github.com/yu31/timewheel-go"
)

type Crontab struct {
	mu       *sync.Mutex
	tw       *timewheel.TimeWheel
	jobs     map[string]*timewheel.Timer
	jobChain JobChain
	location *time.Location
}

// New creates a Crontab.
func New(opts ...Option) *Crontab {
	cron := &Crontab{
		mu:       new(sync.Mutex),
		tw:       nil,
		jobs:     make(map[string]*timewheel.Timer, 64),
		jobChain: nil,
		location: time.Local,
	}
	for _, opt := range opts {
		opt(cron)
	}
	cron.tw = timewheel.Default(timewheel.WithTimezone(cron.location))
	return cron
}

// Start starts the crontab in its own goroutine
func (cron *Crontab) Start() {
	cron.mu.Lock()
	cron.tw.Start()
	cron.mu.Unlock()
}

// Stop stops the crontab.
//
// Notice: By default, Stop does not wait for the running job completed.
// You can use WrapJobWaitGroup to track the completion of jobs.
func (cron *Crontab) Stop() {
	if cron == nil {
		return
	}
	cron.mu.Lock()
	cron.tw.Stop()
	cron.mu.Unlock()
}

// Submit adds or updates a job to the Crontab to be run on the given Schedule.
// The old job with the key will be stopped and delete if exists.
func (cron *Crontab) Submit(ctx context.Context, key string, job Job, schedule Schedule) {
	if key == "" {
		panic("gcron: key cannot be empty")
	}
	cron.mu.Lock()
	// Stops old job if exists before.
	if old, ok := cron.jobs[key]; ok {
		old.Close()
	}
	// Adds and start the new job.
	cron.jobs[key] = cron.tw.ScheduleJob(ctx, schedule, cron.jobChain.Apply(job))
	cron.mu.Unlock()
}

// Remove delete and stop the job with specified id.
func (cron *Crontab) Remove(key string) {
	cron.mu.Lock()
	if old, ok := cron.jobs[key]; ok {
		old.Close()
		delete(cron.jobs, key)
	}
	cron.mu.Unlock()
}
