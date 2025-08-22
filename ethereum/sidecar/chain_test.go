package sidecar

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

func newLocalChain() *localChain {
	return &localChain{}
}

type localChain struct {
	finalizedBlock *big.Int
}

func (lc *localChain) Key() *keystore.Key {
	panic("unimplemented")
}

func (lc *localChain) FinalizedBlock(_ context.Context) (*big.Int, error) {
	return lc.finalizedBlock, nil
}

func (lc *localChain) WatchBlocks(_ context.Context) <-chan uint64 {
	panic("unimplemented")
}

func (lc *localChain) LatestBlock(_ context.Context) (*big.Int, error) {
	panic("unimplemented")
}

func (lc *localChain) setFinalizedBlock(finalizedBlock *big.Int) {
	lc.finalizedBlock = finalizedBlock
}
