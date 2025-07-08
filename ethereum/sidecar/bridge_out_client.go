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
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

type BridgeOutClient struct {
	mutex          sync.Mutex
	requestTimeout time.Duration
	connection     *grpc.ClientConn
}

func NewBridgeOutClient(
	logger log.Logger,
	serverAddress string,
	requestTimeout time.Duration,
	registry types.InterfaceRegistry,
) (*BridgeOutClient, error) {
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

	c := &BridgeOutClient{
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

		_, err := c.GetAssetsUnlockedEntries(
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

func (c *BridgeOutClient) GetAssetsUnlockedEntries(
	ctx context.Context,
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
) ([]bridgetypes.AssetsUnlocked, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	sidecarClient := pb.NewEthereumSidecarBridgeOutServiceClient(c.connection)

	request := &pb.AssetsUnlockedEntriesRequest{
		SequenceStart: sequenceStart,
		SequenceEnd:   sequenceEnd,
	}

	response, err := sidecarClient.AssetsUnlockedEntries(ctxWithTimeout, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsUnlocked entries [%v]", err)
	}

	entries := make([]bridgetypes.AssetsUnlocked, len(response.Entries))
	for i, entry := range response.Entries {
		entries[i] = *entry
	}

	err = validateAssetsUnlockedEntries(sequenceStart, sequenceEnd, entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func validateAssetsUnlockedEntries(
	_ sdkmath.Int,
	_ sdkmath.Int,
	_ []bridgetypes.AssetsUnlocked,
) error {
	// TODO: Add validation.

	return nil
}
