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
	"fmt"

	"github.com/mezo-org/mezod/x/feemarket/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock updates base fee
func (k *Keeper) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	baseFee := k.CalculateBaseFee(sdkCtx)

	// return immediately if base fee is nil
	if baseFee == nil {
		return nil
	}

	k.SetBaseFee(sdkCtx, baseFee)

	defer func() {
		telemetry.SetGauge(float32(baseFee.Int64()), "feemarket", "base_fee")
	}()

	// Store current base fee in event
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeFeeMarket,
			sdk.NewAttribute(types.AttributeKeyBaseFee, baseFee.String()),
		),
	})

	return nil
}

// EndBlock update block gas wanted.
func (k *Keeper) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if sdkCtx.BlockGasMeter() == nil {
		k.Logger(sdkCtx).Error("block gas meter is nil when setting block gas wanted")
		return nil
	}

	gasWanted := sdkmath.NewIntFromUint64(k.GetTransientGasWanted(sdkCtx))
	gasUsed := sdkmath.NewIntFromUint64(sdkCtx.BlockGasMeter().GasConsumedToLimit())

	if !gasWanted.IsInt64() {
		k.Logger(sdkCtx).Error("integer overflow by integer type conversion. Gas wanted > MaxInt64", "gas wanted", gasWanted.String())
		return nil
	}

	if !gasUsed.IsInt64() {
		k.Logger(sdkCtx).Error("integer overflow by integer type conversion. Gas used > MaxInt64", "gas used", gasUsed.String())
		return nil
	}

	// to prevent BaseFee manipulation we limit the gasWanted so that
	// gasWanted = max(gasWanted * MinGasMultiplier, gasUsed)
	// this will be keep BaseFee protected from un-penalized manipulation
	// more info here https://github.com/evmos/ethermint/pull/1105#discussion_r888798925
	minGasMultiplier := k.GetParams(sdkCtx).MinGasMultiplier
	limitedGasWanted := sdkmath.LegacyNewDec(gasWanted.Int64()).Mul(minGasMultiplier)
	updatedGasWanted := sdkmath.LegacyMaxDec(limitedGasWanted, sdkmath.LegacyNewDec(gasUsed.Int64())).TruncateInt().Uint64()
	k.SetBlockGasWanted(sdkCtx, updatedGasWanted)

	defer func() {
		telemetry.SetGauge(float32(updatedGasWanted), "feemarket", "block_gas")
	}()

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"block_gas",
		sdk.NewAttribute("height", fmt.Sprintf("%d", sdkCtx.BlockHeight())),
		sdk.NewAttribute("amount", fmt.Sprintf("%d", updatedGasWanted)),
	))

	return nil
}
