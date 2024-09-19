package keeper_test

import (
	_ "embed"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mezotypes "github.com/mezo-org/mezod/types"
	"github.com/mezo-org/mezod/x/evm/keeper"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func (suite *KeeperTestSuite) TestWithChainID() {
	testCases := []struct {
		name       string
		chainID    string
		expChainID int64
		expPanic   bool
	}{
		{
			"fail - chainID is empty",
			"",
			0,
			true,
		},
		{
			"fail - other chainID",
			"chain_7701-1",
			0,
			true,
		},
		{
			"success - Mezo mainnet chain ID",
			"mezo_31612-1",
			31612,
			false,
		},
		{
			"success - Mezo testnet chain ID",
			"mezo_31611-1",
			31611,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			keeper := keeper.Keeper{}
			ctx := suite.ctx.WithChainID(tc.chainID)

			if tc.expPanic {
				suite.Require().Panics(func() {
					keeper.WithChainID(ctx)
				})
			} else {
				suite.Require().NotPanics(func() {
					keeper.WithChainID(ctx)
					suite.Require().Equal(tc.expChainID, keeper.ChainID().Int64())
				})
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBaseFee() {
	testCases := []struct {
		name            string
		enableLondonHF  bool
		enableFeemarket bool
		expectBaseFee   *big.Int
	}{
		{"not enable london HF, not enable feemarket", false, false, nil},
		{"enable london HF, not enable feemarket", true, false, big.NewInt(0)},
		{"enable london HF, enable feemarket", true, true, big.NewInt(1000000000)},
		{"not enable london HF, enable feemarket", false, true, nil},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest()
			err := suite.app.EvmKeeper.BeginBlock(suite.ctx)
			suite.Require().NoError(err)
			params := suite.app.EvmKeeper.GetParams(suite.ctx)
			ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
			baseFee := suite.app.EvmKeeper.GetBaseFee(suite.ctx, ethCfg)
			suite.Require().Equal(tc.expectBaseFee, baseFee)
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite *KeeperTestSuite) TestGetAccountStorage() {
	testCases := []struct {
		name     string
		malleate func()
		expRes   []int
	}{
		{
			"Only one account that's not a contract (no storage)",
			func() {},
			[]int{0},
		},
		{
			"Two accounts - one contract (with storage), one wallet",
			func() {
				supply := big.NewInt(100)
				suite.DeployTestContract(suite.T(), suite.address, supply)
			},
			[]int{2, 0},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()
			i := 0
			// IterateAccounts also iterates through custom precompile accounts
			// as they get added at genesis, so we need to append the expected
			// storage length of each precompile to expRes to avoid an out of
			// index error during the test
			precompileGenesisAccounts := suite.app.EvmKeeper.CustomPrecompileGenesisAccounts()
			for _, pga := range precompileGenesisAccounts {
				tc.expRes = append(tc.expRes, len(pga.Storage))
			}

			suite.app.AccountKeeper.IterateAccounts(suite.ctx, func(account sdk.AccountI) bool {
				ethAccount, ok := account.(mezotypes.EthAccountI)
				if !ok {
					// ignore non EthAccounts
					return false
				}
				addr := ethAccount.EthAddress()
				storage := suite.app.EvmKeeper.GetAccountStorage(suite.ctx, addr)
				suite.Require().Equal(tc.expRes[i], len(storage))
				i++
				return false
			})
		})
	}
}

func (suite *KeeperTestSuite) TestGetAccountOrEmpty() {
	empty := statedb.Account{
		Balance:  new(uint256.Int),
		CodeHash: evmtypes.EmptyCodeHash,
	}

	supply := big.NewInt(100)
	contractAddr := suite.DeployTestContract(suite.T(), suite.address, supply)

	testCases := []struct {
		name     string
		addr     common.Address
		expEmpty bool
	}{
		{
			"unexisting account - get empty",
			common.Address{},
			true,
		},
		{
			"existing contract account",
			contractAddr,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			res := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, tc.addr)
			if tc.expEmpty {
				suite.Require().Equal(empty, res)
			} else {
				suite.Require().NotEqual(empty, res)
			}
		})
	}
}
