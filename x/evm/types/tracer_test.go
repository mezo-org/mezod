package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNoOpTracer(t *testing.T) {
	tracer, _ := NewNoopTracer()
	require.Equal(t, &NoOpTracer{}, tracer)
}
