package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
)

func NewMezoBridgeContract(
	delegate *portal.MezoBridge,
) *MezoBridgeContract {
	return &MezoBridgeContract{
		delegate: delegate,
	}
}

type MezoBridgeContract struct {
	delegate *portal.MezoBridge
}

func (r *MezoBridgeContract) PastAssetsUnlockConfirmedEvents(
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

func (r *MezoBridgeContract) TbtcToken() (common.Address, error) {
	return r.delegate.TbtcToken()
}

func (r *MezoBridgeContract) PendingBTCWithdrawals(entryHash [32]byte) (bool, error) {
	return r.delegate.PendingBTCWithdrawals(entryHash)
}

func (r *MezoBridgeContract) WithdrawBTC(
	entry portal.MezoBridgeAssetsUnlocked,
	walletPubKeyHash [20]byte,
	mainUtxo portal.BitcoinTxUTXO,
) (*types.Transaction, error) {
	return r.delegate.WithdrawBTC(entry, walletPubKeyHash, mainUtxo)
}

func NewTbtcBridgeContract(
	delegate *tbtc.Bridge,
) *TbtcBridgeContract {
	return &TbtcBridgeContract{
		delegate: delegate,
	}
}

type TbtcBridgeContract struct {
	delegate *tbtc.Bridge
}

func (r *TbtcBridgeContract) PastNewWalletRegisteredEvents(
	startBlock uint64,
	endBlock *uint64,
	ecdsaWalletID [][32]byte,
	walletPubKeyHash [][20]byte,
) ([]*tbtc.BridgeNewWalletRegistered, error) {
	events, err := r.delegate.PastNewWalletRegisteredEvents(
		startBlock,
		endBlock,
		ecdsaWalletID,
		walletPubKeyHash,
	)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *TbtcBridgeContract) Wallets(walletPublicKeyHash [20]byte) (tbtc.Wallet, error) {
	return r.delegate.Wallets(walletPublicKeyHash)
}

func (r *TbtcBridgeContract) PendingRedemptions(redemptionKey *big.Int) (tbtc.RedemptionRequest, error) {
	return r.delegate.PendingRedemptions(redemptionKey)
}

func (r *TbtcBridgeContract) RedemptionDustThreshold() (uint64, error) {
	redemptionParameters, err := r.delegate.RedemptionParameters()
	if err != nil {
		return 0, err
	}
	return redemptionParameters.RedemptionDustThreshold, nil
}
