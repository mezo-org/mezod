package ethereum

import (
	"context"
	"math/big"
)

type Chain interface {
	CurrentBlock() (uint64, error)
	FinalizedBlock(ctx context.Context) (*big.Int, error)
	WatchBlocks(ctx context.Context) <-chan uint64
}
