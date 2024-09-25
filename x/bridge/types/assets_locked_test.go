package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/cmd/config"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

const recipient = "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp"

func TestAssetsLockedEvents_IsValid(t *testing.T) {
	// Set bech32 prefixes to make the recipient address validation possible.
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	tests := map[string]struct {
		events      AssetsLockedEvents
		expectedRes bool
	}{
		"contains invalid element - nil sequence": {
			events:      AssetsLockedEvents{{Sequence: math.NewIntFromBigIntMut(nil), Recipient: recipient, Amount: math.NewInt(1)}},
			expectedRes: false,
		},
		"contains invalid element - non-positive sequence": {
			events:      AssetsLockedEvents{{Sequence: math.NewInt(0), Recipient: recipient, Amount: math.NewInt(1)}},
			expectedRes: false,
		},
		"contains invalid element - empty recipient address": {
			events:      AssetsLockedEvents{{Sequence: math.NewInt(1), Recipient: "", Amount: math.NewInt(1)}},
			expectedRes: false,
		},
		"contains invalid element - invalid recipient address": {
			events:      AssetsLockedEvents{{Sequence: math.NewInt(1), Recipient: "corrupted", Amount: math.NewInt(1)}},
			expectedRes: false,
		},
		"contains invalid element - nil amount": {
			events:      AssetsLockedEvents{{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewIntFromBigIntMut(nil)}},
			expectedRes: false,
		},
		"contains invalid element - non-positive amount": {
			events:      AssetsLockedEvents{{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(0)}},
			expectedRes: false,
		},
		"empty": {
			events:      AssetsLockedEvents{},
			expectedRes: true,
		},
		"nil": {
			events:      nil,
			expectedRes: true,
		},
		"single element": {
			events:      AssetsLockedEvents{{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)}},
			expectedRes: true,
		},
		"strictly increasing from zero": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(0), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"strictly increasing from positive": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(3), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: true,
		},
		"strictly decreasing": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(3), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"increasing (non-strictly)": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(3), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"decreasing (non-strictly)": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(3), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(3), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"gap": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(4), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"duplicate": {
			events: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(3), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
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

func TestAssetsLockedEvents_Equal(t *testing.T) {
	tests := map[string]struct {
		events1     AssetsLockedEvents
		events2     AssetsLockedEvents
		expectedRes bool
	}{
		"both nil": {
			events1:     nil,
			events2:     nil,
			expectedRes: true,
		},
		"both empty": {
			events1:     AssetsLockedEvents{},
			events2:     AssetsLockedEvents{},
			expectedRes: true,
		},
		"first nil second empty": {
			events1:     nil,
			events2:     AssetsLockedEvents{},
			expectedRes: true,
		},
		"first empty second nil": {
			events1:     AssetsLockedEvents{},
			events2:     nil,
			expectedRes: true,
		},
		"first nil second non-empty": {
			events1: nil,
			events2: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"first empty second non-empty": {
			events1: AssetsLockedEvents{},
			events2: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"first non-empty second nil": {
			events1: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			events2:     nil,
			expectedRes: false,
		},
		"first non-empty second empty": {
			events1: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			events2:     AssetsLockedEvents{},
			expectedRes: false,
		},
		"both non-empty - different sizes": {
			events1: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			events2: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"both non-empty - different elements - sequence": {
			events1: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			events2: AssetsLockedEvents{
				{Sequence: math.NewInt(2), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"both non-empty - different elements - recipient": {
			events1: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			events2: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", Amount: math.NewInt(1)},
			},
			expectedRes: false,
		},
		"both non-empty - different elements - amount": {
			events1: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			events2: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(2)},
			},
			expectedRes: false,
		},
		"both non-empty - same size and elements": {
			events1: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			events2: AssetsLockedEvents{
				{Sequence: math.NewInt(1), Recipient: recipient, Amount: math.NewInt(1)},
			},
			expectedRes: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(
				t,
				test.expectedRes,
				test.events1.Equal(test.events2),
			)
		})
	}
}

