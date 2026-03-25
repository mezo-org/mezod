package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithBatchEventFetch_FullRangeSuccess(t *testing.T) {
	var calls [][2]uint64
	fetchFunc := func(start, end uint64) ([]int, error) {
		calls = append(calls, [2]uint64{start, end})
		return []int{1, 2, 3}, nil
	}

	result, err := WithBatchEventFetch(fetchFunc, 10, 20, 600, 5)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, result)
	require.Equal(
		t,
		[][2]uint64{{10, 20}},
		calls,
		"should not batch when full-range fetch succeeds",
	)
}

func TestWithBatchEventFetch_FallbackToBatches(t *testing.T) {
	// Range [1..5], batchSize=2, batches: [1..3], [4..5]
	var calls [][2]uint64
	fetchFunc := func(start, end uint64) ([]int, error) {
		calls = append(calls, [2]uint64{start, end})

		// First call is the full-range attempt; force fallback.
		if len(calls) == 1 && start == 1 && end == 5 {
			return nil, fmt.Errorf("full-range failed")
		}

		switch {
		case start == 1 && end == 3:
			return []int{10, 11, 12}, nil
		case start == 4 && end == 5:
			return []int{13, 14}, nil
		default:
			return nil, fmt.Errorf("unexpected batch")
		}
	}

	start, end := uint64(1), uint64(5)
	result, err := WithBatchEventFetch(fetchFunc, start, end, 600, 2)

	require.NoError(t, err)
	require.Equal(t, []int{10, 11, 12, 13, 14}, result)

	require.Equal(t,
		[][2]uint64{
			{1, 5}, // initial full-range attempt
			{1, 3}, // batch 1
			{4, 5}, // batch 2
		},
		calls,
	)
}
