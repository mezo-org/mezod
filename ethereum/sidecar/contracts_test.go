package sidecar

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

func NewLocalBridgeContract() *LocalBridgeContract {
	return &LocalBridgeContract{
		events: make([]*portal.MezoBridgeAssetsLocked, 0),
	}
}

type LocalBridgeContract struct {
	mutex  sync.RWMutex
	events []*portal.MezoBridgeAssetsLocked
	// queue of errors that will be returned on subsequent calls
	// to `FilterAssetsLocked`
	errors []error
}

func (m *LocalBridgeContract) FilterAssetsLocked(
	_ *bind.FilterOpts,
	_ []*big.Int,
	_ []common.Address,
	_ []common.Address,
) (ethereum.AssetsLockedIterator, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var err error
	if len(m.errors) > 0 {
		err = m.errors[0]
		// pop first error
		m.errors = m.errors[1:]
	}

	if err != nil {
		return nil, err
	}

	return &LocalAssetsLockedIterator{
		events: m.events,
		index:  -1,
	}, nil
}

func (m *LocalBridgeContract) SetEvents(events []*portal.MezoBridgeAssetsLocked) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = events
}

func (m *LocalBridgeContract) SetErrors(errors []error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors = errors
}

type LocalAssetsLockedIterator struct {
	events []*portal.MezoBridgeAssetsLocked
	index  int
}

func (m *LocalAssetsLockedIterator) Next() bool {
	m.index++
	return m.index < len(m.events)
}

func (m *LocalAssetsLockedIterator) Error() error {
	return nil
}

func (m *LocalAssetsLockedIterator) Close() error {
	return nil
}

func (m *LocalAssetsLockedIterator) Event() *portal.MezoBridgeAssetsLocked {
	return m.events[m.index]
}
