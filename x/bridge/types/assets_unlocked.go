package types

import (
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

type TargetChain uint8

const (
	TargetChainEthereum = iota
	TargetChainBitcoin
)

// AssetsUnlockedEvents is a slice of AssetsUnlockedEvent.
type AssetsUnlockedEvents []AssetsUnlockedEvent

// IsStrictlyIncreasingSequence returns true if the unlock sequence numbers of
// the events in the slice form a sequence strictly increasing by 1. Such a
// sequence guarantees that there are no gaps between the unlock sequence
// numbers of the events and that each event is unique by its unlock sequence
// number. Returns true for empty and single-element slices.
func (aue AssetsUnlockedEvents) IsStrictlyIncreasingSequence() bool {
	for i := 0; i < len(aue)-1; i++ {
		expectedNextSequence := aue[i].UnlockSequence.AddRaw(1)
		if !expectedNextSequence.Equal(aue[i+1].UnlockSequence) {
			return false
		}
	}

	return true
}

// IsValid returns true if the event is valid. An event is considered valid if
// its unlock sequence number is positive, its recipient is not an empty byte
// string, its token is a valid EVM hex address, its sender is a valid EVM hex
// address, the amount of unlocked assets is positive, the chain is Ethereum or
// Bitcoin and block time is positive.
func (aue AssetsUnlockedEvent) IsValid() bool {
	if aue.UnlockSequence.IsNil() || !aue.UnlockSequence.IsPositive() {
		return false
	}

	if len(aue.Recipient) == 0 {
		return false
	}

	if !evmtypes.IsHexAddress(aue.Token) {
		return false
	}

	if !evmtypes.IsHexAddress(aue.Sender) {
		return false
	}

	if aue.Amount.IsNil() || !aue.Amount.IsPositive() {
		return false
	}

	if aue.Chain != TargetChainEthereum &&
		aue.Chain != TargetChainBitcoin {
		return false
	}

	if aue.BlockTime == 0 {
		return false
	}

	return true
}

// IsValid returns true if all events in the slice are valid and their sequence
// numbers form a sequence strictly increasing by 1. See AssetsUnlockedEvent.IsValid
// and AssetsUnlockedEvents.IsStrictlyIncreasingSequence for more details.
func (aue AssetsUnlockedEvents) IsValid() bool {
	for _, event := range aue {
		if !event.IsValid() {
			return false
		}
	}

	return aue.IsStrictlyIncreasingSequence()
}
