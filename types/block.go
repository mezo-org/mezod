// Copyright 2022 Evmos Foundation
// This file is part of the Mezo Network packages.
//
// Mezo is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Mezo packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Mezo packages. If not, see https://github.com/mezo-org/mezod/blob/main/LICENSE
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BlockGasLimit returns the max gas (limit) defined in the block gas meter. If the meter is not
// set, it returns the max gas from the application consensus params.
// NOTE: see https://github.com/cosmos/cosmos-sdk/issues/9514 for full reference
func BlockGasLimit(ctx sdk.Context) uint64 {
	blockGasMeter := ctx.BlockGasMeter()

	// Get the limit from the gas meter only if its not null and not an InfiniteGasMeter
	if blockGasMeter != nil && blockGasMeter.Limit() != 0 {
		return blockGasMeter.Limit()
	}

	// Otherwise get from the consensus parameters
	cp := ctx.ConsensusParams()
	if cp.Block == nil {
		return 0
	}

	maxGas := cp.Block.MaxGas
	if maxGas > 0 {
		return uint64(maxGas) // #nosec G701 -- maxGas is int64 type. It can never be greater than math.MaxUint64
	}

	return 0
}
