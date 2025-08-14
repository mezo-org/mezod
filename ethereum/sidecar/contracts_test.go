package sidecar

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

func newLocalBridgeContract() *localBridgeContract {
	return &localBridgeContract{
		assetsLockedEvents: make([]*portal.MezoBridgeAssetsLocked, 0),
	}
}

type localBridgeContract struct {
	mutex                       sync.RWMutex
	assetsLockedEvents          []*portal.MezoBridgeAssetsLocked
	assetsUnlockConfirmedEvents []*ethereum.MezoBridgeAssetsUnlockConfirmed
	// queue of errors that will be returned on subsequent calls
	// to `FilterAssetsLocked` and `FilterAssetsUnlockConfirmed`
	errors []error
}

func (lbc *localBridgeContract) FilterAssetsLocked(
	_ *bind.FilterOpts,
	_ []*big.Int,
	_ []common.Address,
	_ []common.Address,
) (ethereum.AssetsLockedIterator, error) {
	lbc.mutex.Lock()
	defer lbc.mutex.Unlock()

	var err error
	if len(lbc.errors) > 0 {
		err = lbc.errors[0]
		// pop first error
		lbc.errors = lbc.errors[1:]
	}

	if err != nil {
		return nil, err
	}

	return &localAssetsLockedIterator{
		events: lbc.assetsLockedEvents,
		index:  -1,
	}, nil
}

func (lbc *localBridgeContract) FilterAssetsUnlockConfirmed(
	_ *bind.FilterOpts,
	_ []*big.Int,
	_ [][]byte,
	_ []common.Address,
) (ethereum.AssetsUnlockConfirmedIterator, error) {
	lbc.mutex.Lock()
	defer lbc.mutex.Unlock()

	var err error
	if len(lbc.errors) > 0 {
		err = lbc.errors[0]
		// pop first error
		lbc.errors = lbc.errors[1:]
	}

	if err != nil {
		return nil, err
	}

	return &localAssetsUnlockConfirmedIterator{
		events: lbc.assetsUnlockConfirmedEvents,
		index:  -1,
	}, nil
}

func (lbc *localBridgeContract) SetAssetsLockedEvents(events []*portal.MezoBridgeAssetsLocked) {
	lbc.mutex.Lock()
	defer lbc.mutex.Unlock()
	lbc.assetsLockedEvents = events
}

func (lbc *localBridgeContract) SetAssetsUnlockConfirmedEvents(
	events []*ethereum.MezoBridgeAssetsUnlockConfirmed,
) {
	lbc.mutex.Lock()
	defer lbc.mutex.Unlock()
	lbc.assetsUnlockConfirmedEvents = events
}

func (lbc *localBridgeContract) SetErrors(errors []error) {
	lbc.mutex.Lock()
	defer lbc.mutex.Unlock()
	lbc.errors = errors
}

type localAssetsLockedIterator struct {
	events []*portal.MezoBridgeAssetsLocked
	index  int
}

func (laci *localAssetsLockedIterator) Next() bool {
	laci.index++
	return laci.index < len(laci.events)
}

func (laci *localAssetsLockedIterator) Error() error {
	return nil
}

func (laci *localAssetsLockedIterator) Close() error {
	return nil
}

func (laci *localAssetsLockedIterator) Event() *portal.MezoBridgeAssetsLocked {
	return laci.events[laci.index]
}

type localAssetsUnlockConfirmedIterator struct {
	events []*ethereum.MezoBridgeAssetsUnlockConfirmed
	index  int
}

func (lauci *localAssetsUnlockConfirmedIterator) Next() bool {
	lauci.index++
	return lauci.index < len(lauci.events)
}

func (lauci *localAssetsUnlockConfirmedIterator) Error() error {
	return nil
}

func (lauci *localAssetsUnlockConfirmedIterator) Close() error {
	return nil
}

func (lauci *localAssetsUnlockConfirmedIterator) Event() *ethereum.MezoBridgeAssetsUnlockConfirmed {
	return lauci.events[lauci.index]
}
