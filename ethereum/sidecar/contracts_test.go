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
		events: make([]*portal.MezoBridgeAssetsLocked, 0),
	}
}

type localBridgeContract struct {
	mutex  sync.RWMutex
	events []*portal.MezoBridgeAssetsLocked
	// queue of errors that will be returned on subsequent calls
	// to `FilterAssetsLocked`
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
		events: lbc.events,
		index:  -1,
	}, nil
}

func (lbc *localBridgeContract) SetEvents(events []*portal.MezoBridgeAssetsLocked) {
	lbc.mutex.Lock()
	defer lbc.mutex.Unlock()
	lbc.events = events
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
