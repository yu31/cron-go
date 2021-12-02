package gcron

import "github.com/yu31/gcron/pkg/expr"

// Option represents a modification to the default behavior of a Cron.
type Option func(c *Crontab)

func WithParser(parser expr.Parser) Option {
	return func(c *Crontab) {
		c.parser = parser
	}
}

func WithWrapper(w ...ScheduleWrapper) Option {
	return func(c *Crontab) {
		c.chain = append(c.chain, w...)
	}
}
