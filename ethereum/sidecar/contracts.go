package sidecar

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

func NewBridgeContract(
	delegate *portal.MezoBridge,
) *BridgeContract {
	return &BridgeContract{
		delegate: delegate,
	}
}

type BridgeContract struct {
	delegate *portal.MezoBridge
}

func (r *BridgeContract) ValidateAssetsUnlocked(
	assetsUnlocked portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	return r.delegate.ValidateAssetsUnlocked(assetsUnlocked)
}

func (r *BridgeContract) AttestBridgeOut(
	assetsUnlocked *portal.MezoBridgeAssetsUnlocked,
) (*types.Transaction, error) {
	return r.delegate.AttestBridgeOut(*assetsUnlocked)
}

func (r *BridgeContract) ValidatorIDs(address common.Address) (uint8, error) {
	return r.delegate.BridgeValidatorIDs(address)
}

func (r *BridgeContract) ConfirmedUnlocks(sequenceNumber *big.Int) (bool, error) {
	return r.delegate.ConfirmedUnlocks(sequenceNumber)
}

func (r *BridgeContract) Attestations(hash [32]byte) (*big.Int, error) {
	return r.delegate.Attestations(hash)
}

func (r *BridgeContract) PastAssetsLockedEvents(
	startBlock uint64,
	endBlock *uint64,
	sequenceNumber []*big.Int,
	recipient []common.Address,
	token []common.Address,
) ([]*portal.MezoBridgeAssetsLocked, error) {
	events, err := r.delegate.PastAssetsLockedEvents(
		startBlock,
		endBlock,
		sequenceNumber,
		recipient,
		token,
	)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *BridgeContract) PastAssetsUnlockConfirmedEvents(
	startBlock uint64,
	endBlock *uint64,
	unlockSequenceNumber []*big.Int,
	recipient [][]byte,
	token []common.Address,
) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error) {
	events, err := r.delegate.PastAssetsUnlockConfirmedEvents(
		startBlock,
		endBlock,
		unlockSequenceNumber,
		recipient,
		token,
	)
	if err != nil {
		return nil, err
	}
	return events, nil
}
