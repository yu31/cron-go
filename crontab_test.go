package gcron

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yu31/gcron/pkg/expr"
)

func Test_New(t *testing.T) {
	c := New()
	require.Equal(t, expr.Standard, c.parser)
	require.NotNil(t, c.mu)
	require.NotNil(t, c.tw)
	require.NotNil(t, c.jobs)
}

// TODO
func Test_Express_1(t *testing.T) {
	c := New()
	c.Start()
	c.Stop()
}
