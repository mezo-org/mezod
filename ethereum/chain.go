package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

type Chain interface {
	Key() *keystore.Key
	FinalizedBlock(ctx context.Context) (*big.Int, error)
	LatestBlock(ctx context.Context) (*big.Int, error)
	WatchBlocks(ctx context.Context) <-chan uint64
}
