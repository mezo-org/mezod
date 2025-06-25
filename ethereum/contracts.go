package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

type AssetsLockedIterator interface {
	Next() bool
	Error() error
	Close() error
	Event() *portal.MezoBridgeAssetsLocked
}

type BridgeContract interface {
	FilterAssetsLocked(
		opts *bind.FilterOpts,
		sequenceNumber []*big.Int,
		recipient []common.Address,
		token []common.Address,
	) (AssetsLockedIterator, error)
}
