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
package testutil

import (
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/encoding"
	"github.com/mezo-org/mezod/testutil/tx"
)

// Commit commits a block at a given time. Reminder: At the end of each
// Tendermint Consensus round the following methods are run
//  1. BeginBlock
//  2. DeliverTx
//  3. EndBlock
//  4. Commit
func Commit(ctx sdk.Context, app *app.Mezo, t time.Duration, vs *tmtypes.ValidatorSet) (sdk.Context, error) {
	header := ctx.BlockHeader()

	if vs != nil {
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: header.Height})
		if err != nil {
			return ctx, err
		}

		nextVals, err := applyValSetChanges(vs, res.ValidatorUpdates)
		if err != nil {
			return ctx, err
		}
		header.ValidatorsHash = vs.Hash()
		header.NextValidatorsHash = nextVals.Hash()
	} else {
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: header.Height})
		if err != nil {
			return ctx, err
		}
	}

	_, err := app.Commit()
	if err != nil {
		return ctx, err
	}

	header.Height++
	header.Time = header.Time.Add(t)
	header.AppHash = app.LastCommitID().Hash

	// After commit, the finalizeBlockState is set to nil and NewContextLegacy
	// will panic in that case. We need to simulate the behavior of the
	// actual application and call ProcessProposal to set the finalizeBlockState
	// for the new block before creating a new context.
	_, err = app.ProcessProposal(&abci.RequestProcessProposal{Height: header.Height})
	if err != nil {
		return ctx, err
	}

	return app.BaseApp.NewContextLegacy(false, header), nil
}

// DeliverTx delivers a cosmos tx for a given set of msgs
func DeliverTx(
	ctx sdk.Context,
	appMezo *app.Mezo,
	priv cryptotypes.PrivKey,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (*abci.ExecTxResult, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
	tx, err := tx.PrepareCosmosTx(
		ctx,
		appMezo,
		tx.CosmosTxArgs{
			TxCfg:    txConfig,
			Priv:     priv,
			ChainID:  ctx.ChainID(),
			Gas:      10_000_000,
			GasPrice: gasPrice,
			Msgs:     msgs,
		},
	)
	if err != nil {
		return nil, err
	}
	return BroadcastTxBytes(
		appMezo,
		txConfig.TxEncoder(),
		tx,
		ctx.BlockHeight(),
		nil, // proposer is not processed in Cosmos transactions.
	)
}

// DeliverEthTx generates and broadcasts a Cosmos Tx populated with MsgEthereumTx messages.
// If a private key is provided, it will attempt to sign all messages with the given private key,
// otherwise, it will assume the messages have already been signed.
func DeliverEthTx(
	ctx sdk.Context,
	appMezo *app.Mezo,
	proposer sdk.ConsAddress,
	priv cryptotypes.PrivKey,
	msgs ...sdk.Msg,
) (*abci.ExecTxResult, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig

	tx, err := tx.PrepareEthTx(txConfig, appMezo, priv, msgs...)
	if err != nil {
		return nil, err
	}
	return BroadcastTxBytes(
		appMezo,
		txConfig.TxEncoder(),
		tx,
		ctx.BlockHeight(),
		proposer, // proposer must be set as x/evm check the coinbase address
	)
}

// CheckTx checks a cosmos tx for a given set of msgs
func CheckTx(
	ctx sdk.Context,
	appMezo *app.Mezo,
	priv cryptotypes.PrivKey,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (*abci.ResponseCheckTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig

	tx, err := tx.PrepareCosmosTx(
		ctx,
		appMezo,
		tx.CosmosTxArgs{
			TxCfg:    txConfig,
			Priv:     priv,
			ChainID:  ctx.ChainID(),
			GasPrice: gasPrice,
			Gas:      10_000_000,
			Msgs:     msgs,
		},
	)
	if err != nil {
		return nil, err
	}
	return checkTxBytes(appMezo, txConfig.TxEncoder(), tx)
}

// CheckEthTx checks a Ethereum tx for a given set of msgs
func CheckEthTx(
	appMezo *app.Mezo,
	priv cryptotypes.PrivKey,
	msgs ...sdk.Msg,
) (*abci.ResponseCheckTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig

	tx, err := tx.PrepareEthTx(txConfig, appMezo, priv, msgs...)
	if err != nil {
		return nil, err
	}
	return checkTxBytes(appMezo, txConfig.TxEncoder(), tx)
}

// BroadcastTxBytes encodes a transaction and calls DeliverTx on the app.
func BroadcastTxBytes(
	app *app.Mezo,
	txEncoder sdk.TxEncoder,
	tx sdk.Tx,
	blockHeight int64,
	proposer sdk.ConsAddress,
) (*abci.ExecTxResult, error) {
	// bz are bytes to be broadcasted over the network
	bz, err := txEncoder(tx)
	if err != nil {
		return nil, err
	}

	req := &abci.RequestFinalizeBlock{
		Height: blockHeight,
		Txs: [][]byte{bz},
		ProposerAddress: proposer,
	}
	res, err := app.BaseApp.FinalizeBlock(req)
	if err != nil {
		return nil, errortypes.ErrInvalidRequest
	}

	return res.TxResults[0], nil
}

// checkTxBytes encodes a transaction and calls checkTx on the app.
func checkTxBytes(app *app.Mezo, txEncoder sdk.TxEncoder, tx sdk.Tx) (*abci.ResponseCheckTx, error) {
	bz, err := txEncoder(tx)
	if err != nil {
		return nil, err
	}

	req := &abci.RequestCheckTx{Tx: bz}
	res, err := app.BaseApp.CheckTx(req)
	if err != nil {
		return nil, errortypes.ErrInvalidRequest
	}

	return res, nil
}

// applyValSetChanges takes in tmtypes.ValidatorSet and []abci.ValidatorUpdate and will return a new tmtypes.ValidatorSet which has the
// provided validator updates applied to the provided validator set.
func applyValSetChanges(valSet *tmtypes.ValidatorSet, valUpdates []abci.ValidatorUpdate) (*tmtypes.ValidatorSet, error) {
	updates, err := tmtypes.PB2TM.ValidatorUpdates(valUpdates)
	if err != nil {
		return nil, err
	}

	// must copy since validator set will mutate with UpdateWithChangeSet
	newVals := valSet.Copy()
	err = newVals.UpdateWithChangeSet(updates)
	if err != nil {
		return nil, err
	}

	return newVals, nil
}
