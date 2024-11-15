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
package evm

import (
	"bytes"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/holiman/uint256"
	mezotypes "github.com/mezo-org/mezod/types"
	"github.com/mezo-org/mezod/x/evm/keeper"
	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(
	ctx sdk.Context,
	k *keeper.Keeper,
	accountKeeper types.AccountKeeper,
	data types.GenesisState,
) []abci.ValidatorUpdate {
	k.WithChainID(ctx)

	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(fmt.Errorf("error setting params %s", err))
	}

	// ensure evm module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the EVM module account has not been set")
	}

	// add custom precompile genesis accounts to genesis state
	customPrecompileGenesisAccounts := k.CustomPrecompileGenesisAccounts()
	data.Accounts = append(data.Accounts, customPrecompileGenesisAccounts...)

	for _, account := range data.Accounts {
		address := common.HexToAddress(account.Address)
		code := common.Hex2Bytes(account.Code)
		codeHash := crypto.Keccak256Hash(code)

		if k.IsCustomPrecompile(address) {
			err = k.SetAccount(ctx, address, statedb.Account{
				Nonce:    0,
				Balance:  uint256.NewInt(0),
				CodeHash: codeHash.Bytes(),
			})
			if err != nil {
				panic(fmt.Errorf("error setting precompile account %s", err))
			}
		} else {
			accAddress := sdk.AccAddress(address.Bytes())
			// check that the EVM balance matches the account balance
			acc := accountKeeper.GetAccount(ctx, accAddress)
			if acc == nil {
				panic(fmt.Errorf("account not found for address %s", account.Address))
			}

			ethAcct, ok := acc.(mezotypes.EthAccountI)
			if !ok {
				panic(
					fmt.Errorf("account %s must be an EthAccount interface, got %T",
						account.Address, acc,
					),
				)
			}

			// we ignore the empty Code hash checking, see ethermint PR#1234
			if len(account.Code) != 0 && !bytes.Equal(ethAcct.GetCodeHash().Bytes(), codeHash.Bytes()) {
				s := "the evm state code doesn't match with the codehash\n"
				panic(fmt.Sprintf("%s account: %s , evm state codehash: %v, ethAccount codehash: %v, evm state code: %s\n",
					s, account.Address, codeHash, ethAcct.GetCodeHash(), account.Code))

			}
		}

		k.SetCode(ctx, codeHash.Bytes(), code)

		for _, storage := range account.Storage {
			k.SetState(ctx, address, common.HexToHash(storage.Key), common.HexToHash(storage.Value).Bytes())
		}
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state of the EVM module
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper, ak types.AccountKeeper) *types.GenesisState {
	var ethGenAccounts []types.GenesisAccount
	ak.IterateAccounts(ctx, func(account sdk.AccountI) bool {
		ethAccount, ok := account.(mezotypes.EthAccountI)
		if !ok {
			// ignore non EthAccounts
			return false
		}

		addr := ethAccount.EthAddress()

		storage := k.GetAccountStorage(ctx, addr)

		genAccount := types.GenesisAccount{
			Address: addr.String(),
			Code:    common.Bytes2Hex(k.GetCode(ctx, ethAccount.GetCodeHash())),
			Storage: storage,
		}

		ethGenAccounts = append(ethGenAccounts, genAccount)
		return false
	})

	return &types.GenesisState{
		Accounts: ethGenAccounts,
		Params:   k.GetParams(ctx),
	}
}
