package types

import (
	"context"
	"cosmossdk.io/math"
)

// EthereumSidecarClient is an interface for a client that can interact with the
// Ethereum sidecar.
type EthereumSidecarClient interface {
	// GetAssetsLockedEvents returns confirmed AssetsLockedEvents with
	// the sequence number falling within the half-open range, denoted by
	// sequenceStart (included) and sequenceEnd (excluded). Nil can be
	// passed for sequenceStart and sequenceEnd to indicate an unbounded
	// edge of the range.
	GetAssetsLockedEvents(
		ctx context.Context,
		sequenceStart *math.Int,
		sequenceEnd *math.Int,
	) ([]AssetsLockedEvent, error)
}
