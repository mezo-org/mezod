package sidecar

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	assetsUnlockConfirmedEvents []*portal.MezoBridgeAssetsUnlockConfirmed
	// queue of errors that will be returned on subsequent calls
	// to `PastAssetsLockedEvents` and `PastAssetsUnlockConfirmedEvents`
	errors []error
}

func (lbc *localBridgeContract) PastAssetsLockedEvents(
	_ uint64,
	_ *uint64,
	_ []*big.Int,
	_ []common.Address,
	_ []common.Address,
) ([]*portal.MezoBridgeAssetsLocked, error) {
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

	return lbc.assetsLockedEvents, nil
}

func (lbc *localBridgeContract) PastAssetsUnlockConfirmedEvents(
	_ uint64,
	_ *uint64,
	_ []*big.Int,
	_ [][]byte,
	_ []common.Address,
) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error) {
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

	return lbc.assetsUnlockConfirmedEvents, nil
}

func (lbc *localBridgeContract) SetAssetsLockedEvents(events []*portal.MezoBridgeAssetsLocked) {
	lbc.mutex.Lock()
	defer lbc.mutex.Unlock()
	lbc.assetsLockedEvents = events
}

func (lbc *localBridgeContract) SetAssetsUnlockConfirmedEvents(
	events []*portal.MezoBridgeAssetsUnlockConfirmed,
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

func (lbc *localBridgeContract) ValidateAssetsUnlocked(
	_ portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	return true, nil
}

func (lbc *localBridgeContract) AttestBridgeOut(
	_ *portal.MezoBridgeAssetsUnlocked,
) (*types.Transaction, error) {
	return nil, nil
}

func (lbc *localBridgeContract) ValidatorIDs(_ common.Address) (uint8, error) {
	return 0, nil
}

func (lbc *localBridgeContract) ConfirmedUnlocks(_ *big.Int) (bool, error) {
	return true, nil
}

func (lbc *localBridgeContract) Attestations(_ [32]byte) (*big.Int, error) {
	return nil, nil
}

func (lbc *localBridgeContract) BridgeValidatorsCount() (*big.Int, error) {
	return big.NewInt(0), nil
}
