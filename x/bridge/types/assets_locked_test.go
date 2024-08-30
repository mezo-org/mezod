package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestAssetsLockedEvents_IsStrictlyIncreasingSequence(t *testing.T) {
	tests := map[string]struct {
		events      AssetsLockedEvents
		expectedRes bool
	}{
		"empty": {
			events:      AssetsLockedEvents{},
			expectedRes: true,
		},
		"nil": {
			events:      nil,
			expectedRes: true,
		},
		"single element": {
			events:      AssetsLockedEvents{{Sequence: math.NewInt(1)}},
			expectedRes: true,
		},
		"strictly increasing from zero": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(0)},
				{Sequence: math.NewInt(1)},
				{Sequence: math.NewInt(2)},
			},
			expectedRes: true,
		},
		"strictly increasing from positive": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1)},
				{Sequence: math.NewInt(2)},
				{Sequence: math.NewInt(3)},
			},
			expectedRes: true,
		},
		"strictly decreasing": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(3)},
				{Sequence: math.NewInt(2)},
				{Sequence: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"increasing (non-strictly)": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1)},
				{Sequence: math.NewInt(1)},
				{Sequence: math.NewInt(2)},
				{Sequence: math.NewInt(3)},
			},
			expectedRes: false,
		},
		"decreasing (non-strictly)": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(3)},
				{Sequence: math.NewInt(3)},
				{Sequence: math.NewInt(2)},
				{Sequence: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"gap": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1)},
				{Sequence: math.NewInt(2)},
				{Sequence: math.NewInt(4)},
			},
			expectedRes: false,
		},
		"duplicate": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1)},
				{Sequence: math.NewInt(2)},
				{Sequence: math.NewInt(3)},
				{Sequence: math.NewInt(1)},
			},
			expectedRes: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(
				t,
				test.expectedRes,
				test.events.IsStrictlyIncreasingSequence(),
			)
		})
	}
}
