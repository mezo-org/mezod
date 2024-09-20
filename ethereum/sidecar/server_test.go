package sidecar

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"

	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestServer_AssetsLockedEvents(t *testing.T) {
	tests := map[string]struct {
		events           []bridgetypes.AssetsLockedEvent
		request          *pb.AssetsLockedEventsRequest
		expectedResponse *pb.AssetsLockedEventsResponse
		expectedErr      error
	}{
		"when sequence start is equal to end": {
			events: []bridgetypes.AssetsLockedEvent{},
			request: &pb.AssetsLockedEventsRequest{
				SequenceStart: sdkmath.NewInt(2),
				SequenceEnd:   sdkmath.NewInt(2),
			},
			expectedResponse: nil,
			expectedErr:      ErrSequenceStartNotLower,
		},
		"when sequence start is greater than end": {
			events: []bridgetypes.AssetsLockedEvent{},
			request: &pb.AssetsLockedEventsRequest{
				SequenceStart: sdkmath.NewInt(3),
				SequenceEnd:   sdkmath.NewInt(2),
			},
			expectedResponse: nil,
			expectedErr:      ErrSequenceStartNotLower,
		},
		"when sequence start unbounded": {
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "aaa",
					Amount:    sdkmath.NewInt(111),
				},
			},
			request: &pb.AssetsLockedEventsRequest{
				SequenceStart: sdkmath.Int{},
				SequenceEnd:   sdkmath.NewInt(4),
			},
			expectedResponse: &pb.AssetsLockedEventsResponse{
				Events: []*bridgetypes.AssetsLockedEvent{
					{
						Sequence:  sdkmath.NewInt(1),
						Recipient: "aaa",
						Amount:    sdkmath.NewInt(111),
					},
				},
			},
			expectedErr: nil,
		},
		"when sequence end unbounded": {
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "ccc",
					Amount:    sdkmath.NewInt(333),
				},
			},
			request: &pb.AssetsLockedEventsRequest{
				SequenceStart: sdkmath.NewInt(3),
				SequenceEnd:   sdkmath.Int{},
			},
			expectedResponse: &pb.AssetsLockedEventsResponse{
				Events: []*bridgetypes.AssetsLockedEvent{
					{
						Sequence:  sdkmath.NewInt(3),
						Recipient: "ccc",
						Amount:    sdkmath.NewInt(333),
					},
				},
			},
			expectedErr: nil,
		},
		"when sequence start and end unbounded": {
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "aaa",
					Amount:    sdkmath.NewInt(111),
				},
				{
					Sequence:  sdkmath.NewInt(2),
					Recipient: "bbb",
					Amount:    sdkmath.NewInt(222),
				},
			},
			request: &pb.AssetsLockedEventsRequest{
				SequenceStart: sdkmath.Int{},
				SequenceEnd:   sdkmath.Int{},
			},
			expectedResponse: &pb.AssetsLockedEventsResponse{
				Events: []*bridgetypes.AssetsLockedEvent{
					{
						Sequence:  sdkmath.NewInt(1),
						Recipient: "aaa",
						Amount:    sdkmath.NewInt(111),
					},
					{
						Sequence:  sdkmath.NewInt(2),
						Recipient: "bbb",
						Amount:    sdkmath.NewInt(222),
					},
				},
			},
			expectedErr: nil,
		},
		"when sequence start and end bounded": {
			events: []bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "aaa",
					Amount:    sdkmath.NewInt(111),
				},
				{
					Sequence:  sdkmath.NewInt(2),
					Recipient: "bbb",
					Amount:    sdkmath.NewInt(222),
				},
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "ccc",
					Amount:    sdkmath.NewInt(333),
				},
				{
					Sequence:  sdkmath.NewInt(4),
					Recipient: "ddd",
					Amount:    sdkmath.NewInt(444),
				},
			},
			request: &pb.AssetsLockedEventsRequest{
				SequenceStart: sdkmath.NewInt(2),
				SequenceEnd:   sdkmath.NewInt(4),
			},
			expectedResponse: &pb.AssetsLockedEventsResponse{
				Events: []*bridgetypes.AssetsLockedEvent{
					{
						Sequence:  sdkmath.NewInt(2),
						Recipient: "bbb",
						Amount:    sdkmath.NewInt(222),
					},
					{
						Sequence:  sdkmath.NewInt(3),
						Recipient: "ccc",
						Amount:    sdkmath.NewInt(333),
					},
				},
			},
			expectedErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			server := &Server{
				events: test.events,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			response, err := server.AssetsLockedEvents(ctx, test.request)

			require.ErrorIs(t, err, test.expectedErr)
			require.Equal(t, test.expectedResponse, response)
		})
	}
}
