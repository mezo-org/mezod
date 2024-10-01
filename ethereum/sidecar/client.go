package sidecar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

var (
	// ErrInvalidEventsSequence is the error returned when the returned sequence
	// of events is invalid.
	ErrInvalidEventsSequence = fmt.Errorf(
		"server returned invalid events sequence",
	)

	// ErrRequestedBoundariesViolated is the error returned when the sequence of
	// events returned from the server violate the requested boundaries.
	ErrRequestedBoundariesViolated = fmt.Errorf(
		"server returned events sequence violating requested boundaries",
	)
)

// Client connects to the Ethereum sidecar server and queries for the
// `AssetsLocked` events.
type Client struct {
	requestTimeout time.Duration
	connection     *grpc.ClientConn
	mutex          sync.Mutex
}

func NewClient(
	serverAddress string,
	requestTimeout time.Duration,
	registry types.InterfaceRegistry,
	logger log.Logger,
) (*Client, error) {
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
			"failed to create grpc connection to Ethereum sidecar server: [%w]",
			err,
		)
	}

	c := &Client{
		requestTimeout: requestTimeout,
		connection:     connection,
	}

	go func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		state := c.connection.GetState()
		if state == connectivity.Idle {
			ctx, cancelCtx := context.WithTimeout(
				context.Background(),
				requestTimeout,
			)
			defer cancelCtx()

			c.connection.Connect()

			if c.connection.WaitForStateChange(ctx, state) {
				logger.Info(
					"ethereum sidecar connection test completed successfully",
				)
			} else {
				logger.Error(
					"ethereum sidecar connection test failed; possible " +
						"problem with sidecar configuration or connectivity",
				)
			}
		}
	}()

	return c, nil
}

// GetAssetsLockedEvents returns confirmed AssetsLockedEvents with
// the sequence number falling within the half-open range, denoted by
// sequenceStart (included) and sequenceEnd (excluded). Nil can be
// passed for sequenceStart and sequenceEnd to indicate an unbounded
// edge of the range.
func (c *Client) GetAssetsLockedEvents(
	ctx context.Context,
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
) ([]bridgetypes.AssetsLockedEvent, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	sidecarClient := pb.NewEthereumSidecarClient(c.connection)

	request := &pb.AssetsLockedEventsRequest{
		SequenceStart: sequenceStart,
		SequenceEnd:   sequenceEnd,
	}

	response, err := sidecarClient.AssetsLockedEvents(ctxWithTimeout, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsLocked events [%v]", err)
	}

	events := make([]bridgetypes.AssetsLockedEvent, len(response.Events))
	for i, event := range response.Events {
		events[i] = *event
	}

	err = validateAssetsLockedEvents(sequenceStart, sequenceEnd, events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// validateAssetsLockedEvents validates the AssetsLocked events returned from
// the Ethereum sidecar server.
func validateAssetsLockedEvents(
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
	events []bridgetypes.AssetsLockedEvent,
) error {
	if len(events) == 0 {
		return nil
	}

	if !bridgetypes.AssetsLockedEvents(events).IsValid() {
		return ErrInvalidEventsSequence
	}

	if (!sequenceStart.IsNil() && events[0].Sequence.LT(sequenceStart)) ||
		(!sequenceEnd.IsNil() && events[len(events)-1].Sequence.GTE(sequenceEnd)) {
		return ErrRequestedBoundariesViolated
	}

	return nil
}

func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.connection.Close()
}
