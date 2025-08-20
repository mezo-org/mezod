package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/portal/sepolia/gen/abi"
)

type BridgeContract interface {
	ValidateAssetsUnlocked(
		assetsUnlocked abi.MezoBridgeAssetsUnlocked,
	) (bool, error)
	AttestBridgeOut(
		assetsUnlocked *portal.MezoBridgeAssetsUnlocked,
	) (*types.Transaction, error)
	ValidatorIDs(address common.Address) (uint8, error)
	ConfirmedUnlocks(sequenceNumber *big.Int) (bool, error)
	Attestations(hash [32]byte) (*big.Int, error)
	BridgeValidatorsCount() (*big.Int, error)
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
