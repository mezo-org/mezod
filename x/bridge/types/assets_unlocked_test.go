package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestAssetsUnlockedEvents_IsValid(t *testing.T) {
	recipient := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
	sender := "0x1650b4A3B2f508d13C3caC9105DD76c87fA176F8"
	token := "0xB9c3840bf0Dc03DfBCa72837040594A894444a29" //nolint:gosec

	tests := map[string]struct {
		events      AssetsUnlockedEvents
		expectedRes bool
	}{
		"contains invalid element - nil sequence": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewIntFromBigIntMut(nil),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - non-positive sequence": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(0),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - empty recipient address": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{},
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - empty token": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          "",
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - invalid token": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          "corrupted",
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - empty sender": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         "",
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - invalid sender": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         "corrupted",
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - nil amount": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewIntFromBigIntMut(nil),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - non-positive amount": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.ZeroInt(),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - wrong chain": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          2,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"contains invalid element - zero time": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      0,
				},
			},
			expectedRes: false,
		},
		"empty": {
			events:      AssetsUnlockedEvents{},
			expectedRes: true,
		},
		"nil": {
			events:      nil,
			expectedRes: true,
		},
		"single element": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: true,
		},
		"strictly increasing from zero": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(0),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"strictly increasing from positive": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: true,
		},
		"strictly decreasing": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"increasing (non-strictly)": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"decreasing (non-strictly)": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"gap": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(4),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
		"duplicate": {
			events: AssetsUnlockedEvents{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      recipient,
					Token:          token,
					Sender:         sender,
					Amount:         math.NewInt(1),
					Chain:          TargetChainBitcoin,
					BlockTime:      1,
				},
			},
			expectedRes: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(
				t,
				test.expectedRes,
				test.events.IsValid(),
			)
		})
	}
}
