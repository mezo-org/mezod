package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

func TestFetchABIEvents(t *testing.T) {
	onchainErr := fmt.Errorf("onchain failure")
	bridgeContract := newLocalBridgeContract()

	server := &Server{
		logger:            log.NewNopLogger(),
		events:            make([]bridgetypes.AssetsLockedEvent, 0),
		bridgeContract:    bridgeContract,
		batchSize:         3,
		requestsPerMinute: uint64(600),
	}

	tests := map[string]struct {
		startBlock, endBlock uint64
		onchainEvents        []*portal.MezoBridgeAssetsLocked
		onchainErrors        []error
		expectedEvents       []*portal.MezoBridgeAssetsLocked
		expectedErr          error
	}{
		"fetching whole range successful": {
			200,
			300,
			[]*portal.MezoBridgeAssetsLocked{
				{
					SequenceNumber: big.NewInt(100),
					Recipient:      common.HexToAddress("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(1000000),
				},
				{
					SequenceNumber: big.NewInt(101),
					Recipient:      common.HexToAddress("0xd728eB5aB3C743e0c2Cf5aFd4c81FEEC0f8f7300"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(2000000),
				},
			},
			nil,
			[]*portal.MezoBridgeAssetsLocked{
				{
					SequenceNumber: big.NewInt(100),
					Recipient:      common.HexToAddress("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(1000000),
				},
				{
					SequenceNumber: big.NewInt(101),
					Recipient:      common.HexToAddress("0xd728eB5aB3C743e0c2Cf5aFd4c81FEEC0f8f7300"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(2000000),
				},
			},
			nil,
		},
		"fetching whole range unsuccessful": {
			2,
			5,
			[]*portal.MezoBridgeAssetsLocked{
				{
					SequenceNumber: big.NewInt(100),
					Recipient:      common.HexToAddress("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(1000000),
				},
				{
					SequenceNumber: big.NewInt(101),
					Recipient:      common.HexToAddress("0xd728eB5aB3C743e0c2Cf5aFd4c81FEEC0f8f7300"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(2000000),
				},
			},
			[]error{onchainErr, nil},
			[]*portal.MezoBridgeAssetsLocked{
				{
					SequenceNumber: big.NewInt(100),
					Recipient:      common.HexToAddress("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(1000000),
				},
				{
					SequenceNumber: big.NewInt(101),
					Recipient:      common.HexToAddress("0xd728eB5aB3C743e0c2Cf5aFd4c81FEEC0f8f7300"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(2000000),
				},
			},
			nil,
		},
		"error when fetching": {
			200,
			300,
			[]*portal.MezoBridgeAssetsLocked{},
			[]error{onchainErr, onchainErr}, // return error twice
			nil,
			onchainErr,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			bridgeContract.SetErrors(test.onchainErrors)
			bridgeContract.SetEvents(test.onchainEvents)

			events, err := server.fetchABIEvents(test.startBlock, test.endBlock)

			require.ErrorIs(t, err, test.expectedErr)

			if !reflect.DeepEqual(test.expectedEvents, events) {
				t.Errorf(
					"unexpected events\n expected: %v\n actual:   %v",
					test.expectedEvents,
					events,
				)
			}
		})
	}
}

func TestFetchFinalizedEvents(t *testing.T) {
	bitcoinBridge := newLocalBridgeContract()

	server := &Server{
		logger:            log.NewNopLogger(),
		events:            []bridgetypes.AssetsLockedEvent{},
		bridgeContract:    bitcoinBridge,
		batchSize:         3,
		requestsPerMinute: uint64(600),
	}

	sdk.GetConfig().SetBech32PrefixForAccount(config.Bech32Prefix, "")

	tests := map[string]struct {
		startBlock, endBlock uint64
		// Events already stored in server
		serversEvents  []bridgetypes.AssetsLockedEvent
		onchainEvents  []*portal.MezoBridgeAssetsLocked
		onchainErrors  []error
		expectedEvents []bridgetypes.AssetsLockedEvent
		expectedErr    error
	}{
		"no new events": {
			2,
			5,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "cosmos1rct3qqhlh2drmy48e204s5n9hpctcwwkvgv2ds",
					Amount:    sdkmath.NewInt(10000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
			},
			[]*portal.MezoBridgeAssetsLocked{},
			nil,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "cosmos1rct3qqhlh2drmy48e204s5n9hpctcwwkvgv2ds",
					Amount:    sdkmath.NewInt(10000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
			},
			nil,
		},
		"gap between events": {
			2,
			5,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "mezo1jjzztg6j49nv3tz4mcsp74gr86g0zm0n7lhr9j",
					Amount:    sdkmath.NewInt(10000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
			},
			[]*portal.MezoBridgeAssetsLocked{
				{
					SequenceNumber: big.NewInt(3), // Greater by 2 from the already fetched events
					Recipient:      common.HexToAddress("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(1000000),
				},
				{
					SequenceNumber: big.NewInt(4),
					Recipient:      common.HexToAddress("0xd728eB5aB3C743e0c2Cf5aFd4c81FEEC0f8f7300"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(2000000),
				},
			},
			nil,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "mezo1jjzztg6j49nv3tz4mcsp74gr86g0zm0n7lhr9j",
					Amount:    sdkmath.NewInt(10000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
			},
			errSequenceGap,
		},
		"invalid events": {
			2,
			5,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "recipient1",
					Amount:    sdkmath.NewInt(10000),
					Token:     "token1",
				},
			},
			[]*portal.MezoBridgeAssetsLocked{
				{
					SequenceNumber: big.NewInt(2),
					Recipient:      common.HexToAddress("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(1000000),
				},
				{
					SequenceNumber: big.NewInt(4), // Greater by 2 from the previous event
					Recipient:      common.HexToAddress("0xd728eB5aB3C743e0c2Cf5aFd4c81FEEC0f8f7300"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(2000000),
				},
			},
			nil,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "recipient1",
					Amount:    sdkmath.NewInt(10000),
					Token:     "token1",
				},
			},
			errInvalidEvents,
		},
		"fetched events valid": {
			2,
			5,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "mezo1jjzztg6j49nv3tz4mcsp74gr86g0zm0n7lhr9j",
					Amount:    sdkmath.NewInt(10000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
			},
			[]*portal.MezoBridgeAssetsLocked{
				{
					SequenceNumber: big.NewInt(2),
					Recipient:      common.HexToAddress("0x948425A352A966c8AC55De201F55033E90F16Df3"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(1000000),
				},
				{
					SequenceNumber: big.NewInt(3),
					Recipient:      common.HexToAddress("0xd728eB5aB3C743e0c2Cf5aFd4c81FEEC0f8f7300"),
					Token:          common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
					Amount:         big.NewInt(2000000),
				},
			},
			nil,
			[]bridgetypes.AssetsLockedEvent{
				{
					Sequence:  sdkmath.NewInt(1),
					Recipient: "mezo1jjzztg6j49nv3tz4mcsp74gr86g0zm0n7lhr9j",
					Amount:    sdkmath.NewInt(10000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
				{
					Sequence:  sdkmath.NewInt(2),
					Recipient: "mezo1jjzztg6j49nv3tz4mcsp74gr86g0zm0n7lhr9j",
					Amount:    sdkmath.NewInt(1000000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
				{
					Sequence:  sdkmath.NewInt(3),
					Recipient: "mezo16u5wkk4ncap7psk0tt75eq07as8c7ucqua575k",
					Amount:    sdkmath.NewInt(2000000),
					Token:     "0x3A128b915bee3645396d43Fe7A13A59a66C427d6",
				},
			},
			nil,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			bitcoinBridge.SetErrors(test.onchainErrors)
			bitcoinBridge.SetEvents(test.onchainEvents)

			server.events = test.serversEvents

			err := server.fetchFinalizedEvents(test.startBlock, test.endBlock)
			require.ErrorIs(t, err, test.expectedErr)

			if !reflect.DeepEqual(test.expectedEvents, server.events) {
				t.Errorf(
					"unexpected events\n expected: %v\n actual:   %v",
					test.expectedEvents,
					server.events,
				)
			}
		})
	}
}

func TestAssetsLockedEvents(t *testing.T) {
	server := &Server{
		events: []bridgetypes.AssetsLockedEvent{
			{Sequence: sdkmath.NewIntFromBigInt(big.NewInt(1)), Recipient: "recipient1", Amount: sdkmath.NewIntFromBigInt(big.NewInt(100)), Token: "token1"},
			{Sequence: sdkmath.NewIntFromBigInt(big.NewInt(2)), Recipient: "recipient2", Amount: sdkmath.NewIntFromBigInt(big.NewInt(200)), Token: "token2"},
		},
	}

	req := &pb.AssetsLockedEventsRequest{
		SequenceStart: sdkmath.NewInt(1),
		SequenceEnd:   sdkmath.NewInt(3),
	}

	resp, err := server.AssetsLockedEvents(context.Background(), req)

	require.NoError(t, err)
	assert.Len(t, resp.Events, 2)
	assert.Equal(t, "recipient1", resp.Events[0].Recipient)
	assert.Equal(t, "recipient2", resp.Events[1].Recipient)
	assert.Equal(t, int64(100), resp.Events[0].Amount.Int64())
	assert.Equal(t, int64(200), resp.Events[1].Amount.Int64())
	assert.Equal(t, int64(1), resp.Events[0].Sequence.Int64())
	assert.Equal(t, int64(2), resp.Events[1].Sequence.Int64())
	assert.Equal(t, "token1", resp.Events[0].Token)
	assert.Equal(t, "token2", resp.Events[1].Token)
}

func TestFetchRecentAssetsUnlockedEvents(t *testing.T) {
	// Mock function always returning 10000 seconds since Unix epoch as the
	// current time.
	mockTimeFunc := func() time.Time {
		return time.Unix(10_000, 0)
	}

	// Look back period of 1000 seconds. In combination with the mock time
	// function above, it determines that the cut-off time for AssetsUnlocked
	// events will be a block time of 9000 (10000 - 1000). If we encounter
	// an event older than that we will stop fetching events immediately
	// without including that event.
	lookBackPeriod := 1_000 * time.Second

	// Events will be fetched in batches of 3.
	batchSize := 3

	// Events used in tests only need have unlock sequence and block time set.
	// Others fields omitted.
	tests := map[string]struct {
		mezoEvents     []bridgetypes.AssetsUnlockedEvent
		expectedEvents []bridgetypes.AssetsUnlockedEvent
	}{
		"no events": {
			mezoEvents:     []bridgetypes.AssetsUnlockedEvent{},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{},
		},
		"single event, within look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9000, // on the border of look back period
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9000,
				},
			},
		},
		"number of events one smaller than batch size, within look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9900,
				},
			},
		},
		"number of events equal to batch size, within look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9900,
				},
			},
		},
		"number of events exceeds batch size, within look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9900,
				},
			},
		},
		"number of events equal two batch sizes, within look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9400,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9500,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9400,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9500,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9900,
				},
			},
		},
		"number of events exceeds two batch sizes, within look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9300,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9400,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9500,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(7),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      9300,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9400,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9500,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(7),
					BlockTime:      9900,
				},
			},
		},
		"single event, outside look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      8999, // just outside look back period
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{},
		},
		"event from first batch outside look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      8800,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      8900, // below cut-off
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9900,
				},
			},
		},
		"event from second batch outside look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      8800,
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      8900, // below cut-off
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9900,
				},
			},
		},
		"event from third batch outside look-back period": {
			mezoEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(1),
					BlockTime:      8800, // first from third batch outside cut-off
				},
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9400, // oldest from second batch
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9500,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9700, // oldest from first batch
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(7),
					BlockTime:      9900,
				},
			},
			expectedEvents: []bridgetypes.AssetsUnlockedEvent{
				{
					UnlockSequence: sdkmath.NewInt(2),
					BlockTime:      9400,
				},
				{
					UnlockSequence: sdkmath.NewInt(3),
					BlockTime:      9500,
				},
				{
					UnlockSequence: sdkmath.NewInt(4),
					BlockTime:      9600,
				},
				{
					UnlockSequence: sdkmath.NewInt(5),
					BlockTime:      9700,
				},
				{
					UnlockSequence: sdkmath.NewInt(6),
					BlockTime:      9800,
				},
				{
					UnlockSequence: sdkmath.NewInt(7),
					BlockTime:      9900,
				},
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			ctx, cancelCtx := context.WithCancel(context.Background())
			defer cancelCtx()

			assetsUnlockedEndpoint := newLocalAssetsUnlockedEndpoint()
			assetsUnlockedEndpoint.SetAssetsUnlockedEvents(test.mezoEvents)

			server := Server{
				assetsUnlockedEndpoint:       assetsUnlockedEndpoint,
				assetsUnlockedLookBackPeriod: lookBackPeriod,
				assetsUnlockedBatchSize:      batchSize,
				attestationQueue:             make([]bridgetypes.AssetsUnlockedEvent, 0),
				timeFunc:                     mockTimeFunc,
			}

			actualEvents, err := server.fetchRecentAssetsUnlockedEvents(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(test.expectedEvents, actualEvents) {
				t.Errorf(
					"unexpected events\n expected: %v\n actual:   %v",
					test.expectedEvents,
					actualEvents,
				)
			}
		})
	}
}
