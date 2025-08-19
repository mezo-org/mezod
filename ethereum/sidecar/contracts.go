package sidecar

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
