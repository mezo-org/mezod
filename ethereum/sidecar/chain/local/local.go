package local

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/sidecar/chain"
)

func NewBridgeContract() *BridgeContract {
	return &BridgeContract{
		events: make([]*portal.MezoBridgeAssetsLocked, 0),
	}
}

type BridgeContract struct {
	mutex  sync.RWMutex
	events []*portal.MezoBridgeAssetsLocked
	// queue of errors that will be returned on subsequent calls
	// to `FilterAssetsLocked`
	errors []error
}

func (m *BridgeContract) FilterAssetsLocked(
	_ *bind.FilterOpts,
	_ []*big.Int,
	_ []common.Address,
	_ []common.Address,
) (chain.AssetsLockedIterator, error) {
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

	return &AssetsLockedIterator{
		events: m.events,
		index:  -1,
	}, nil
}

func (m *BridgeContract) SetEvents(events []*portal.MezoBridgeAssetsLocked) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = events
}

func (m *BridgeContract) SetErrors(errors []error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors = errors
}

type AssetsLockedIterator struct {
	events []*portal.MezoBridgeAssetsLocked
	index  int
}

func (m *AssetsLockedIterator) Next() bool {
	m.index++
	return m.index < len(m.events)
}

func (m *AssetsLockedIterator) Error() error {
	return nil
}

func (m *AssetsLockedIterator) Close() error {
	return nil
}

func (m *AssetsLockedIterator) Event() *portal.MezoBridgeAssetsLocked {
	return m.events[m.index]
}
