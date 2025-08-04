package sidecar

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum"
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

func (r *BridgeContract) FilterAssetsLocked(
	opts *bind.FilterOpts,
	sequenceNumber []*big.Int,
	recipient []common.Address,
	token []common.Address,
) (ethereum.AssetsLockedIterator, error) {
	iter, err := r.delegate.FilterAssetsLocked(opts, sequenceNumber, recipient, token)
	if err != nil {
		return nil, err
	}
	return &AssetsLockedIterator{iter: iter}, nil
}

func (r *BridgeContract) PastAssetsUnlockConfirmedEvents(
	startBlock uint64,
	endBlock *uint64,
	unlockSequenceNumberFilter []*big.Int,
	recipientFilter [][]byte,
	tokenFilter []common.Address,
) ([]*ethereum.MezoBridgeAssetsUnlockConfirmed, error) {
	// TODO: Leaving unimplemented for now. Call `PastAssetsUnlockConfirmedEvents`
	//       on r.delegate once bindings for MezoBridge re-generated.
	return nil, nil
}

func (r *BridgeContract) ConfirmedUnlocks(arg0 *big.Int) (bool, error) {
	// TODO: Leaving unimplemented for now. Call `ConfirmedUnlocks`
	//       on r.delegate once bindings for MezoBridge re-generated.
	return false, nil
}

type AssetsLockedIterator struct {
	iter *portal.MezoBridgeAssetsLockedIterator
}

func (r *AssetsLockedIterator) Next() bool {
	return r.iter.Next()
}

func (r *AssetsLockedIterator) Error() error {
	return r.iter.Error()
}

func (r *AssetsLockedIterator) Close() error {
	return r.iter.Close()
}

func (r *AssetsLockedIterator) Event() *portal.MezoBridgeAssetsLocked {
	return r.iter.Event
}
