package gcron

// Option represents a modification to the default behavior of a Cron.
type Option func(cron *Crontab)

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
