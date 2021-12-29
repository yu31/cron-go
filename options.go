package gcron

import "time"

// Option represents a modification to the default behavior of a Cron.
type Option func(cron *Crontab)

// WithTimezone reset the timezone in Cron.
func WithTimezone(loc *time.Location) Option {
	return func(cron *Crontab) {
		cron.location = loc
	}
}

// WithJobWrapper append JobWrapper into jobChain
func WithJobWrapper(w ...JobWrapper) Option {
	return func(cron *Crontab) {
		cron.jobChain = append(cron.jobChain, w...)
	}
}

// WithJobChain overwrite the jobChain.
func WithJobChain(jobChain JobChain) Option {
	return func(cron *Crontab) {
		cron.jobChain = jobChain
	}
}
