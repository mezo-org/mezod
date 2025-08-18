package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

type BridgeContract interface {
	ValidatorIDs(opts *bind.CallOpts, address common.Address) (uint8, error)
	ConfirmedUnlocks(opts *bind.CallOpts, sequenceNumber *big.Int) (bool, error)
	Attestations(opts *bind.CallOpts, hash [32]byte) (*big.Int, error)
	PastAssetsLockedEvents(
		startBlock uint64,
		endBlock *uint64,
		sequenceNumber []*big.Int,
		recipient []common.Address,
		token []common.Address,
	) ([]*portal.MezoBridgeAssetsLocked, error)

	PastAssetsUnlockConfirmedEvents(
		startBlock uint64,
		endBlock *uint64,
		unlockSequenceNumber []*big.Int,
		recipient [][]byte,
		token []common.Address,
	) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error)
}
