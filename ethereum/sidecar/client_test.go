package sidecar

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmdcfg "github.com/mezo-org/mezod/cmd/config"

	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

//nolint:gosec
const token = "0x517f2982701695D4E52f1ECFBEf3ba31Df470161"

func TestClient_ValidateAssetsLockedEvents(t *testing.T) {
	config := sdk.GetConfig()
	cmdcfg.SetBech32Prefixes(config)

	tests := map[string]struct {
		sequenceStart sdkmath.Int
		sequenceEnd   sdkmath.Int
		events        []bridgetypes.AssetsLockedEvent
		expectedErr   error
	}{
		"when the events slice is empty": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.NewInt(5),
			events:        []bridgetypes.AssetsLockedEvent{},
			expectedErr:   nil,
		},
		"when one of the events is invalid": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.NewInt(5),
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(2),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(-1000000), // Negative amount.
					Token:     token,
				},
			},
			expectedErr: ErrInvalidEventsSequence,
		},
		"when the sequence of events is not strictly increasing": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.NewInt(6),
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
				{
					Sequence:  sdkmath.NewInt(5),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: ErrInvalidEventsSequence,
		},
		"when the first event's sequence is lower than requested sequence start": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.NewInt(5),
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: ErrRequestedBoundariesViolated,
		},
		"when the last event's sequence is equal to requested sequence end": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.NewInt(5),
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(5),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: ErrRequestedBoundariesViolated,
		},
		"when the last event's sequence is greater than requested sequence end": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.NewInt(5),
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(6),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: ErrRequestedBoundariesViolated,
		},
		"when the requested sequence start is nil and events are valid": {
			sequenceStart: sdkmath.Int{},
			sequenceEnd:   sdkmath.NewInt(5),
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(2),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: nil,
		},
		"when the requested sequence end is nil and events are valid": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.Int{},
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
				{
					Sequence:  sdkmath.NewInt(4),
					Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: nil,
		},
		"when the requested sequence start and end are nil and events are valid": {
			sequenceStart: sdkmath.Int{},
			sequenceEnd:   sdkmath.Int{},
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(2),
					Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: nil,
		},
		"when the requested sequence start and end are not nil and events are valid": {
			sequenceStart: sdkmath.NewInt(2),
			sequenceEnd:   sdkmath.NewInt(5),
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(2),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "mezo1pd4u0j77ydrsrv8z8m9854rsmg3jh45kjqwg54",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
				{
					Sequence:  sdkmath.NewInt(4),
					Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
					Amount:    sdkmath.NewInt(1000000),
					Token:     token,
				},
			},
			expectedErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateAssetsLockedEvents(
				test.sequenceStart,
				test.sequenceEnd,
				test.events,
			)

			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}
