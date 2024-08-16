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
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/cometbft/cometbft/crypto/ed25519"
	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
	poatypes "github.com/mezo-org/mezod/x/poa/types"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/mezo-org/mezod/encoding"
	feemarkettypes "github.com/mezo-org/mezod/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/utils"
)

func init() {
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)
	config.SetBip44CoinType(cfg)
}

// DefaultConsensusParams defines the default Tendermint consensus params used in
// Mezo testing.
var DefaultConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   10000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

func init() {
	feemarkettypes.DefaultMinGasPrice = sdkmath.LegacyZeroDec()
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)
	config.SetBip44CoinType(cfg)
}

// Setup initializes a new Mezo. A Nop logger is set in Mezo.
func Setup(
	isCheckTx bool,
	feemarketGenesis *feemarkettypes.GenesisState,
) *Mezo {
	privVal := ed25519.GenPrivKey()
	pubKey := privVal.PubKey()

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdkmath.NewInt(100000000000000))),
	}

	db := dbm.NewMemDB()

	chainID := utils.MainnetChainID + "-1"

	app := NewMezo(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		5,
		encoding.MakeConfig(ModuleBasics),
		newAppOptions(DefaultNodeHome, chainID),
		baseapp.SetChainID(chainID),
	)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState()

		genesisState = GenesisStateWithValSet(
			app,
			genesisState,
			acc.GetAddress(),
			valSet,
			[]authtypes.GenesisAccount{acc},
			balance,
		)

		// Verify feeMarket genesis
		if feemarketGenesis != nil {
			if err := feemarketGenesis.Validate(); err != nil {
				panic(err)
			}
			genesisState[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
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

func newAppOptions(homePath, chainID string) servertypes.AppOptions {
	return simtestutil.AppOptionsMap{
		flags.FlagHome:    homePath,
		flags.FlagChainID: chainID,
	}
}

func GenesisStateWithValSet(
	app *Mezo,
	genesisState simapp.GenesisState,
	owner sdk.AccAddress,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) simapp.GenesisState {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]poatypes.Validator, 0, len(valSet.Validators))

	for _, val := range valSet.Validators {
		pk, _ := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
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
	poaParams := poatypes.DefaultParams()
	poaGenesis := poatypes.NewGenesisState(poaParams, owner, validators)
	genesisState[poatypes.ModuleName] = app.AppCodec().MustMarshalJSON(&poaGenesis)

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
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState
}
