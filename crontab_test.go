package gcron

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrontab_New(t *testing.T) {
	cron := New()
	require.NotNil(t, cron.mu)
	require.NotNil(t, cron.tw)
	require.NotNil(t, cron.jobs)
}

func TestCrontab_StartAndStop(t *testing.T) {
	cron := New()
	require.NotPanics(t, func() {
		cron.Start()
	})
	require.NotPanics(t, func() {
		cron.Stop()
	})

	require.NotPanics(t, func() {
		cron.Stop()
	})

	require.Panics(t, func() {
		cron.Start()
	})
}
