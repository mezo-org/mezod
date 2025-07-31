package keeper

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestAssetsUnlockedSequenceTip(t *testing.T) {
	ctx, k := mockContext()
	qs := queryServer{k}

	sequenceTip := math.NewInt(15)
	k.setAssetsUnlockedSequenceTip(ctx, sequenceTip)

	expectedResponse := &bridgetypes.QueryAssetsUnlockedSequenceTipResponse{
		SequenceTip: sequenceTip,
	}

	actualResponse, err := qs.AssetsUnlockedSequenceTip(
		ctx, &bridgetypes.QueryAssetsUnlockedSequenceTipRequest{},
	)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, expectedResponse, actualResponse)
}

func TestAssetsUnlockedEvents(t *testing.T) {
	tests := map[string]struct {
		events           []bridgetypes.AssetsUnlockedEvent
		request          *bridgetypes.QueryAssetsUnlockedEventsRequest
		expectedResponse *bridgetypes.QueryAssetsUnlockedEventsResponse
		expectedError    error
	}{
		"sequence start is equal to end": {
			events: []bridgetypes.AssetsUnlockedEvent{},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(4),
				SequenceEnd:   math.NewInt(4),
			},
			expectedResponse: nil,
			expectedError: fmt.Errorf(
				"sequence start is not lower than sequence end",
			),
		},
		"sequence start is greater than end": {
			events: []bridgetypes.AssetsUnlockedEvent{},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(5),
				SequenceEnd:   math.NewInt(4),
			},
			expectedResponse: nil,
			expectedError: fmt.Errorf(
				"sequence start is not lower than sequence end",
			),
		},
		"sequence start is negative": {
			events: []bridgetypes.AssetsUnlockedEvent{},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(-3),
				SequenceEnd:   math.NewInt(4),
			},
			expectedResponse: nil,
			expectedError: fmt.Errorf(
				"invalid non-positive sequence range",
			),
		},
		"sequence start is zero": {
			events: []bridgetypes.AssetsUnlockedEvent{},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(0),
				SequenceEnd:   math.NewInt(2),
			},
			expectedResponse: nil,
			expectedError: fmt.Errorf(
				"invalid non-positive sequence range",
			),
		},
		"there are no events": {
			events: []bridgetypes.AssetsUnlockedEvent{},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(3),
				SequenceEnd:   math.NewInt(6),
			},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{},
			},
			expectedError: nil,
		},
		"sequence range is unbounded on both sides": {
			events: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{0x01, 0x11},
					Token:          "0x1111111111111111111111111111111111111111",
					Sender:         []byte{0x01, 0x11},
					Amount:         math.NewInt(100),
					Chain:          1,
					BlockTime:      time.Unix(1, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      []byte{0x02, 0x22},
					Token:          "0x2222222222222222222222222222222222222222",
					Sender:         []byte{0x02, 0x22},
					Amount:         math.NewInt(200),
					Chain:          2,
					BlockTime:      time.Unix(2, 0).UTC(),
				},
			},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{
					{
						UnlockSequence: math.NewInt(1),
						Recipient:      []byte{0x01, 0x11},
						Token:          "0x1111111111111111111111111111111111111111",
						Sender:         []byte{0x01, 0x11},
						Amount:         math.NewInt(100),
						Chain:          1,
						BlockTime:      time.Unix(1, 0).UTC(),
					},
					{
						UnlockSequence: math.NewInt(2),
						Recipient:      []byte{0x02, 0x22},
						Token:          "0x2222222222222222222222222222222222222222",
						Sender:         []byte{0x02, 0x22},
						Amount:         math.NewInt(200),
						Chain:          2,
						BlockTime:      time.Unix(2, 0).UTC(),
					},
				},
			},
			expectedError: nil,
		},
		"sequence range is unbounded on lower side": {
			events: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{0x01, 0x11},
					Token:          "0x1111111111111111111111111111111111111111",
					Sender:         []byte{0x01, 0x11},
					Amount:         math.NewInt(100),
					Chain:          1,
					BlockTime:      time.Unix(1, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      []byte{0x02, 0x22},
					Token:          "0x2222222222222222222222222222222222222222",
					Sender:         []byte{0x02, 0x22},
					Amount:         math.NewInt(200),
					Chain:          2,
					BlockTime:      time.Unix(2, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      []byte{0x03, 0x33},
					Token:          "0x3333333333333333333333333333333333333333",
					Sender:         []byte{0x03, 0x33},
					Amount:         math.NewInt(300),
					Chain:          1,
					BlockTime:      time.Unix(3, 0).UTC(),
				},
			},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceEnd: math.NewInt(3),
			},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{
					{
						UnlockSequence: math.NewInt(1),
						Recipient:      []byte{0x01, 0x11},
						Token:          "0x1111111111111111111111111111111111111111",
						Sender:         []byte{0x01, 0x11},
						Amount:         math.NewInt(100),
						Chain:          1,
						BlockTime:      time.Unix(1, 0).UTC(),
					},
					{
						UnlockSequence: math.NewInt(2),
						Recipient:      []byte{0x02, 0x22},
						Token:          "0x2222222222222222222222222222222222222222",
						Sender:         []byte{0x02, 0x22},
						Amount:         math.NewInt(200),
						Chain:          2,
						BlockTime:      time.Unix(2, 0).UTC(),
					},
				},
			},
			expectedError: nil,
		},
		"sequence range is unbounded on upper side": {
			events: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{0x01, 0x11},
					Token:          "0x1111111111111111111111111111111111111111",
					Sender:         []byte{0x01, 0x11},
					Amount:         math.NewInt(100),
					Chain:          1,
					BlockTime:      time.Unix(1, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      []byte{0x02, 0x22},
					Token:          "0x2222222222222222222222222222222222222222",
					Sender:         []byte{0x02, 0x22},
					Amount:         math.NewInt(200),
					Chain:          2,
					BlockTime:      time.Unix(2, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      []byte{0x03, 0x33},
					Token:          "0x3333333333333333333333333333333333333333",
					Sender:         []byte{0x03, 0x33},
					Amount:         math.NewInt(300),
					Chain:          1,
					BlockTime:      time.Unix(3, 0).UTC(),
				},
			},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(2),
			},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{
					{
						UnlockSequence: math.NewInt(2),
						Recipient:      []byte{0x02, 0x22},
						Token:          "0x2222222222222222222222222222222222222222",
						Sender:         []byte{0x02, 0x22},
						Amount:         math.NewInt(200),
						Chain:          2,
						BlockTime:      time.Unix(2, 0).UTC(),
					},
					{
						UnlockSequence: math.NewInt(3),
						Recipient:      []byte{0x03, 0x33},
						Token:          "0x3333333333333333333333333333333333333333",
						Sender:         []byte{0x03, 0x33},
						Amount:         math.NewInt(300),
						Chain:          1,
						BlockTime:      time.Unix(3, 0).UTC(),
					},
				},
			},
			expectedError: nil,
		},
		"sequence start is greater than tip": {
			events: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{0x01, 0x11},
					Token:          "0x1111111111111111111111111111111111111111",
					Sender:         []byte{0x01, 0x11},
					Amount:         math.NewInt(100),
					Chain:          1,
					BlockTime:      time.Unix(1, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      []byte{0x02, 0x22},
					Token:          "0x2222222222222222222222222222222222222222",
					Sender:         []byte{0x02, 0x22},
					Amount:         math.NewInt(200),
					Chain:          2,
					BlockTime:      time.Unix(2, 0).UTC(),
				},
			},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(3),
				SequenceEnd:   math.NewInt(6),
			},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{},
			},
			expectedError: nil,
		},
		"sequence end is greater than tip": {
			events: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{0x01, 0x11},
					Token:          "0x1111111111111111111111111111111111111111",
					Sender:         []byte{0x01, 0x11},
					Amount:         math.NewInt(100),
					Chain:          1,
					BlockTime:      time.Unix(1, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      []byte{0x02, 0x22},
					Token:          "0x2222222222222222222222222222222222222222",
					Sender:         []byte{0x02, 0x22},
					Amount:         math.NewInt(200),
					Chain:          2,
					BlockTime:      time.Unix(2, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      []byte{0x03, 0x33},
					Token:          "0x3333333333333333333333333333333333333333",
					Sender:         []byte{0x03, 0x33},
					Amount:         math.NewInt(300),
					Chain:          1,
					BlockTime:      time.Unix(3, 0).UTC(),
				},
			},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(2),
				SequenceEnd:   math.NewInt(5),
			},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{
					{
						UnlockSequence: math.NewInt(2),
						Recipient:      []byte{0x02, 0x22},
						Token:          "0x2222222222222222222222222222222222222222",
						Sender:         []byte{0x02, 0x22},
						Amount:         math.NewInt(200),
						Chain:          2,
						BlockTime:      time.Unix(2, 0).UTC(),
					},
					{
						UnlockSequence: math.NewInt(3),
						Recipient:      []byte{0x03, 0x33},
						Token:          "0x3333333333333333333333333333333333333333",
						Sender:         []byte{0x03, 0x33},
						Amount:         math.NewInt(300),
						Chain:          1,
						BlockTime:      time.Unix(3, 0).UTC(),
					},
				},
			},
			expectedError: nil,
		},
		"sequence range matches a single event": {
			events: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{0x01, 0x11},
					Token:          "0x1111111111111111111111111111111111111111",
					Sender:         []byte{0x01, 0x11},
					Amount:         math.NewInt(100),
					Chain:          1,
					BlockTime:      time.Unix(1, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      []byte{0x02, 0x22},
					Token:          "0x2222222222222222222222222222222222222222",
					Sender:         []byte{0x02, 0x22},
					Amount:         math.NewInt(200),
					Chain:          2,
					BlockTime:      time.Unix(2, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(3),
					Recipient:      []byte{0x03, 0x33},
					Token:          "0x3333333333333333333333333333333333333333",
					Sender:         []byte{0x03, 0x33},
					Amount:         math.NewInt(300),
					Chain:          1,
					BlockTime:      time.Unix(3, 0).UTC(),
				},
			},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(2),
				SequenceEnd:   math.NewInt(3),
			},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{
					{
						UnlockSequence: math.NewInt(2),
						Recipient:      []byte{0x02, 0x22},
						Token:          "0x2222222222222222222222222222222222222222",
						Sender:         []byte{0x02, 0x22},
						Amount:         math.NewInt(200),
						Chain:          2,
						BlockTime:      time.Unix(2, 0).UTC(),
					},
				},
			},
			expectedError: nil,
		},
		"sequence range matches all events": {
			events: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: math.NewInt(1),
					Recipient:      []byte{0x01, 0x11},
					Token:          "0x1111111111111111111111111111111111111111",
					Sender:         []byte{0x01, 0x11},
					Amount:         math.NewInt(100),
					Chain:          1,
					BlockTime:      time.Unix(1, 0).UTC(),
				},
				{
					UnlockSequence: math.NewInt(2),
					Recipient:      []byte{0x02, 0x22},
					Token:          "0x2222222222222222222222222222222222222222",
					Sender:         []byte{0x02, 0x22},
					Amount:         math.NewInt(200),
					Chain:          2,
					BlockTime:      time.Unix(2, 0).UTC(),
				},
			},
			request: &bridgetypes.QueryAssetsUnlockedEventsRequest{
				SequenceStart: math.NewInt(1),
				SequenceEnd:   math.NewInt(3),
			},
			expectedResponse: &bridgetypes.QueryAssetsUnlockedEventsResponse{
				Events: []bridgetypes.AssetsUnlockedEvent{
					{
						UnlockSequence: math.NewInt(1),
						Recipient:      []byte{0x01, 0x11},
						Token:          "0x1111111111111111111111111111111111111111",
						Sender:         []byte{0x01, 0x11},
						Amount:         math.NewInt(100),
						Chain:          1,
						BlockTime:      time.Unix(1, 0).UTC(),
					},
					{
						UnlockSequence: math.NewInt(2),
						Recipient:      []byte{0x02, 0x22},
						Token:          "0x2222222222222222222222222222222222222222",
						Sender:         []byte{0x02, 0x22},
						Amount:         math.NewInt(200),
						Chain:          2,
						BlockTime:      time.Unix(2, 0).UTC(),
					},
				},
			},
			expectedError: nil,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			ctx, k := mockContext()
			qs := queryServer{k}

			for _, event := range test.events {
				k.saveAssetsUnlocked(ctx, &event)
			}
			k.setAssetsUnlockedSequenceTip(ctx, math.NewInt(
				int64(len(test.events))),
			)

			actualResponse, err := qs.AssetsUnlockedEvents(
				ctx,
				test.request,
			)

			require.Equal(t, test.expectedResponse, actualResponse)
			require.Equal(t, test.expectedError, err)
		})
	}
}
