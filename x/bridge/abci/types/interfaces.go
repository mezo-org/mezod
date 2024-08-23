package types

import (
	"context"

	"cosmossdk.io/math"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// EthereumSidecarClient is an interface for a client that can interact with the
// Ethereum sidecar.
type EthereumSidecarClient interface {
	// GetAssetsLockedEvents returns confirmed AssetsLockedEvents with
	// the sequence number falling within the half-open range, denoted by
	// sequenceStart (included) and sequenceEnd (excluded). Nil can be
	// passed for sequenceStart and sequenceEnd to indicate an unbounded
	// edge of the range.
	//
	// The implementation should ensure that sequence numbers of the returned
	// events form a sequence strictly increasing by 1. Such a sequence
	// guarantees that there are no gaps between the sequence numbers of the
	// events and that each event is unique by its sequence number.
	GetAssetsLockedEvents(
		ctx context.Context,
		sequenceStart *math.Int,
		sequenceEnd *math.Int,
	) ([]types.AssetsLockedEvent, error)
}
