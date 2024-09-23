package sidecar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// Client connects to the Ethereum sidecar server and queries for the
// `AssetsLocked` events.
type Client struct {
	serverAddress  string
	requestTimeout time.Duration
	logger         log.Logger
}

func NewClient(
	serverAddress string,
	requestTimeout time.Duration,
	logger log.Logger,
) *Client {
	// TODO: Consider adding validation of the connection.
	return &Client{
		serverAddress:  serverAddress,
		requestTimeout: requestTimeout,
		logger:         logger,
	}
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
	// Establish connection to the server.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	// TODO: Consider adding options for maximum message size.
	serverConnection, err := grpc.Dial(
		c.serverAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to the Ethereum sidecar server: [%v]",
			err,
		)
	}
	defer serverConnection.Close()

	c.logger.Info("successfully dialed the Ethereum sidecar server")

	sidecarClient := pb.NewEthereumSidecarClient(serverConnection)

	// Query the server for the `AssetsLocked` events.
	request := &pb.AssetsLockedEventsRequest{
		SequenceStart: *sequenceStart,
		SequenceEnd:   *sequenceEnd,
	}

	response, err := sidecarClient.AssetsLockedEvents(ctxWithTimeout, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsLocked events [%v]", err)
	}

	c.logger.Info(
		"received response from the Ethereum sidecar server",
		"number of events", len(response.Events),
	)

	events := make([]bridgetypes.AssetsLockedEvent, len(response.Events))
	for i, event := range response.Events {
		events[i] = *event
	}

	// TODO: Should we validate the events (make sure each event has a sequence
	//       one greater than the previous one)? The code that uses the client
	//       already does that.
	return events, nil
}
