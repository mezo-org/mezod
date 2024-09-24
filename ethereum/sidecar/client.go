package sidecar

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdkmath "cosmossdk.io/math"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// Client connects to the Ethereum sidecar server and queries for the
// `AssetsLocked` events.
type Client struct {
	requestTimeout time.Duration
	connection     *grpc.ClientConn
}

func NewClient(
	serverAddress string,
	requestTimeout time.Duration,
	registry types.InterfaceRegistry,
) (*Client, error) {
	connection, err := grpc.Dial(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(registry).GRPCCodec()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to the Ethereum sidecar server: [%v]",
			err,
		)
	}

	return &Client{
		requestTimeout: requestTimeout,
		connection:     connection,
	}, nil
}

// GetAssetsLockedEvents returns confirmed AssetsLockedEvents with
// the sequence number falling within the half-open range, denoted by
// sequenceStart (included) and sequenceEnd (excluded). Nil can be
// passed for sequenceStart and sequenceEnd to indicate an unbounded
// edge of the range.
func (c *Client) GetAssetsLockedEvents(
	ctx context.Context,
	sequenceStart *sdkmath.Int,
	sequenceEnd *sdkmath.Int,
) ([]bridgetypes.AssetsLockedEvent, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	sidecarClient := pb.NewEthereumSidecarClient(c.connection)

	request := &pb.AssetsLockedEventsRequest{
		SequenceStart: *sequenceStart,
		SequenceEnd:   *sequenceEnd,
	}

	response, err := sidecarClient.AssetsLockedEvents(ctxWithTimeout, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsLocked events [%v]", err)
	}

	events := make([]bridgetypes.AssetsLockedEvent, len(response.Events))
	for i, event := range response.Events {
		events[i] = *event
	}

	// Make sure the events form a sequence strictly increasing by 1.
	if !bridgetypes.AssetsLockedEvents(events).IsStrictlyIncreasingSequence() {
		return nil, fmt.Errorf("events do not form a proper sequence")
	}

	return events, nil
}

func (c *Client) Close() error {
	return c.connection.Close()
}
