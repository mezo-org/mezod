package bridgeworker

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdkmath "cosmossdk.io/math"
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

// requestTimeout is the timeout for requests sent to the mezod validator node.
const requestTimeout = 5 * time.Second

// AssetsUnlockedGrpcEndpoint enables gRPC communication with mezod validator node
// needed for the bridge-out process.
type AssetsUnlockedGrpcEndpoint struct {
	connection *grpc.ClientConn
}

func NewAssetsUnlockedGrpcEndpoint(
	serverAddress string,
	registry types.InterfaceRegistry,
) (*AssetsUnlockedGrpcEndpoint, error) {
	connection, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(registry).GRPCCodec()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create gRPC bridge-out client for Ethereum sidecar [%w]",
			err,
		)
	}

	endpoint := &AssetsUnlockedGrpcEndpoint{
		connection: connection,
	}

	return endpoint, nil
}

// GetAssetsUnlockedEvents gets the AssetsUnlocked events from the Mezo chain.
// The requested range of events is inclusive on the lower side and exclusive
// on the upper side.
func (auge *AssetsUnlockedGrpcEndpoint) GetAssetsUnlockedEvents(
	ctx context.Context,
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
) ([]bridgetypes.AssetsUnlockedEvent, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	queryClient := bridgetypes.NewQueryClient(auge.connection)

	request := &bridgetypes.QueryAssetsUnlockedEventsRequest{
		SequenceStart: sequenceStart,
		SequenceEnd:   sequenceEnd,
	}

	response, err := queryClient.AssetsUnlockedEvents(ctxWithTimeout, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetsUnlocked events: [%w]", err)
	}

	err = validateAssetsUnlockedEntries(
		sequenceStart,
		sequenceEnd,
		response.Events,
	)
	if err != nil {
		return nil, err
	}

	return response.Events, nil
}

// validateAssetsUnlockedEntries validates the AssetsUnlocked events fetched
// from the Mezo chain.
func validateAssetsUnlockedEntries(
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
	events []bridgetypes.AssetsUnlockedEvent,
) error {
	if len(events) == 0 {
		return nil
	}

	if !bridgetypes.AssetsUnlockedEvents(events).IsValid() {
		return ErrInvalidEventsSequence
	}

	if (!sequenceStart.IsNil() && events[0].UnlockSequence.LT(sequenceStart)) ||
		(!sequenceEnd.IsNil() && events[len(events)-1].UnlockSequence.GTE(sequenceEnd)) {
		return ErrRequestedBoundariesViolated
	}

	return nil
}
