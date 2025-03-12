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
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	storetypes "cosmossdk.io/store/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BeginBlock sets the sdk Context and EIP155 chain id to the Keeper.
func (k *Keeper) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.WithChainID(sdkCtx)
	return nil
}

// EndBlock also retrieves the bloom filter value from the transient store and commits it to the
// KVStore. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Gas costs are handled within msg handler so costs should be ignored
	infCtx := sdkCtx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	bloom := ethtypes.BytesToBloom(k.GetBlockBloomTransient(infCtx).Bytes())
	k.EmitBlockBloomEvent(infCtx, bloom)

	transferFundsToChainFeeSplitter(k, sdkCtx)

	return nil
}

// Move entire balance of the fee collector module account to the chain fee splitter
// contract.
func transferFundsToChainFeeSplitter(k *Keeper, sdkCtx sdk.Context) {
	feeCollectorAddr := authtypes.NewModuleAddress(authtypes.FeeCollectorName)
	balance := k.bankKeeper.GetBalance(sdkCtx, feeCollectorAddr, k.GetParams(sdkCtx).EvmDenom)
	if balance.IsZero() {
		sdkCtx.Logger().Info("fee collector balance is zero, skipping transfer to chain fee splitter")

		return
	}

	chainFeeSplitterAddress := common.HexToAddress(k.GetParams(sdkCtx).ChainFeeSplitterAddress)

	// Check if the chain fee splitter address is the zero address.
	// In case of zero address, fees are still being collected in the fee collector
	// module account.
	if chainFeeSplitterAddress == (common.Address{}) {
		sdkCtx.Logger().Info("chain fee splitter address is zero, skipping transfer to chain fee splitter")
	
		return
	}

	chainFeeSplitterAddressBytes := chainFeeSplitterAddress.Bytes()

	// Check if the chain fee splitter address is valid format
	if err := sdk.VerifyAddressFormat(chainFeeSplitterAddressBytes); err != nil {
		sdkCtx.Logger().Error("invalid chain fee splitter address", "error", err)

		return
	}

	// Transfer chain fee to the chain fee splitter contract
	err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, authtypes.FeeCollectorName, chainFeeSplitterAddressBytes, sdk.NewCoins(balance))
	if err != nil {
		sdkCtx.Logger().Error("failed to send chain fee to the chain fee splitter address", "error", err)
	
		return
	}

	sdkCtx.Logger().Info("chain fee transferred to chain fee splitter contract", "amount", balance, "address", chainFeeSplitterAddress)
}
