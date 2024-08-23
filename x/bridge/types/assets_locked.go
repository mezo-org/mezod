package types

import "golang.org/x/exp/slices"

// AssetsLockedEventsCmp is a sequence-based comparator for AssetsLockedEvent
// that defines the natural order for a slice of AssetsLockedEvent.
func AssetsLockedEventsCmp(a, b AssetsLockedEvent) int {
	return a.Sequence.BigInt().Cmp(b.Sequence.BigInt())
}

// AssetsLockedEvents is a slice of AssetsLockedEvent.
type AssetsLockedEvents []AssetsLockedEvent

// IsStrictlyIncreasingSequence returns true if the sequence numbers of the
// events in the slice form a sequence strictly increasing by 1. Such a sequence
// guarantees that there are no gaps between the sequence numbers of the events
// and that each event is unique by its sequence number.
func (ale AssetsLockedEvents) IsStrictlyIncreasingSequence() bool {
	if !slices.IsSortedFunc(ale, AssetsLockedEventsCmp) {
		return false
	}

	for i := 0; i < len(ale)-1; i++ {
		expectedNextSequence := ale[i].Sequence.AddRaw(1)
		if !expectedNextSequence.Equal(ale[i+1].Sequence) {
			return false
		}
	}

	return true
}
