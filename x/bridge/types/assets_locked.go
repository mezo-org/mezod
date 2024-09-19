package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"slices"
)

// IsValid returns true if the event is valid. An event is considered valid if
// its sequence number is positive, its recipient address is a valid Bech32
// account address, and the amount of locked assets is positive.
func (ale AssetsLockedEvent) IsValid() bool {
	sequenceValid := !ale.Sequence.IsNil() && ale.Sequence.IsPositive()
	if !sequenceValid {
		return false
	}

	if _, err := sdk.AccAddressFromBech32(ale.Recipient); err != nil {
		return false
	}

	amountValid := !ale.Amount.IsNil() && ale.Amount.IsPositive()
	if !amountValid {
		return false
	}

	return true
}

// Equal returns true if this AssetsLockedEvents is equal to the other event.
// Two events are considered equal if their sequence numbers, recipient addresses,
// and amounts of locked assets are equal.
func (ale AssetsLockedEvent) Equal(other AssetsLockedEvent) bool {
	return ale.Sequence.Equal(other.Sequence) &&
		ale.Recipient == other.Recipient &&
		ale.Amount.Equal(other.Amount)
}

// AssetsLockedEvents is a slice of AssetsLockedEvent.
type AssetsLockedEvents []AssetsLockedEvent

// IsStrictlyIncreasingSequence returns true if the sequence numbers of the
// events in the slice form a sequence strictly increasing by 1. Such a sequence
// guarantees that there are no gaps between the sequence numbers of the events
// and that each event is unique by its sequence number. Returns true for
// empty and single-element slices.
func (ale AssetsLockedEvents) IsStrictlyIncreasingSequence() bool {
	for i := 0; i < len(ale)-1; i++ {
		expectedNextSequence := ale[i].Sequence.AddRaw(1)
		if !expectedNextSequence.Equal(ale[i+1].Sequence) {
			return false
		}
	}

	return true
}

// IsValid returns true if all events in the slice are valid and their sequence
// numbers form a sequence strictly increasing by 1. See AssetsLockedEvent.IsValid
// and AssetsLockedEvents.IsStrictlyIncreasingSequence for more details.
func (ale AssetsLockedEvents) IsValid() bool {
	for _, event := range ale {
		if !event.IsValid() {
			return false
		}
	}

	return ale.IsStrictlyIncreasingSequence()
}

// Equal returns true if this AssetsLockedEvent slice is equal to the other slice.
// Two slices are considered equal if they have the same length and all their
// elements are equal. See AssetsLockedEvent.Equal for more details.
func (ale AssetsLockedEvents) Equal(other AssetsLockedEvents) bool {
	return slices.EqualFunc(ale, other, func(a, b AssetsLockedEvent) bool {
		return a.Equal(b)
	})
}