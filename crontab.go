package gcron

import (
	"errors"
	"sync"
	"time"

	"github.com/yu31/timewheel"

	"github.com/yu31/gcron/pkg/expr"
)

const defaultJobNum = 64

type Crontab struct {
	parser expr.Parser
	tw     *timewheel.TimeWheel
	mu     *sync.Mutex
	jobs   map[string]*timewheel.Timer
}

// New creates an Crontab.
func New() *Crontab {
	return &Crontab{
		parser: expr.Standard,
		mu:     new(sync.Mutex),
		tw:     timewheel.Default(),
		jobs:   make(map[string]*timewheel.Timer, defaultJobNum),
	}
}

func (c *Crontab) WithParser(parser expr.Parser) *Crontab {
	c.mu.Lock()
	c.parser = parser
	c.mu.Unlock()
	return c
}

// Reset do reset the time wheel and jobs.
func (c *Crontab) Reset() {
	c.mu.Lock()
	if len(c.jobs) > 0 {
		c.tw.Stop()
		// reset time wheel.
		c.tw = timewheel.Default()
		c.jobs = make(map[string]*timewheel.Timer, defaultJobNum)
	}
	c.mu.Unlock()
}

// Start starts the current crontab.
func (c *Crontab) Start() {
	c.mu.Lock()
	c.tw.Start()
	c.mu.Unlock()
}

// Stop stops the current crontab and wait until timing exit.
//
// If there is any timer's task being running in its own goroutine, Stop does
// not wait for the task to complete before returning. If the caller needs to
// know whether the task is completed, it must coordinate with the task explicitly.
func (c *Crontab) Stop() {
	c.mu.Lock()
	c.tw.Stop()
	c.mu.Unlock()
}

// Add adds or updates a job to the Crontab to be run on the given Schedule.
// The old job with the key will be stop and delete if exists.
func (c *Crontab) Add(key string, schedule Schedule) (err error) {
	c.add(key, schedule)
	return
}

// Express adds or updates a job to the Crontab to be run on the given schedule Express.
// The old job with the key will be stop and delete if exists.
func (c *Crontab) Express(key string, schedule *Express) (err error) {
	// Check the crontab express spec and build cron.Schedule.
	schedule.schedule, err = c.parser.Parse(schedule.Express)
	if err != nil {
		return
	}
	c.add(key, schedule)
	return
}

// Interval adds or updates a job to the Crontab to be run on the given schedule Interval.
// The old job with the key will be stop and delete if exists.
func (c *Crontab) Interval(key string, schedule *Interval) (err error) {
	if schedule.Interval < time.Millisecond {
		return errors.New("gcron: the interval minimum is 1ms")
	}
	c.add(key, schedule)
	return
}

// Once adds or updates a job to the Crontab to be run on the given schedule Once.
// The old job with the key will be stop and delete if exists.
func (c *Crontab) Once(key string, schedule *Once) (err error) {
	c.add(key, schedule)
	return
}

// Remove delete and stop the job with specified id.
func (c *Crontab) Remove(key string) (err error) {
	c.remove(key)
	return
}

func (c *Crontab) add(key string, schedule Schedule) {
	if key == "" {
		panic("gcron: key cannot be empty")
	}
	c.mu.Lock()
	// Stops old job if exists before.
	if old, ok := c.jobs[key]; ok {
		old.Close()
	}
	// Adds and start the new job.
	c.jobs[key] = c.tw.Schedule(schedule)
	c.mu.Unlock()
}

func (c *Crontab) remove(key string) {
	c.mu.Lock()
	if old, ok := c.jobs[key]; ok {
		old.Close()
		delete(c.jobs, key)
	}
	c.mu.Unlock()
}
