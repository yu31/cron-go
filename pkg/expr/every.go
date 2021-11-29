package expr

import "time"

// everySchedule represents a simple recurring duty cycle, e.g. "Every 5 minuteBounds".
// It does not support jobs more frequent than once a second.
type everySchedule struct {
	interval time.Duration
}

// Next returns the next time this should be run.
// This rounds so that the next activation time will be on the second.
func (schedule everySchedule) Next(t time.Time) time.Time {
	return t.Add(schedule.interval)
}

// Every returns a crontab Schedule that activates once every duration.
func Every(interval time.Duration) Schedule {
	return everySchedule{interval: interval}
}
