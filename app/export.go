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

package app

import (
	"encoding/json"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/simapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/mezo-org/mezod/encoding"
)

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() simapp.GenesisState {
	encCfg := encoding.MakeConfig(ModuleBasics)
	return ModuleBasics.DefaultGenesis(encCfg.Codec)
}

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *Mezo) ExportAppStateAndValidators(
	forZeroHeight bool, _ []string,
) (servertypes.ExportedApp, error) {
	// Creates context with current height and checks txs for ctx to be usable by start of next block
	ctx := app.NewContextLegacy(true, tmproto.Header{Height: app.LastBlockHeight()})

	// We export at last height + 1, because that's the height at which
	// Tendermint will start InitChain.
	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		panic("for-zero-height export is not supported")
	}

	genState, err := app.mm.ExportGenesis(ctx, app.appCodec)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := app.PoaKeeper.ExportGenesisValidators(ctx)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, nil
}
