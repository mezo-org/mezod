package ethereum

import (
	"context"
	"math/big"
)

type Chain interface {
	FinalizedBlock(ctx context.Context) (*big.Int, error)
	LatestBlock(ctx context.Context) (*big.Int, error)
	WatchBlocks(ctx context.Context) <-chan uint64
}
