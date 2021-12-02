package gcron

import (
	"errors"
	"sync"
	"time"

	"github.com/yu31/timewheel"

	"github.com/yu31/gcron/pkg/expr"
)

const defaultTaskNum = 64

type Crontab struct {
	parser  expr.Parser
	tw      *timewheel.TimeWheel
	mu      *sync.Mutex
	jobs    map[string]*timewheel.Timer
	chain   Chain
	wg      *sync.WaitGroup
	running bool
}

// New creates an Crontab.
func New(opts ...Option) *Crontab {
	wg := new(sync.WaitGroup)
	c := &Crontab{
		parser:  expr.Standard,
		mu:      new(sync.Mutex),
		tw:      timewheel.Default(),
		jobs:    make(map[string]*timewheel.Timer, defaultTaskNum),
		chain:   Chain{chainRecover(), chainWaitGroup(wg)},
		wg:      wg,
		running: false,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Start starts the crontab in its own goroutine
func (c *Crontab) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}

	c.running = true
	c.tw.Start()
	c.mu.Unlock()
}

// Stop stops the crontab and wait for all tasks to completed.
func (c *Crontab) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}

	c.running = false
	c.tw.Stop()
	c.mu.Unlock()

	c.wg.Wait()
}

// ForceStop stop the crontab immediately.
func (c *Crontab) ForceStop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}

	c.running = false
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
	schedule.key = key
	c.add(key, schedule)
	return
}

// Interval adds or updates a job to the Crontab to be run on the given schedule Interval.
// The old job with the key will be stop and delete if exists.
func (c *Crontab) Interval(key string, schedule *Interval) (err error) {
	schedule.key = key
	if schedule.Interval < time.Millisecond {
		return errors.New("gcron: the interval minimum is 1ms")
	}
	c.add(key, schedule)
	return
}

// Once adds or updates a job to the Crontab to be run on the given schedule Once.
// The old job with the key will be stop and delete if exists.
func (c *Crontab) Once(key string, schedule *Once) (err error) {
	schedule.key = key
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
	c.jobs[key] = c.tw.Schedule(c.chain.Apply(schedule))
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
