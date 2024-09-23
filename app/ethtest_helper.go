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

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/mezo-org/mezod/utils"

	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
	poatypes "github.com/mezo-org/mezod/x/poa/types"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/encoding"
	ethsidecar "github.com/mezo-org/mezod/ethereum/sidecar"
)

// EthSetup initializes a new MezoApp. A Nop logger is set in MezoApp.
func EthSetup(isCheckTx bool, patchGenesis func(*Mezo, simapp.GenesisState) simapp.GenesisState) *Mezo {
	return EthSetupWithDB(isCheckTx, patchGenesis, dbm.NewMemDB())
}

// EthSetupWithDB initializes a new MezoApp. A Nop logger is set in MezoApp.
func EthSetupWithDB(isCheckTx bool, patchGenesis func(*Mezo, simapp.GenesisState) simapp.GenesisState, db dbm.DB) *Mezo {
	chainID := utils.TestnetChainID + "-1"
	app := NewMezo(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		5,
		encoding.MakeConfig(ModuleBasics),
		ethsidecar.NewClientMock(),
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(chainID),
	)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewTestGenesisState(app.AppCodec())
		if patchGenesis != nil {
			genesisState = patchGenesis(app, genesisState)
		}

		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		_, err = app.InitChain(
			&abci.RequestInitChain{
				ChainId:         chainID,
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
		if err != nil {
			panic(err)
		}
	}

	return app
}

// NewTestGenesisState generate genesis state with single validator
func NewTestGenesisState(codec codec.Codec) simapp.GenesisState {
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(err)
	}
	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}

	genesisState := NewDefaultGenesisState()
	return genesisStateWithValSet(
		codec,
		genesisState,
		acc.GetAddress(),
		valSet,
		[]authtypes.GenesisAccount{acc},
		balance,
	)
}

func genesisStateWithValSet(
	codec codec.Codec,
	genesisState simapp.GenesisState,
	owner sdk.AccAddress,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) simapp.GenesisState {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]poatypes.Validator, 0, len(valSet.Validators))

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		if err != nil {
			panic(err)
		}
		validator := poatypes.Validator{
			OperatorBech32: sdk.ValAddress(val.Address).String(),
			//nolint:staticcheck
			ConsPubKeyBech32: legacybech32.MustMarshalPubKey(
				legacybech32.ConsPK,
				pk,
			),
			Description: poatypes.Description{},
		}
		validators = append(validators, validator)
	}
	// set validators and delegations
	stakingGenesis := poatypes.NewGenesisState(
		poatypes.DefaultParams(),
		owner,
		validators,
	)
	genesisState[poatypes.ModuleName] = codec.MustMarshalJSON(&stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	return genesisState
}
