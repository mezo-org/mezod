package sidecar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// BridgeOutGrpcClient enables gRPC communication with mezod validator node needed
// for bridge-out process.
type BridgeOutGrpcClient struct {
	mutex          sync.Mutex
	requestTimeout time.Duration
	connection     *grpc.ClientConn
}

func NewBridgeOutGrpcClient(
	logger log.Logger,
	serverAddress string,
	requestTimeout time.Duration,
	registry types.InterfaceRegistry,
) (*BridgeOutGrpcClient, error) {
	connection, err := grpc.NewClient(
		serverAddress,
		// TODO: Consider using TLS protocol so that the Mezo node and Ethereum
		//       sidecar can be run on separate servers.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(registry).GRPCCodec()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create grpc bridge-out client for Ethereum sidecar [%w]",
			err,
		)
	}

	c := &BridgeOutGrpcClient{
		requestTimeout: requestTimeout,
		connection:     connection,
	}

	go func() {
		// Test the connection to the Ethereum sidecar bridge-out server by
		// verifying we can successfully execute `GetAssetsUnlockedEntries`.
		ctxWithTimeout, cancel := context.WithTimeout(
			context.Background(),
			requestTimeout,
		)
		defer cancel()

		_, err := c.GetAssetsUnlockedEvents(
			ctxWithTimeout,
			sdkmath.NewInt(1),
			sdkmath.NewInt(2),
		)
		if err != nil {
			logger.Error(
				"ethereum sidecar bridge-out connection test failed; possible " +
					"problem with sidecar configuration or connectivity",
			)
		} else {
			logger.Info(
				"ethereum bridge-out sidecar connection test completed " +
					"successfully",
			)
		}
	}()

	return c, nil
}

// GetAssetsUnlockedEvents gets the AssetsUnlocked events from the Mezo chain.
// The requested range of events is inclusive on the lower side and exclusive
// on the upper side.
func (c *BridgeOutGrpcClient) GetAssetsUnlockedEvents(
	ctx context.Context,
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
) ([]bridgetypes.AssetsUnlockedEvent, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	queryClient := bridgetypes.NewQueryClient(c.connection)

	request := &bridgetypes.QueryAssetsUnlockedEventsRequest{
		SequenceStart: sequenceStart,
		SequenceEnd:   sequenceEnd,
	}

	response, err := queryClient.AssetsUnlockedEvents(ctxWithTimeout, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsUnlocked events: [%w]", err)
	}

	events := make([]bridgetypes.AssetsUnlockedEvent, len(response.Events))
	copy(events, response.Events)

	err = validateAssetsUnlockedEntries(sequenceStart, sequenceEnd, events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetAssetsUnlockedSequenceTip gets the assets unlocked sequence tip from the
// Mezo chain. The returned sequence tip is equal to the number of AssetsUnlocked
// events made so far. It is also equal to the value of the unlock sequence
// in the newest AssetsUnlocked event.
func (c *BridgeOutGrpcClient) GetAssetsUnlockedSequenceTip(
	ctx context.Context,
) (sdkmath.Int, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	queryClient := bridgetypes.NewQueryClient(c.connection)

	request := &bridgetypes.QueryAssetsUnlockedSequenceTipRequest{}

	response, err := queryClient.AssetsUnlockedSequenceTip(
		ctxWithTimeout,
		request,
	)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf(
			"failed to get assets unlocked sequence tip [%v]",
			err,
		)
	}

	tip := response.SequenceTip
	if tip.IsNil() || tip.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf(
			"assets unlocked sequence tip is nil or negative",
		)
	}

	return tip, nil
}

func validateAssetsUnlockedEntries(
	_ sdkmath.Int,
	_ sdkmath.Int,
	_ []bridgetypes.AssetsUnlockedEvent,
) error {
	// TODO: Add validation.

	return nil
}
