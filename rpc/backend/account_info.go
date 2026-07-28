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
package backend

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	rpctypes "github.com/mezo-org/mezod/rpc/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/pkg/errors"
)

// GetCode returns the contract code at the given address and block number.
func (b *Backend) GetCode(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Code(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetProof returns an account object with proof and any storage proofs.
//
// The node's GetProofStorageKeysCap setting limits the storage keys as given,
// which bounds the size of a single request and can be checked before any store
// is touched. Keys are decoded next, so a malformed key is reported instead of
// being coerced into the zero key. Repeated keys are then resolved once, so the
// number of store queries is at most the number of distinct keys.
func (b *Backend) GetProof(address common.Address, storageKeys []string, blockNrOrHash rpctypes.BlockNumberOrHash) (*rpctypes.AccountResult, error) {
	if c := b.RPCGetProofStorageKeysCap(); c > 0 && len(storageKeys) > int(c) {
		return nil, fmt.Errorf("max number of storage keys exceeded (max allowed %v)", c)
	}

	keys := make([]common.Hash, len(storageKeys))

	for i, storageKey := range storageKeys {
		key, err := decodeStorageKey(storageKey)
		if err != nil {
			return nil, fmt.Errorf("invalid storage key at index %d: %w", i, err)
		}

		keys[i] = key
	}

	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	height := blockNum.Int64()

	_, err = b.TendermintBlockByNumber(blockNum)
	if err != nil {
		// the error message imitates geth behavior
		return nil, errors.New("header not found")
	}
	ctx := rpctypes.ContextWithHeight(height)

	// if the height is equal to zero, meaning the query condition of the block is either "pending" or "latest"
	if height == 0 {
		bn, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}

		if bn > math.MaxInt64 {
			return nil, fmt.Errorf("not able to query block number greater than MaxInt64")
		}

		height = int64(bn) //nolint:gosec // checked for int overflow already
	}

	clientCtx := b.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]rpctypes.StorageResult, len(storageKeys))

	// A proof is a pure function of the store, the decoded key and the height,
	// all fixed here, so the same key repeated in one request is queried once.
	// Each entry gets its own value. The proof slice is shared; nothing
	// downstream writes to it, it is only serialized into the response.
	type storageProof struct {
		valueBz []byte
		proof   []string
	}
	queried := make(map[common.Hash]storageProof, len(keys))

	for i, key := range keys {
		resolved, ok := queried[key]
		if !ok {
			valueBz, proof, err := b.queryClient.GetProof(clientCtx, evmtypes.StoreKey, evmtypes.StateKey(address, key.Bytes()))
			if err != nil {
				return nil, err
			}

			resolved = storageProof{
				valueBz: valueBz,
				proof:   GetHexProofs(proof),
			}
			queried[key] = resolved
		}

		storageProofs[i] = rpctypes.StorageResult{
			Key:   storageKeys[i],
			Value: (*hexutil.Big)(new(big.Int).SetBytes(resolved.valueBz)),
			Proof: resolved.proof,
		}
	}

	// query EVM account
	req := &evmtypes.QueryAccountRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Account(ctx, req)
	if err != nil {
		return nil, err
	}

	// query account proofs
	//nolint:gocritic
	accountKey := append(authtypes.AddressStoreKeyPrefix, sdk.AccAddress(address.Bytes())...)
	_, proof, err := b.queryClient.GetProof(clientCtx, authtypes.StoreKey, accountKey)
	if err != nil {
		return nil, err
	}

	balance, ok := sdkmath.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return &rpctypes.AccountResult{
		Address:      address,
		AccountProof: GetHexProofs(proof),
		Balance:      (*hexutil.Big)(balance.BigInt()),
		CodeHash:     common.HexToHash(res.CodeHash),
		Nonce:        hexutil.Uint64(res.Nonce),
		StorageHash:  common.Hash{}, // NOTE: Mezo doesn't have a storage hash. TODO: implement?
		StorageProof: storageProofs,
	}, nil
}

// decodeStorageKey parses an `eth_getProof` storage key into the key it denotes.
// A key shorter than 32 bytes is left-padded with zeros; an odd number of hex
// digits is left-padded with one zero digit. Input that is not hex, or longer
// than 32 bytes, is rejected.
//
// The accepted input follows decodeHash from go-ethereum's internal/ethapi
// package, which cannot be imported.
func decodeStorageKey(s string) (common.Hash, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}

	if (len(s) & 1) > 0 {
		s = "0" + s
	}

	if len(s) > 64 {
		return common.Hash{}, errors.New("hex string too long, want at most 32 bytes")
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return common.Hash{}, errors.New("hex string invalid")
	}

	return common.BytesToHash(b), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (b *Backend) GetStorageAt(address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryStorageRequest{
		Address: address.String(),
		Key:     key,
	}

	res, err := b.queryClient.Storage(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(res.Value)
	return value.Bytes(), nil
}

// GetBalance returns the provided account's balance up to the provided block number.
func (b *Backend) GetBalance(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryBalanceRequest{
		Address: address.String(),
	}

	_, err = b.TendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	res, err := b.queryClient.Balance(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	val, ok := sdkmath.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	// balance can only be negative in case of pruned node
	if val.IsNegative() {
		return nil, errors.New("couldn't fetch balance. Node state is pruned")
	}

	return (*hexutil.Big)(val.BigInt()), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (b *Backend) GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	n := hexutil.Uint64(0)
	bn, err := b.BlockNumber()
	if err != nil {
		return &n, err
	}
	height := blockNum.Int64()

	currentHeight := int64(bn) //nolint:gosec // checked for int overflow already
	if height > currentHeight {
		return &n, errorsmod.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"cannot query with height in the future (current: %d, queried: %d); please provide a valid height",
			currentHeight, height,
		)
	}
	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := b.clientCtx.AccountRetriever

	err = accRet.EnsureExists(b.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		return &n, nil
	}

	includePending := blockNum == rpctypes.EthPendingBlockNumber
	nonce, err := b.getAccountNonce(address, includePending, blockNum.Int64(), b.logger)
	if err != nil {
		return nil, err
	}

	n = hexutil.Uint64(nonce)
	return &n, nil
}
