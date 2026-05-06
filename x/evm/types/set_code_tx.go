// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package types

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"

	"github.com/mezo-org/mezod/types"
)

// NewSetCodeTx constructs a SetCodeTx from a geth ethtypes.Transaction of
// type SetCodeTxType. It mirrors NewDynamicFeeTx but additionally carries
// the EIP-7702 authorization list.
func NewSetCodeTx(tx *ethtypes.Transaction) (*SetCodeTx, error) {
	txData := &SetCodeTx{
		Nonce:    tx.Nonce(),
		Data:     tx.Data(),
		GasLimit: tx.Gas(),
	}

	v, r, s := tx.RawSignatureValues()
	if to := tx.To(); to != nil {
		txData.To = to.Hex()
	}

	if tx.Value() != nil {
		amountInt, err := types.SafeNewIntFromBigInt(tx.Value())
		if err != nil {
			return nil, err
		}
		txData.Amount = &amountInt
	}

	if tx.GasFeeCap() != nil {
		gasFeeCapInt, err := types.SafeNewIntFromBigInt(tx.GasFeeCap())
		if err != nil {
			return nil, err
		}
		txData.GasFeeCap = &gasFeeCapInt
	}

	if tx.GasTipCap() != nil {
		gasTipCapInt, err := types.SafeNewIntFromBigInt(tx.GasTipCap())
		if err != nil {
			return nil, err
		}
		txData.GasTipCap = &gasTipCapInt
	}

	if tx.AccessList() != nil {
		al := tx.AccessList()
		txData.Accesses = NewAccessList(&al)
	}

	txData.AuthList = NewAuthorizationList(tx.SetCodeAuthorizations())

	txData.SetSignatureValues(tx.ChainId(), v, r, s)
	return txData, nil
}

// TxType returns the tx type
func (tx *SetCodeTx) TxType() uint8 {
	return ethtypes.SetCodeTxType
}

// Copy returns an instance with the same field values.
func (tx *SetCodeTx) Copy() TxData {
	var authList AuthorizationList
	if tx.AuthList != nil {
		authList = make(AuthorizationList, len(tx.AuthList))
		for i, a := range tx.AuthList {
			authList[i] = SetCodeAuthorization{
				ChainID: a.ChainID,
				Address: a.Address,
				Nonce:   a.Nonce,
				V:       common.CopyBytes(a.V),
				R:       common.CopyBytes(a.R),
				S:       common.CopyBytes(a.S),
			}
		}
	}

	return &SetCodeTx{
		ChainID:   tx.ChainID,
		Nonce:     tx.Nonce,
		GasTipCap: tx.GasTipCap,
		GasFeeCap: tx.GasFeeCap,
		GasLimit:  tx.GasLimit,
		To:        tx.To,
		Amount:    tx.Amount,
		Data:      common.CopyBytes(tx.Data),
		Accesses:  tx.Accesses,
		AuthList:  authList,
		V:         common.CopyBytes(tx.V),
		R:         common.CopyBytes(tx.R),
		S:         common.CopyBytes(tx.S),
	}
}

// GetChainID returns the chain id field from the SetCodeTx
func (tx *SetCodeTx) GetChainID() *big.Int {
	if tx.ChainID == nil {
		return nil
	}

	return tx.ChainID.BigInt()
}

// GetAccessList returns the AccessList field.
func (tx *SetCodeTx) GetAccessList() ethtypes.AccessList {
	if tx.Accesses == nil {
		return nil
	}
	return *tx.Accesses.ToEthAccessList()
}

// GetAuthorizationList returns the EIP-7702 authorization list.
func (tx *SetCodeTx) GetAuthorizationList() []ethtypes.SetCodeAuthorization {
	return tx.AuthList.ToEthAuthorizationList()
}

// GetData returns a copy of the input data bytes.
func (tx *SetCodeTx) GetData() []byte {
	return common.CopyBytes(tx.Data)
}

// GetGas returns the gas limit.
func (tx *SetCodeTx) GetGas() uint64 {
	return tx.GasLimit
}

// GetGasPrice returns the gas fee cap field.
func (tx *SetCodeTx) GetGasPrice() *big.Int {
	return tx.GetGasFeeCap()
}

// GetGasTipCap returns the gas tip cap field.
func (tx *SetCodeTx) GetGasTipCap() *big.Int {
	if tx.GasTipCap == nil {
		return nil
	}
	return tx.GasTipCap.BigInt()
}

// GetGasFeeCap returns the gas fee cap field.
func (tx *SetCodeTx) GetGasFeeCap() *big.Int {
	if tx.GasFeeCap == nil {
		return nil
	}
	return tx.GasFeeCap.BigInt()
}

// GetValue returns the tx amount.
func (tx *SetCodeTx) GetValue() *big.Int {
	if tx.Amount == nil {
		return nil
	}

	return tx.Amount.BigInt()
}

// GetNonce returns the account sequence for the transaction.
func (tx *SetCodeTx) GetNonce() uint64 { return tx.Nonce }

// GetTo returns the pointer to the recipient address.
func (tx *SetCodeTx) GetTo() *common.Address {
	if tx.To == "" {
		return nil
	}
	to := common.HexToAddress(tx.To)
	return &to
}

// AsEthereumData returns a geth-side SetCodeTx built from this proto-formatted
// TxData. Validate() must have run first: it guarantees ChainID, gas caps,
// V/R/S, and To are all within bounds for uint256.MustFromBig. A nil Amount
// is treated as zero, mirroring DynamicFeeTx semantics.
func (tx *SetCodeTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()

	var to common.Address
	if t := tx.GetTo(); t != nil {
		to = *t
	}

	chainID := uint256.MustFromBig(tx.GetChainID())
	gasTipCap := uint256.MustFromBig(tx.GetGasTipCap())
	gasFeeCap := uint256.MustFromBig(tx.GetGasFeeCap())
	value := new(uint256.Int)
	if amount := tx.GetValue(); amount != nil {
		value = uint256.MustFromBig(amount)
	}

	vU := new(uint256.Int)
	if v != nil {
		vU = uint256.MustFromBig(v)
	}
	rU := new(uint256.Int)
	if r != nil {
		rU = uint256.MustFromBig(r)
	}
	sU := new(uint256.Int)
	if s != nil {
		sU = uint256.MustFromBig(s)
	}

	return &ethtypes.SetCodeTx{
		ChainID:    chainID,
		Nonce:      tx.GetNonce(),
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		Gas:        tx.GetGas(),
		To:         to,
		Value:      value,
		Data:       tx.GetData(),
		AccessList: tx.GetAccessList(),
		AuthList:   tx.GetAuthorizationList(),
		V:          vU,
		R:          rU,
		S:          sU,
	}
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *SetCodeTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return rawSignatureValues(tx.V, tx.R, tx.S)
}

// SetSignatureValues sets the signature values to the transaction.
func (tx *SetCodeTx) SetSignatureValues(chainID, v, r, s *big.Int) {
	if v != nil {
		tx.V = v.Bytes()
	}
	if r != nil {
		tx.R = r.Bytes()
	}
	if s != nil {
		tx.S = s.Bytes()
	}
	if chainID != nil {
		chainIDInt := sdkmath.NewIntFromBigInt(chainID)
		tx.ChainID = &chainIDInt
	}
}

// Validate performs a stateless validation of the tx fields.
func (tx SetCodeTx) Validate() error {
	if tx.GasTipCap == nil {
		return errorsmod.Wrap(ErrInvalidGasCap, "gas tip cap cannot nil")
	}

	if tx.GasFeeCap == nil {
		return errorsmod.Wrap(ErrInvalidGasCap, "gas fee cap cannot nil")
	}

	if tx.GasTipCap.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidGasCap, "gas tip cap cannot be negative %s", tx.GasTipCap)
	}

	if tx.GasFeeCap.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidGasCap, "gas fee cap cannot be negative %s", tx.GasFeeCap)
	}

	if !types.IsValidInt256(tx.GetGasTipCap()) {
		return errorsmod.Wrap(ErrInvalidGasCap, "out of bound")
	}

	if !types.IsValidInt256(tx.GetGasFeeCap()) {
		return errorsmod.Wrap(ErrInvalidGasCap, "out of bound")
	}

	if tx.GasFeeCap.LT(*tx.GasTipCap) {
		return errorsmod.Wrapf(
			ErrInvalidGasCap, "max priority fee per gas higher than max fee per gas (%s > %s)",
			tx.GasTipCap, tx.GasFeeCap,
		)
	}

	if !types.IsValidInt256(tx.Fee()) {
		return errorsmod.Wrap(ErrInvalidGasFee, "out of bound")
	}

	amount := tx.GetValue()
	// Amount can be nil (treated as zero) or zero, mirroring DynamicFeeTx.
	if amount != nil && amount.Sign() == -1 {
		return errorsmod.Wrapf(ErrInvalidAmount, "amount cannot be negative %s", amount)
	}
	if !types.IsValidInt256(amount) {
		return errorsmod.Wrap(ErrInvalidAmount, "out of bound")
	}

	if tx.To == "" {
		return errorsmod.Wrap(ErrSetCodeMissingTo, "to address required")
	}
	if err := types.ValidateAddress(tx.To); err != nil {
		return errorsmod.Wrap(err, "invalid to address")
	}

	chainID := tx.GetChainID()

	if chainID == nil {
		return errorsmod.Wrap(
			errortypes.ErrInvalidChainID,
			"chain ID must be present on SetCode txs",
		)
	}

	if !types.IsValidInt256(chainID) {
		return errorsmod.Wrap(errortypes.ErrInvalidChainID, "out of bound")
	}

	if !(chainID.Cmp(big.NewInt(31612)) == 0 || chainID.Cmp(big.NewInt(31611)) == 0) {
		return errorsmod.Wrapf(
			errortypes.ErrInvalidChainID,
			"chain ID must be 31611 or 31612 on Mezo, got %s", chainID,
		)
	}

	if v, r, s := tx.GetRawSignatureValues(); !types.IsValidInt256(v) ||
		!types.IsValidInt256(r) || !types.IsValidInt256(s) {
		return errorsmod.Wrap(ErrInvalidSigner, "V, R or S out of bound")
	}

	if len(tx.AuthList) == 0 {
		return errorsmod.Wrap(ErrSetCodeEmptyAuthList, "auth list must be non-empty")
	}

	var vTmp, rTmp, sTmp big.Int
	for i, auth := range tx.AuthList {
		// The tx-level chain ID above is bound to a specific Mezo network.
		// Per-authorization chain IDs are intentionally only nil-and-bounds
		// checked here: EIP-7702 allows auth.ChainID to be 0 (cross-chain) or
		// the current chain id, with any other value making that single tuple
		// invalid. Per EIP-7702 invalid tuples are silently skipped at apply
		// time, not tx-fatal, so the 0-or-current check belongs in the keeper.
		if auth.ChainID == nil {
			return errorsmod.Wrapf(
				errortypes.ErrInvalidChainID,
				"authorization[%d] chain ID cannot be nil", i,
			)
		}
		if !types.IsValidInt256(auth.ChainID.BigInt()) {
			return errorsmod.Wrapf(
				errortypes.ErrInvalidChainID,
				"authorization[%d] chain ID out of bound", i,
			)
		}
		if err := types.ValidateAddress(auth.Address); err != nil {
			return errorsmod.Wrapf(err, "authorization[%d] invalid address", i)
		}
		// Auth V/R/S shape is checked only for int256 bounds, mirroring
		// tx-level signature validation. Canonical-form checks (V in {0,1},
		// R/S non-zero, low-S, length <= 32) are deferred to the keeper's
		// authorization recovery, where failures result in silent per-tuple
		// skips per EIP-7702.
		vTmp.SetBytes(auth.V)
		rTmp.SetBytes(auth.R)
		sTmp.SetBytes(auth.S)
		if !types.IsValidInt256(&vTmp) ||
			!types.IsValidInt256(&rTmp) ||
			!types.IsValidInt256(&sTmp) {
			return errorsmod.Wrapf(
				ErrInvalidSigner,
				"authorization[%d] V, R or S out of bound", i,
			)
		}
	}

	return nil
}

// Fee returns gasprice * gaslimit.
func (tx SetCodeTx) Fee() *big.Int {
	return fee(tx.GetGasFeeCap(), tx.GasLimit)
}

// Cost returns amount + gasprice * gaslimit.
func (tx SetCodeTx) Cost() *big.Int {
	return cost(tx.Fee(), tx.GetValue())
}

// EffectiveGasPrice returns the effective gas price
func (tx *SetCodeTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	return EffectiveGasPrice(baseFee, tx.GasFeeCap.BigInt(), tx.GasTipCap.BigInt())
}

// EffectiveFee returns effective_gasprice * gaslimit.
func (tx SetCodeTx) EffectiveFee(baseFee *big.Int) *big.Int {
	return fee(tx.EffectiveGasPrice(baseFee), tx.GasLimit)
}

// EffectiveCost returns amount + effective_gasprice * gaslimit.
func (tx SetCodeTx) EffectiveCost(baseFee *big.Int) *big.Int {
	return cost(tx.EffectiveFee(baseFee), tx.GetValue())
}
