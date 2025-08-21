package ethereum

import (
	"context"
	"math/big"

	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
)

type Chain interface {
	FinalizedBlock(ctx context.Context) (*big.Int, error)
	WatchBlocks(ctx context.Context) <-chan uint64
	Client() ethutil.EthereumClient
}
