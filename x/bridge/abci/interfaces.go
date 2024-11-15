package abci

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/math"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// EthereumSidecarClient is an interface for a client that can interact with the
// Ethereum sidecar.
type EthereumSidecarClient interface {
	// GetAssetsLockedEvents returns confirmed AssetsLockedEvents with
	// the sequence number falling within the half-open range, denoted by
	// sequenceStart (included) and sequenceEnd (excluded). To indicate an
	// unbounded edge of the range, use a zero initialized value (`math.Int{}`)
	// for sequenceStart or sequenceEnd.
	//
	// The implementation should ensure that sequence numbers of the returned
	// events form a sequence strictly increasing by 1. Such a sequence
	// guarantees that there are no gaps between the sequence numbers of the
	// events and that each event is unique by its sequence number.
	GetAssetsLockedEvents(
		ctx context.Context,
		sequenceStart math.Int,
		sequenceEnd math.Int,
	) ([]types.AssetsLockedEvent, error)
}

// BridgeKeeper is an interface to the x/bridge module keeper.
type BridgeKeeper interface {
	GetAssetsLockedSequenceTip(ctx sdk.Context) math.Int
	AcceptAssetsLocked(ctx sdk.Context, events types.AssetsLockedEvents) error
}
