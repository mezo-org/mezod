package keeper_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ethlogger "github.com/ethereum/go-ethereum/eth/tracers/logger"
	rpctypes "github.com/mezo-org/mezod/rpc/types"
	"github.com/mezo-org/mezod/server/config"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// Not valid Ethereum address
const invalidAddress = "0x0000"

func (suite *KeeperTestSuite) TestQueryAccount() {
	var (
		req        *types.QueryAccountRequest
		expAccount *types.QueryAccountResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid address",
			func() {
				expAccount = &types.QueryAccountResponse{
					Balance:  "0",
					CodeHash: common.BytesToHash(crypto.Keccak256(nil)).Hex(),
					Nonce:    0,
				}
				req = &types.QueryAccountRequest{
					Address: invalidAddress,
				}
			},
			false,
		},
		{
			"success",
			func() {
				amt := sdk.Coins{sdk.NewInt64Coin(types.DefaultEVMDenom, 100)}
				err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, amt)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, suite.address.Bytes(), amt)
				suite.Require().NoError(err)

				expAccount = &types.QueryAccountResponse{
					Balance:  "100",
					CodeHash: common.BytesToHash(crypto.Keccak256(nil)).Hex(),
					Nonce:    0,
				}
				req = &types.QueryAccountRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.Account(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expAccount, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCosmosAccount() {
	var (
		req        *types.QueryCosmosAccountRequest
		expAccount *types.QueryCosmosAccountResponse
	)

	// custom precompile accounts are added at genesis
	// offset expected account numbers by the number of precompiles at genesis
	//
	// 1 is added to the offset to account for the x/bridge module account
	// which is also created at genesis
	precompileOffset := uint64(len(suite.app.EvmKeeper.CustomPrecompileGenesisAccounts())) + 1

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid address",
			func() {
				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(common.Address{}.Bytes()).String(),
				}
				req = &types.QueryCosmosAccountRequest{
					Address: invalidAddress,
				}
			},
			false,
		},
		{
			"success",
			func() {
				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:      0,
					AccountNumber: precompileOffset + 2, // this is set during the test setup
				}
				req = &types.QueryCosmosAccountRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
		{
			"success with seq and account number",
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, suite.address.Bytes())
				suite.Require().NoError(acc.SetSequence(10))
				nextAccNumber := suite.app.AccountKeeper.NextAccountNumber(suite.ctx)
				suite.Require().NoError(acc.SetAccountNumber(nextAccNumber))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:      10,
					AccountNumber: precompileOffset + 3,
				}
				req = &types.QueryCosmosAccountRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.CosmosAccount(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expAccount, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryBalance() {
	var (
		req        *types.QueryBalanceRequest
		expBalance string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid address",
			func() {
				expBalance = "0"
				req = &types.QueryBalanceRequest{
					Address: invalidAddress,
				}
			},
			false,
		},
		{
			"success",
			func() {
				amt := sdk.Coins{sdk.NewInt64Coin(types.DefaultEVMDenom, 100)}
				err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, amt)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, suite.address.Bytes(), amt)
				suite.Require().NoError(err)

				expBalance = "100"
				req = &types.QueryBalanceRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.Balance(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expBalance, res.Balance)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryStorage() {
	var (
		req      *types.QueryStorageRequest
		expValue string
	)

	testCases := []struct {
		msg      string
		malleate func(vm.StateDB)
		expPass  bool
	}{
		{
			"invalid address",
			func(vm.StateDB) {
				req = &types.QueryStorageRequest{
					Address: invalidAddress,
				}
			},
			false,
		},
		{
			"success",
			func(vmdb vm.StateDB) {
				key := common.BytesToHash([]byte("key"))
				value := common.BytesToHash([]byte("value"))
				expValue = value.String()
				vmdb.SetState(suite.address, key, value)
				req = &types.QueryStorageRequest{
					Address: suite.address.String(),
					Key:     key.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			vmdb := suite.StateDB()
			tc.malleate(vmdb)
			suite.Require().NoError(vmdb.Commit())

			res, err := suite.queryClient.Storage(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expValue, res.Value)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCode() {
	var (
		req     *types.QueryCodeRequest
		expCode []byte
	)

	testCases := []struct {
		msg      string
		malleate func(vm.StateDB)
		expPass  bool
	}{
		{
			"invalid address",
			func(vm.StateDB) {
				req = &types.QueryCodeRequest{
					Address: invalidAddress,
				}
				exp := &types.QueryCodeResponse{}
				expCode = exp.Code
			},
			false,
		},
		{
			"success",
			func(vmdb vm.StateDB) {
				expCode = []byte("code")
				vmdb.SetCode(suite.address, expCode)

				req = &types.QueryCodeRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			vmdb := suite.StateDB()
			tc.malleate(vmdb)
			suite.Require().NoError(vmdb.Commit())

			res, err := suite.queryClient.Code(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expCode, res.Code)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryTxLogs() {
	var expLogs []*types.Log
	txHash := common.BytesToHash([]byte("tx_hash"))
	txIndex := uint(1)
	logIndex := uint(1)

	testCases := []struct {
		msg      string
		malleate func(vm.StateDB)
	}{
		{
			"empty logs",
			func(vm.StateDB) {
				expLogs = nil
			},
		},
		{
			"success",
			func(vmdb vm.StateDB) {
				expLogs = []*types.Log{
					{
						Address:     suite.address.String(),
						Topics:      []string{common.BytesToHash([]byte("topic")).String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      txHash.String(),
						TxIndex:     uint64(txIndex),
						BlockHash:   common.BytesToHash(suite.ctx.HeaderHash()).Hex(),
						Index:       uint64(logIndex),
						Removed:     false,
					},
				}

				for _, log := range types.LogsToEthereum(expLogs) {
					vmdb.AddLog(log)
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			vmdb := statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewTxConfig(common.BytesToHash(suite.ctx.HeaderHash()), txHash, txIndex, logIndex))
			tc.malleate(vmdb)
			suite.Require().NoError(vmdb.Commit())

			logs := vmdb.Logs()
			suite.Require().Equal(expLogs, types.NewLogsFromEth(logs))
		})
	}
}

func (suite *KeeperTestSuite) TestQueryParams() {
	expParams := types.DefaultParams()

	res, err := suite.queryClient.Params(suite.ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

func (suite *KeeperTestSuite) TestQueryValidatorAccount() {
	var (
		req        *types.QueryValidatorAccountRequest
		expAccount *types.QueryValidatorAccountResponse
	)

	// custom precompile accounts are added at genesis
	// offset expected account numbers by the number of precompiles at genesis
	//
	// 1 is added to the offset to account for the x/bridge module account
	// which is also created at genesis
	precompileOffset := uint64(len(suite.app.EvmKeeper.CustomPrecompileGenesisAccounts())) + 1

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid address",
			func() {
				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(common.Address{}.Bytes()).String(),
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: "",
				}
			},
			false,
		},
		{
			"success",
			func() {
				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:       0,
					AccountNumber:  precompileOffset + 2, // this is set during the test setup
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: suite.consAddress.String(),
				}
			},
			true,
		},
		{
			"success with seq and account number",
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, suite.address.Bytes())
				suite.Require().NoError(acc.SetSequence(10))
				nextAccNumber := suite.app.AccountKeeper.NextAccountNumber(suite.ctx)
				suite.Require().NoError(acc.SetAccountNumber(nextAccNumber))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:       10,
					AccountNumber:  precompileOffset + 3,
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: suite.consAddress.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			res, err := suite.queryClient.ValidatorAccount(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expAccount, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateGas() {
	gasHelper := hexutil.Uint64(20000)
	higherGas := hexutil.Uint64(25000)
	hexBigInt := hexutil.Big(*big.NewInt(1))

	var (
		args   interface{}
		gasCap uint64
	)
	testCases := []struct {
		msg             string
		malleate        func()
		expPass         bool
		expGas          uint64
		enableFeemarket bool
	}{
		// should success, because transfer value is zero
		{
			"default args - special case for ErrIntrinsicGas on contract creation, raise gas limit",
			func() {
				args = types.TransactionArgs{}
			},
			true,
			ethparams.TxGasContractCreation,
			false,
		},
		// should success, because transfer value is zero
		{
			"default args with 'to' address",
			func() {
				args = types.TransactionArgs{To: &common.Address{}}
			},
			true,
			ethparams.TxGas,
			false,
		},
		// should fail, because the default From address(zero address) don't have fund
		{
			"not enough balance",
			func() {
				args = types.TransactionArgs{To: &common.Address{}, Value: (*hexutil.Big)(big.NewInt(100))}
			},
			false,
			0,
			false,
		},
		// should success, enough balance now
		{
			"enough balance",
			func() {
				args = types.TransactionArgs{To: &common.Address{}, From: &suite.address, Value: (*hexutil.Big)(big.NewInt(100))}
			}, false, 0, false,
		},
		// should success, because gas limit lower than 21000 is ignored
		{
			"gas exceed allowance",
			func() {
				args = types.TransactionArgs{To: &common.Address{}, Gas: &gasHelper}
			},
			true,
			ethparams.TxGas,
			false,
		},
		// should fail, invalid gas cap
		{
			"gas exceed global allowance",
			func() {
				args = types.TransactionArgs{To: &common.Address{}}
				gasCap = 20000
			},
			false,
			0,
			false,
		},
		// estimate gas of an erc20 contract deployment, the exact gas number is checked with geth
		{
			"contract deployment",
			func() {
				ctorArgs, err := types.ERC20Contract.ABI.Pack("", &suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Require().NoError(err)

				data := types.ERC20Contract.Bin
				data = append(data, ctorArgs...)
				args = types.TransactionArgs{
					From: &suite.address,
					Data: (*hexutil.Bytes)(&data),
				}
			},
			true,
			1187108,
			false,
		},
		// estimate gas of an erc20 transfer, the exact gas number is checked with geth
		{
			"erc20 transfer",
			func() {
				contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Commit()
				transferData, err := types.ERC20Contract.ABI.Pack("transfer", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
				suite.Require().NoError(err)
				args = types.TransactionArgs{To: &contractAddr, From: &suite.address, Data: (*hexutil.Bytes)(&transferData)}
			},
			true,
			51880,
			false,
		},
		// repeated tests with enableFeemarket
		{
			"default args w/ enableFeemarket",
			func() {
				args = types.TransactionArgs{To: &common.Address{}}
			},
			true,
			ethparams.TxGas,
			true,
		},
		{
			"not enough balance w/ enableFeemarket",
			func() {
				args = types.TransactionArgs{To: &common.Address{}, Value: (*hexutil.Big)(big.NewInt(100))}
			},
			false,
			0,
			true,
		},
		{
			"enough balance w/ enableFeemarket",
			func() {
				args = types.TransactionArgs{To: &common.Address{}, From: &suite.address, Value: (*hexutil.Big)(big.NewInt(100))}
			},
			false,
			0,
			true,
		},
		{
			"gas exceed allowance w/ enableFeemarket",
			func() {
				args = types.TransactionArgs{To: &common.Address{}, Gas: &gasHelper}
			},
			true,
			ethparams.TxGas,
			true,
		},
		{
			"gas exceed global allowance w/ enableFeemarket",
			func() {
				args = types.TransactionArgs{To: &common.Address{}}
				gasCap = 20000
			},
			false,
			0,
			true,
		},
		{
			"contract deployment w/ enableFeemarket",
			func() {
				ctorArgs, err := types.ERC20Contract.ABI.Pack("", &suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Require().NoError(err)
				data := types.ERC20Contract.Bin
				data = append(data, ctorArgs...)
				args = types.TransactionArgs{
					From: &suite.address,
					Data: (*hexutil.Bytes)(&data),
				}
			},
			true,
			1187108,
			true,
		},
		{
			"erc20 transfer w/ enableFeemarket",
			func() {
				contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Commit()
				transferData, err := types.ERC20Contract.ABI.Pack("transfer", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
				suite.Require().NoError(err)
				args = types.TransactionArgs{To: &contractAddr, From: &suite.address, Data: (*hexutil.Bytes)(&transferData)}
			},
			true,
			51880,
			true,
		},
		{
			"contract creation but 'create' param disabled",
			func() {
				ctorArgs, err := types.ERC20Contract.ABI.Pack("", &suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Require().NoError(err)
				data := types.ERC20Contract.Bin
				data = append(data, ctorArgs...)
				args = types.TransactionArgs{
					From: &suite.address,
					Data: (*hexutil.Bytes)(&data),
				}
				params := suite.app.EvmKeeper.GetParams(suite.ctx)
				params.EnableCreate = false
				err = suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			false,
			0,
			false,
		},
		{
			"specified gas in args higher than ethparams.TxGas (21,000)",
			func() {
				args = types.TransactionArgs{
					To:  &common.Address{},
					Gas: &higherGas,
				}
			},
			true,
			ethparams.TxGas,
			false,
		},
		{
			"specified gas in args higher than request gasCap",
			func() {
				gasCap = 22_000
				args = types.TransactionArgs{
					To:  &common.Address{},
					Gas: &higherGas,
				}
			},
			true,
			ethparams.TxGas,
			false,
		},
		{
			"invalid args - specified both gasPrice and maxFeePerGas",
			func() {
				args = types.TransactionArgs{
					To:           &common.Address{},
					GasPrice:     &hexBigInt,
					MaxFeePerGas: &hexBigInt,
				}
			},
			false,
			0,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.SetupTest()
			gasCap = 25_000_000
			tc.malleate()

			args, err := json.Marshal(&args)
			suite.Require().NoError(err)
			req := types.EthCallRequest{
				Args:            args,
				GasCap:          gasCap,
				ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
			}

			rsp, err := suite.queryClient.EstimateGas(suite.ctx, &req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(int64(tc.expGas), int64(rsp.Gas)) //nolint:gosec
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.enableFeemarket = false // reset flag
}

func (suite *KeeperTestSuite) TestTraceTx() {
	// TODO deploy contract that triggers internal transactions
	var (
		txMsg        *types.MsgEthereumTx
		traceConfig  *types.TraceConfig
		predecessors []*types.MsgEthereumTx
		chainID      *sdkmath.Int
	)

	testCases := []struct {
		msg             string
		malleate        func()
		expPass         bool
		traceResponse   string
		enableFeemarket bool
	}{
		{
			msg: "default trace",
			malleate: func() {
				traceConfig = nil
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:       true,
			traceResponse: "{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
		},
		{
			msg: "default trace with filtered response",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:         true,
			traceResponse:   "{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
			enableFeemarket: false,
		},
		{
			msg: "javascript tracer",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:       true,
			traceResponse: "[]",
		},
		{
			msg: "default trace with enableFeemarket",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:         true,
			traceResponse:   "{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
			enableFeemarket: true,
		},
		{
			msg: "javascript tracer with enableFeemarket",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:         true,
			traceResponse:   "[]",
			enableFeemarket: true,
		},
		{
			msg: "default tracer with predecessors",
			malleate: func() {
				traceConfig = nil

				// increase nonce to avoid address collision
				vmdb := suite.StateDB()
				vmdb.SetNonce(suite.address, vmdb.GetNonce(suite.address)+1)
				suite.Require().NoError(vmdb.Commit())

				contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Commit()
				// Generate token transfer transaction
				firstTx := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
				txMsg = suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
				suite.Commit()

				predecessors = append(predecessors, firstTx)
			},
			expPass:         true,
			traceResponse:   "{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
			enableFeemarket: false,
		},
		{
			msg: "invalid trace config - Negative Limit",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Limit:          -1,
				}
			},
			expPass: false,
		},
		{
			msg: "invalid trace config - Invalid Tracer",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Tracer:         "invalid_tracer",
				}
			},
			expPass: false,
		},
		{
			msg: "invalid trace config - Invalid Timeout",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Timeout:        "wrong_time",
				}
			},
			expPass: false,
		},
		{
			msg: "default tracer with contract creation tx as predecessor but 'create' param disabled",
			malleate: func() {
				traceConfig = nil

				// increase nonce to avoid address collision
				vmdb := suite.StateDB()
				vmdb.SetNonce(suite.address, vmdb.GetNonce(suite.address)+1)
				suite.Require().NoError(vmdb.Commit())

				chainID := suite.app.EvmKeeper.ChainID()
				nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)
				data := types.ERC20Contract.Bin
				ethTxParams := &types.EvmTxArgs{
					ChainID:  chainID,
					Nonce:    nonce,
					GasLimit: ethparams.TxGasContractCreation,
					Input:    data,
				}
				contractTx := types.NewTx(ethTxParams)

				predecessors = append(predecessors, contractTx)
				suite.Commit()

				params := suite.app.EvmKeeper.GetParams(suite.ctx)
				params.EnableCreate = false
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			expPass:       true,
			traceResponse: "{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PUSH1\",\"gas\":",
		},
		{
			msg: "invalid chain id",
			malleate: func() {
				traceConfig = nil
				predecessors = []*types.MsgEthereumTx{}
				tmp := sdkmath.NewInt(1)
				chainID = &tmp
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.SetupTest()
			// Deploy contract
			contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
			suite.Commit()
			// Generate token transfer transaction
			txMsg = suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
			suite.Commit()

			tc.malleate()
			traceReq := types.QueryTraceTxRequest{
				Msg:          txMsg,
				TraceConfig:  traceConfig,
				Predecessors: predecessors,
			}

			if chainID != nil {
				traceReq.ChainId = chainID.Int64()
			}
			res, err := suite.queryClient.TraceTx(suite.ctx, &traceReq)

			if tc.expPass {
				suite.Require().NoError(err)
				// if data is to big, slice the result
				if len(res.Data) > 150 {
					suite.Require().Equal(tc.traceResponse, string(res.Data[:150]))
				} else {
					suite.Require().Equal(tc.traceResponse, string(res.Data))
				}
				if traceConfig == nil || traceConfig.Tracer == "" {
					var result ethlogger.ExecutionResult
					suite.Require().NoError(json.Unmarshal(res.Data, &result))
					suite.Require().Positive(result.Gas)
				}
			} else {
				suite.Require().Error(err)
			}
			// Reset for next test case
			chainID = nil
		})
	}

	suite.enableFeemarket = false // reset flag
}

func (suite *KeeperTestSuite) TestTraceBlock() {
	var (
		txs         []*types.MsgEthereumTx
		traceConfig *types.TraceConfig
		chainID     *sdkmath.Int
	)

	testCases := []struct {
		msg             string
		malleate        func()
		expPass         bool
		traceResponse   string
		enableFeemarket bool
	}{
		{
			msg: "default trace",
			malleate: func() {
				traceConfig = nil
			},
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
		},
		{
			msg: "filtered trace",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
			},
			expPass:       true,
			traceResponse: "[{\"result\":{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
		},
		{
			msg: "javascript tracer",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
			},
			expPass:       true,
			traceResponse: "[{\"result\":[]}]",
		},
		{
			msg: "default trace with enableFeemarket and filtered return",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
			},
			expPass:         true,
			traceResponse:   "[{\"result\":{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
			enableFeemarket: true,
		},
		{
			msg: "javascript tracer with enableFeemarket",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
			},
			expPass:         true,
			traceResponse:   "[{\"result\":[]}]",
			enableFeemarket: true,
		},
		{
			msg: "tracer with multiple transactions",
			malleate: func() {
				traceConfig = nil

				// increase nonce to avoid address collision
				vmdb := suite.StateDB()
				vmdb.SetNonce(suite.address, vmdb.GetNonce(suite.address)+1)
				suite.Require().NoError(vmdb.Commit())

				contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
				suite.Commit()
				// create multiple transactions in the same block
				firstTx := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
				secondTx := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
				suite.Commit()
				// overwrite txs to include only the ones on new block
				txs = append([]*types.MsgEthereumTx{}, firstTx, secondTx)
			},
			expPass:         true,
			traceResponse:   "[{\"result\":{\"gas\":34828,\"failed\":false,\"returnValue\":\"0000000000000000000000000000000000000000000000000000000000000001\",\"structLogs\":[{\"pc\":0,\"op\":\"PU",
			enableFeemarket: false,
		},
		{
			msg: "invalid trace config - Negative Limit",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Limit:          -1,
				}
			},
			expPass: false,
		},
		{
			msg: "invalid trace config - Invalid Tracer",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
					Tracer:         "invalid_tracer",
				}
			},
			expPass:       true,
			traceResponse: "[{\"error\":\"rpc error: code = Internal desc = ReferenceError: invalid_tracer is not defined at \\u003ceval\\u003e:1:2(0)\"}]",
		},
		{
			msg: "invalid chain id",
			malleate: func() {
				traceConfig = nil
				tmp := sdkmath.NewInt(1)
				chainID = &tmp
			},
			expPass:       true,
			traceResponse: "[{\"error\":\"rpc error: code = Internal desc = invalid chain id for signer: have 31611 want 1\"}]",
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			txs = []*types.MsgEthereumTx{}
			suite.enableFeemarket = tc.enableFeemarket
			suite.SetupTest()
			// Deploy contract
			contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
			suite.Commit()
			// Generate token transfer transaction
			txMsg := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdkmath.NewIntWithDecimal(1, 18).BigInt())
			suite.Commit()

			txs = append(txs, txMsg)

			tc.malleate()
			traceReq := types.QueryTraceBlockRequest{
				Txs:         txs,
				TraceConfig: traceConfig,
			}

			if chainID != nil {
				traceReq.ChainId = chainID.Int64()
			}

			res, err := suite.queryClient.TraceBlock(suite.ctx, &traceReq)

			if tc.expPass {
				suite.Require().NoError(err)
				// if data is to big, slice the result
				if len(res.Data) > 150 {
					suite.Require().Equal(tc.traceResponse, string(res.Data[:150]))
				} else {
					suite.Require().Equal(tc.traceResponse, string(res.Data))
				}
			} else {
				suite.Require().Error(err)
			}
			// Reset for next case
			chainID = nil
		})
	}

	suite.enableFeemarket = false // reset flag
}

func (suite *KeeperTestSuite) TestNonceInQuery() {
	address := utiltx.GenerateAddress()
	suite.Require().Equal(uint64(0), suite.app.EvmKeeper.GetNonce(suite.ctx, address))
	supply := sdkmath.NewIntWithDecimal(1000, 18).BigInt()

	// accupy nonce 0
	_ = suite.DeployTestContract(suite.T(), address, supply)

	// do an EthCall/EstimateGas with nonce 0
	ctorArgs, err := types.ERC20Contract.ABI.Pack("", address, supply)
	suite.Require().NoError(err)

	data := types.ERC20Contract.Bin
	data = append(data, ctorArgs...)
	args, err := json.Marshal(&types.TransactionArgs{
		From: &address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)
	proposerAddress := suite.ctx.BlockHeader().ProposerAddress
	_, err = suite.queryClient.EstimateGas(suite.ctx, &types.EthCallRequest{
		Args:            args,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: proposerAddress,
	})
	suite.Require().NoError(err)

	_, err = suite.queryClient.EthCall(suite.ctx, &types.EthCallRequest{
		Args:            args,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: proposerAddress,
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestQueryBaseFee() {
	var (
		aux    sdkmath.Int
		expRes *types.QueryBaseFeeResponse
	)

	testCases := []struct {
		name            string
		malleate        func()
		expPass         bool
		enableFeemarket bool
		enableLondonHF  bool
	}{
		{
			"pass - default Base Fee",
			func() {
				initialBaseFee := sdkmath.NewInt(ethparams.InitialBaseFee)
				expRes = &types.QueryBaseFeeResponse{BaseFee: &initialBaseFee}
			},
			true, true, true,
		},
		{
			"pass - non-nil Base Fee",
			func() {
				baseFee := sdkmath.OneInt().BigInt()
				suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, baseFee)

				aux = sdkmath.NewIntFromBigInt(baseFee)
				expRes = &types.QueryBaseFeeResponse{BaseFee: &aux}
			},
			true, true, true,
		},
		{
			"pass - nil Base Fee when london hardfork not activated",
			func() {
				baseFee := sdkmath.OneInt().BigInt()
				suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, baseFee)

				expRes = &types.QueryBaseFeeResponse{}
			},
			true, true, false,
		},
		{
			"pass - zero Base Fee when feemarket not activated",
			func() {
				baseFee := sdkmath.ZeroInt()
				expRes = &types.QueryBaseFeeResponse{BaseFee: &baseFee}
			},
			true, false, true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest()

			tc.malleate()

			res, err := suite.queryClient.BaseFee(suite.ctx.Context(), &types.QueryBaseFeeRequest{})
			if tc.expPass {
				suite.Require().NotNil(res)
				suite.Require().Equal(expRes, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite *KeeperTestSuite) TestEthCall() {
	var req *types.EthCallRequest

	address := utiltx.GenerateAddress()
	suite.Require().Equal(uint64(0), suite.app.EvmKeeper.GetNonce(suite.ctx, address))
	supply := sdkmath.NewIntWithDecimal(1000, 18).BigInt()

	hexBigInt := hexutil.Big(*big.NewInt(1))
	ctorArgs, err := types.ERC20Contract.ABI.Pack("", address, supply)
	suite.Require().NoError(err)

	data := types.ERC20Contract.Bin
	data = append(data, ctorArgs...)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid args",
			func() {
				req = &types.EthCallRequest{Args: []byte("invalid args"), GasCap: config.DefaultGasCap}
			},
			false,
		},
		{
			"invalid args - specified both gasPrice and maxFeePerGas",
			func() {
				args, err := json.Marshal(&types.TransactionArgs{
					From:         &address,
					Data:         (*hexutil.Bytes)(&data),
					GasPrice:     &hexBigInt,
					MaxFeePerGas: &hexBigInt,
				})

				suite.Require().NoError(err)
				req = &types.EthCallRequest{Args: args, GasCap: config.DefaultGasCap}
			},
			false,
		},
		{
			"set param EnableCreate = false",
			func() {
				args, err := json.Marshal(&types.TransactionArgs{
					From: &address,
					Data: (*hexutil.Bytes)(&data),
				})

				suite.Require().NoError(err)
				req = &types.EthCallRequest{Args: args, GasCap: config.DefaultGasCap}

				params := suite.app.EvmKeeper.GetParams(suite.ctx)
				params.EnableCreate = false
				err = suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"invalid state override",
			func() {
				args, err := json.Marshal(&types.TransactionArgs{
					From: &address,
					Data: (*hexutil.Bytes)(&data),
				})
				suite.Require().NoError(err)
				req = &types.EthCallRequest{
					Args:          args,
					GasCap:        config.DefaultGasCap,
					StateOverride: []byte("not valid json"),
				}
			},
			false,
		},
		{
			"valid state override",
			func() {
				args, err := json.Marshal(&types.TransactionArgs{
					From: &address,
					Data: (*hexutil.Bytes)(&data),
				})
				suite.Require().NoError(err)

				balanceOverride := map[common.Address]map[string]interface{}{
					address: {
						"balance": "0x64",
					},
				}
				overrideBytes, err := json.Marshal(balanceOverride)
				suite.Require().NoError(err)

				req = &types.EthCallRequest{
					Args:          args,
					GasCap:        config.DefaultGasCap,
					StateOverride: overrideBytes,
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			res, err := suite.queryClient.EthCall(suite.ctx, req)
			if tc.expPass {
				suite.Require().NotNil(res)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEmptyRequest() {
	k := suite.app.EvmKeeper

	testCases := []struct {
		name      string
		queryFunc func() (interface{}, error)
	}{
		{
			"Account method",
			func() (interface{}, error) {
				return k.Account(suite.ctx, nil)
			},
		},
		{
			"CosmosAccount method",
			func() (interface{}, error) {
				return k.CosmosAccount(suite.ctx, nil)
			},
		},
		{
			"ValidatorAccount method",
			func() (interface{}, error) {
				return k.ValidatorAccount(suite.ctx, nil)
			},
		},
		{
			"Balance method",
			func() (interface{}, error) {
				return k.Balance(suite.ctx, nil)
			},
		},
		{
			"Storage method",
			func() (interface{}, error) {
				return k.Storage(suite.ctx, nil)
			},
		},
		{
			"Code method",
			func() (interface{}, error) {
				return k.Code(suite.ctx, nil)
			},
		},
		{
			"EthCall method",
			func() (interface{}, error) {
				return k.EthCall(suite.ctx, nil)
			},
		},
		{
			"EstimateGas method",
			func() (interface{}, error) {
				return k.EstimateGas(suite.ctx, nil)
			},
		},
		{
			"TraceTx method",
			func() (interface{}, error) {
				return k.TraceTx(suite.ctx, nil)
			},
		},
		{
			"TraceBlock method",
			func() (interface{}, error) {
				return k.TraceBlock(suite.ctx, nil)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			_, err := tc.queryFunc()
			suite.Require().Error(err)
		})
	}
}

// -----------------------------------------------------------------------------
// SimulateV1 — eth_simulateV1 gRPC handler
// -----------------------------------------------------------------------------

// simulateV1Request constructs a SimulateV1Request for the keeper's
// public gRPC handler from a pre-encoded opts JSON payload, populating
// the chain-id / proposer fields from the test suite's active context.
// Tests build opts JSON by hand rather than reaching into the driver's
// private SimOpts type.
func (suite *KeeperTestSuite) simulateV1Request(optsJSON []byte) *types.SimulateV1Request {
	suite.T().Helper()

	bn := rpctypes.BlockNumber(suite.ctx.BlockHeight())
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	bnhBz, err := json.Marshal(bnh)
	suite.Require().NoError(err)

	return &types.SimulateV1Request{
		Opts:              optsJSON,
		BlockNumberOrHash: bnhBz,
		GasCap:            21_000_000,
		ProposerAddress:   sdk.ConsAddress(suite.ctx.BlockHeader().ProposerAddress),
		ChainId:           suite.app.EvmKeeper.ChainID().Int64(),
	}
}

// simulateV1BlockResults unmarshals the keeper's JSON Result payload.
func (suite *KeeperTestSuite) simulateV1BlockResults(resp *types.SimulateV1Response) []map[string]interface{} {
	suite.T().Helper()
	var out []map[string]interface{}
	suite.Require().NoError(json.Unmarshal(resp.Result, &out))
	return out
}

// TestSimulateV1_EmptyOpts: empty blockStateCalls returns an empty
// result payload with no structured error on the response.
func (suite *KeeperTestSuite) TestSimulateV1_EmptyOpts() {
	suite.SetupTest()

	optsJSON, err := json.Marshal(map[string]interface{}{})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Empty(results)
}

// TestSimulateV1_SingleCallHappyPath: value transfer with a balance
// override on the sender — asserts status=1 and non-zero gasUsed on
// the per-call result carried in the JSON block envelope.
func (suite *KeeperTestSuite) TestSimulateV1_SingleCallHappyPath() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls, ok := results[0]["calls"].([]interface{})
	suite.Require().True(ok)
	suite.Require().Len(calls, 1)

	call := calls[0].(map[string]interface{})
	suite.Require().Nil(call["error"], "simple transfer must not fail")
	suite.Require().Equal("0x1", call["status"])
	suite.Require().NotEqual("0x0", call["gasUsed"])
}

// TestSimulateV1_BaseBlockHashOverridesParentHash: when the request
// carries a non-zero BaseBlockHash, the first simulated block's
// parentHash must echo that value verbatim. Without it, the keeper
// would hash the synthetic baseHeaderFromContext, producing a value
// unrelated to the canonical chain hash eth_getBlockByNumber surfaces.
func (suite *KeeperTestSuite) TestSimulateV1_BaseBlockHashOverridesParentHash() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
	})
	suite.Require().NoError(err)

	canonical := common.HexToHash("0x65fdad50586258b80fdeec1e9d108e975d20a1a34ab3dfadd97eeedffa0727cc")
	req := suite.simulateV1Request(optsJSON)
	req.BaseBlockHash = canonical.Bytes()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, req)
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	suite.Require().Equal(canonical.Hex(), results[0]["parentHash"])
}

// TestSimulateV1_InheritsBaseBlockGasLimit: even if the incoming SDK
// context is missing consensus params, the first simulated block should
// still inherit the base block gas limit from consensus keeper state.
func (suite *KeeperTestSuite) TestSimulateV1_InheritsBaseBlockGasLimit() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
	})
	suite.Require().NoError(err)

	ctx := suite.ctx.WithConsensusParams(tmproto.ConsensusParams{})
	resp, err := suite.app.EvmKeeper.SimulateV1(ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	consensusParamsResp, err := suite.app.ConsensusParamsKeeper.Params(ctx, nil)
	suite.Require().NoError(err)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	suite.Require().Equal(
		hexutil.EncodeUint64(uint64(consensusParamsResp.GetParams().GetBlock().MaxGas)), //nolint:gosec
		results[0]["gasLimit"],
	)
}

// TestSimulateV1_StateOverrideSentinelBubblesUp: a self-referencing
// MovePrecompileTo override must surface on response.Error with code
// -38022 — not as a gRPC error.
func (suite *KeeperTestSuite) TestSimulateV1_StateOverrideSentinelBubblesUp() {
	suite.SetupTest()

	sha256Addr := common.HexToAddress("0x0000000000000000000000000000000000000002")
	sender := suite.address

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sha256Addr: {"movePrecompileToAddress": sha256Addr},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &sender}},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeMovePrecompileSelfReference), resp.Error.Code)
	suite.Require().Contains(resp.Error.Message, sha256Addr.Hex())
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_MovePrecompileToSha256: relocate stdlib sha256 to a
// fresh destination; calling the destination must return the
// canonical sha256 digest for empty input.
func (suite *KeeperTestSuite) TestSimulateV1_MovePrecompileToSha256() {
	suite.SetupTest()

	sha256Addr := common.HexToAddress("0x0000000000000000000000000000000000000002")
	dest := common.HexToAddress("0x1234000000000000000000000000000000000000")
	sender := suite.address

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sha256Addr: {"movePrecompileToAddress": dest},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &dest}},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 1)

	call := calls[0].(map[string]interface{})
	suite.Require().Nil(call["error"], "call to moved precompile must succeed")
	suite.Require().Equal(
		"0xe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		call["returnData"],
	)
}

// TestSimulateV1_NilRequest: nil request returns an InvalidArgument
// gRPC error.
func (suite *KeeperTestSuite) TestSimulateV1_NilRequest() {
	suite.SetupTest()
	_, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, nil)
	suite.Require().Error(err)
	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Require().Equal(codes.InvalidArgument, st.Code())
}

// TestSimulateV1_UnsupportedOverrideRejected: BeaconRoot override is
// rejected at the opts unmarshal step with a -32602 SimError on
// response.Error, not as a gRPC status.
func (suite *KeeperTestSuite) TestSimulateV1_UnsupportedOverrideRejected() {
	suite.SetupTest()

	beacon := common.Hash{1}
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"blockOverrides": map[string]interface{}{"beaconRoot": beacon},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeInvalidParams), resp.Error.Code)
	suite.Require().Contains(resp.Error.Message, "BeaconRoot")
	suite.Require().Empty(resp.Result)
}

// -----------------------------------------------------------------------------
// SimulateV1 — multi-call / multi-block / BLOCKHASH end-to-end tests
// -----------------------------------------------------------------------------

// branchingSlot0Bytecode is a tiny in-test contract whose behavior forks
// on CALLDATASIZE:
//   - empty calldata → SLOAD slot 0, RETURN 32-byte value.
//   - 32-byte calldata → SSTORE slot 0 = calldata, STOP.
//
// Assembly trace (offsets correspond to the hex below):
//
//	@00  CALLDATASIZE         stack=[size]
//	@01  ISZERO               stack=[size==0]
//	@02  PUSH1 0x0C           stack=[cond, 12]
//	@04  JUMPI                jump to 12 iff calldata empty
//	@05  PUSH1 0x00           write path: offset for CALLDATALOAD
//	@07  CALLDATALOAD         stack=[value]
//	@08  PUSH1 0x00           stack=[value, slot=0]
//	@0A  SSTORE               storage[0]=value
//	@0B  STOP
//	@0C  JUMPDEST             read path
//	@0D  PUSH1 0x00           slot
//	@0F  SLOAD                stack=[value]
//	@10  PUSH1 0x00           mem offset
//	@12  MSTORE               memory[0..32]=value
//	@13  PUSH1 0x20           size
//	@15  PUSH1 0x00           offset
//	@17  RETURN
const branchingSlot0Bytecode = "0x3615600C57600035600055005B60005460005260206000F3"

// blockhashReaderBytecode reads calldata[0:32] as a uint256 height,
// returns BLOCKHASH(height) as 32 bytes.
//
//	PUSH1 0 CALLDATALOAD BLOCKHASH PUSH1 0 MSTORE PUSH1 0x20 PUSH1 0 RETURN
const blockhashReaderBytecode = "0x6000354060005260206000F3"

// revertAfterWriteSlot1Bytecode branches on CALLDATASIZE so the same
// deployed contract covers both sides of the revert-isolation assertion:
//   - non-empty calldata → SSTORE slot 1 = 42, then REVERT(0,0).
//   - empty calldata     → SLOAD slot 1, RETURN 32-byte value.
//
// Having call 1 (writer+revert) and call 2 (reader) hit the same
// address makes the test diagnostic: if the revert leaked, call 2
// would return 42; with correct isolation it returns 0.
//
//	@00  CALLDATASIZE
//	@01  ISZERO
//	@02  PUSH1 0x0F
//	@04  JUMPI                → jump to reader on empty calldata
//	@05  PUSH1 0x2A           writer path: value 42
//	@07  PUSH1 0x01           slot 1
//	@09  SSTORE
//	@0A  PUSH1 0x00
//	@0C  PUSH1 0x00
//	@0E  REVERT
//	@0F  JUMPDEST             reader path
//	@10  PUSH1 0x01
//	@12  SLOAD
//	@13  PUSH1 0x00
//	@15  MSTORE
//	@16  PUSH1 0x20
//	@18  PUSH1 0x00
//	@1A  RETURN
const revertAfterWriteSlot1Bytecode = "0x3615600F57602A60015560006000FD5B60015460005260206000F3"

// emptyCodeDeployer deploys a contract whose runtime code is zero bytes.
// Per EIP-161 (Spurious Dragon) the newly-created account is still
// initialized with nonce=1, which is what makes the deployed address
// distinct from the sender's next CREATE target. Used by the
// nonce-auto-increment test: two CREATEs from the same sender must
// succeed at distinct addresses, which only happens if the sender's
// nonce advances between calls (otherwise the second CREATE collides
// at the first address).
//
//	PUSH1 0x00 PUSH1 0x00 RETURN
const emptyCodeDeployer = "0x60006000F3"

// TestSimulateV1_MultiCall_StateChainsAcrossCalls — call 1 writes slot 0
// via the branching-slot-0 contract; call 2 reads slot 0 and must see
// the written value. Proves the shared StateDB carries storage state
// from one call to the next.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_StateChainsAcrossCalls() {
	suite.SetupTest()

	sender := suite.address
	contract := common.HexToAddress("0xaaaa000000000000000000000000000000000100")

	// 32-byte calldata for the write call encoding the value 255 in slot 0.
	writeData := make([]byte, 32)
	writeData[31] = 0xFF

	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:   {"balance": balance},
				contract: {"code": branchingSlot0Bytecode},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &contract, Input: (*hexutil.Bytes)(&writeData)},
				{From: &sender, To: &contract},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 2)

	suite.Require().Equal("0x1", calls[0].(map[string]interface{})["status"], "call 1 should succeed")
	suite.Require().Equal("0x1", calls[1].(map[string]interface{})["status"], "call 2 should succeed")

	// Call 2's returnData must be the 32-byte encoding of 255 that call 1
	// wrote — confirms the shared StateDB propagated storage.
	readBack := calls[1].(map[string]interface{})["returnData"].(string)
	suite.Require().Equal(
		"0x00000000000000000000000000000000000000000000000000000000000000ff",
		readBack,
	)
}

// TestSimulateV1_MultiCall_RevertDoesNotLeak — call 1 writes slot 1 and
// reverts; call 2 reads slot 1 on the SAME contract and must see zero.
// Under a (hypothetical) bug where the per-call EVM revert failed to
// roll the shared StateDB back, call 2 would return 42.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_RevertDoesNotLeak() {
	suite.SetupTest()

	sender := suite.address
	contract := common.HexToAddress("0xaaaa000000000000000000000000000000000200")
	writeCalldata := []byte{0x01} // any non-empty calldata takes the writer branch

	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:   {"balance": balance},
				contract: {"code": revertAfterWriteSlot1Bytecode},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &contract, Input: (*hexutil.Bytes)(&writeCalldata)},
				{From: &sender, To: &contract},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 2)

	call1 := calls[0].(map[string]interface{})
	suite.Require().Equal("0x0", call1["status"])
	suite.Require().NotNil(call1["error"])
	suite.Require().EqualValues(float64(types.SimErrCodeReverted), call1["error"].(map[string]interface{})["code"])

	call2 := calls[1].(map[string]interface{})
	suite.Require().Equal("0x1", call2["status"])
	suite.Require().Equal(
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		call2["returnData"],
		"reader on the same contract must observe slot 1 = 0 — the SSTORE in call 1 must have rolled back with the revert",
	)
}

// Two calls against a tight block gas limit; the second requests gas
// over the remaining block budget and trips a request-level -38015.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_BlockGasLimit() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000001")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	value := (*hexutil.Big)(big.NewInt(1))
	tightGasLimit := hexutil.Uint64(50_000)
	overBudgetGas := hexutil.Uint64(30_000)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"blockOverrides": map[string]interface{}{
				"gasLimit": tightGasLimit,
			},
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipient, Value: value},
				// 30k > 50k-21k=29k remaining: trips -38015.
				{From: &sender, To: &recipient, Value: value, Gas: &overBudgetGas},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err, "must NOT collapse to gRPC Internal")
	suite.Require().NotNil(resp.Error, "over-budget call must surface a top-level fatal SimError")
	suite.Require().Equal(int32(types.SimErrCodeBlockGasLimitReached), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// A call combining legacy GasPrice with EIP-1559 MaxFeePerGas fails
// ToMessage and surfaces a request-level -32602; any later call in the
// block must not execute.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_ToMessageRejection() {
	suite.SetupTest()

	sender := suite.address
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	gasPrice := (*hexutil.Big)(big.NewInt(1))
	maxFee := (*hexutil.Big)(big.NewInt(1))
	initCode := hexutil.MustDecode(emptyLogDeployer)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, GasPrice: gasPrice, MaxFeePerGas: maxFee, Input: (*hexutil.Bytes)(&initCode)},
				{From: &sender, Input: (*hexutil.Bytes)(&initCode)},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err, "must NOT collapse to gRPC Internal")
	suite.Require().NotNil(resp.Error, "ToMessage failure must surface a top-level fatal SimError")
	suite.Require().Equal(int32(types.SimErrCodeInvalidParams), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_MultiCall_NonceAutoIncrement — two CREATE calls from
// the same sender with no explicit nonce. Both must succeed: if the
// nonce didn't advance between calls, the second CREATE would resolve
// to the same computed address as the first and fail with an address-
// collision error.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_NonceAutoIncrement() {
	suite.SetupTest()

	sender := suite.address
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	initCode := hexutil.MustDecode(emptyCodeDeployer)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, Input: (*hexutil.Bytes)(&initCode)},
				{From: &sender, Input: (*hexutil.Bytes)(&initCode)},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 2)

	suite.Require().Equal("0x1", calls[0].(map[string]interface{})["status"],
		"first CREATE must succeed")
	suite.Require().Equal("0x1", calls[1].(map[string]interface{})["status"],
		"second CREATE must succeed — fails with address collision if nonce stayed at 0")
}

// emptyLogDeployer is init code that emits an empty LOG0 and returns
// zero-byte runtime. The emitted log's Address field is the deployed
// contract's own address (set by the EVM as evm.Address during init
// code execution), which lets a test observe which nonce the CREATE
// consumed without a follow-up call to the predicted address.
//
// Bytecode (10 bytes):
//
//	@00  PUSH1 0x00   ; log data size = 0
//	@02  PUSH1 0x00   ; log data offset = 0
//	@04  LOG0
//	@05  PUSH1 0x00   ; runtime size = 0
//	@07  PUSH1 0x00   ; runtime offset = 0
//	@09  RETURN
const emptyLogDeployer = "0x60006000A060006000F3"

// Two value transfers from the same sender with no explicit nonce.
// The driver must bump the StateDB nonce after every successful
// non-CREATE call; otherwise both calls would default to the same
// nonce, the synthesized tx hashes would collide, and the assembled
// block would carry duplicate entries in `transactions`.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_CallNonceAdvances() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000010")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	value := (*hexutil.Big)(big.NewInt(1))
	pinnedGas := hexutil.Uint64(30_000)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipient, Value: value, Gas: &pinnedGas},
				{From: &sender, To: &recipient, Value: value, Gas: &pinnedGas},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 2)
	suite.Require().Equal("0x1", calls[0].(map[string]interface{})["status"], "first CALL must succeed")
	suite.Require().Equal("0x1", calls[1].(map[string]interface{})["status"], "second CALL must succeed")

	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 2)
	suite.Require().NotEqual(txs[0], txs[1],
		"synthetic tx hashes must differ — proves the StateDB nonce advanced between CALLs")
}

// TestSimulateV1_MultiCall_CallThenCreateAddressUsesPostCallNonce —
// one CALL followed by one CREATE from the same sender, no explicit
// nonces. The CREATE deploys at addr(sender, 1) only if the prior
// CALL advanced the StateDB nonce; without the post-CALL bump it
// would land at addr(sender, 0). The init code emits an empty LOG0
// so the deployed contract's own address surfaces in the response
// via Log.Address — independent of the synthetic tx-hash signal.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_CallThenCreateAddressUsesPostCallNonce() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000020")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	value := (*hexutil.Big)(big.NewInt(1))
	initCode := hexutil.MustDecode(emptyLogDeployer)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipient, Value: value},
				{From: &sender, Input: (*hexutil.Bytes)(&initCode)},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 2)
	suite.Require().Equal("0x1", calls[0].(map[string]interface{})["status"], "CALL must succeed")
	suite.Require().Equal("0x1", calls[1].(map[string]interface{})["status"], "CREATE must succeed")

	createLogs := calls[1].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Len(createLogs, 1, "CREATE must emit one log")

	gotAddr := common.HexToAddress(createLogs[0].(map[string]interface{})["address"].(string))
	expectedAddr := crypto.CreateAddress(sender, 1)
	suite.Require().Equal(expectedAddr, gotAddr,
		"CREATE deployed at addr(sender, 1); the prior CALL must have advanced the StateDB nonce — "+
			"would be addr(sender, 0) if the bump didn't fire")
}

// TestSimulateV1_MultiBlock_StateChains — block 1 writes slot 0 on a
// contract; block 2 reads slot 0 on the same contract and sees the
// block-1 write. Confirms both the shared StateDB and the per-block
// finalize step preserve state across block boundaries.
func (suite *KeeperTestSuite) TestSimulateV1_MultiBlock_StateChains() {
	suite.SetupTest()

	sender := suite.address
	contract := common.HexToAddress("0xaaaa000000000000000000000000000000000300")
	writeData := make([]byte, 32)
	writeData[31] = 0x7A // arbitrary non-zero
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender:   {"balance": balance},
					contract: {"code": branchingSlot0Bytecode},
				},
				"calls": []types.TransactionArgs{
					{From: &sender, To: &contract, Input: (*hexutil.Bytes)(&writeData)},
				},
			},
			{
				"calls": []types.TransactionArgs{
					{From: &sender, To: &contract},
				},
			},
		},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 2, "two blocks in the response")

	block0Calls := results[0]["calls"].([]interface{})
	suite.Require().Len(block0Calls, 1)
	suite.Require().Equal("0x1", block0Calls[0].(map[string]interface{})["status"])

	block1Calls := results[1]["calls"].([]interface{})
	suite.Require().Len(block1Calls, 1)
	block1Call := block1Calls[0].(map[string]interface{})
	suite.Require().Equal("0x1", block1Call["status"])
	suite.Require().Equal(
		"0x000000000000000000000000000000000000000000000000000000000000007a",
		block1Call["returnData"],
		"block 2 read must observe block 1 write",
	)
}

// TestSimulateV1_MultiBlock_ChainLinkage — block 3 calls a BLOCKHASH
// reader with three heights (base.Number, base+1, base+2). The canonical
// hash must match `suite.ctx.HeaderHash()` for base.Number; the
// simulated-sibling hashes must match the in-memory headers for
// base+1 / base+2 as reported by the response envelope.
func (suite *KeeperTestSuite) TestSimulateV1_MultiBlock_ChainLinkage() {
	suite.SetupTest()

	sender := suite.address
	reader := common.HexToAddress("0xaaaa000000000000000000000000000000000400")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	baseHeight := suite.ctx.BlockHeight()
	heights := []int64{baseHeight, baseHeight + 1, baseHeight + 2}

	callsBlock3 := make([]types.TransactionArgs, 0, len(heights))
	for _, h := range heights {
		hx := h
		buf := make([]byte, 32)
		big.NewInt(hx).FillBytes(buf)
		calldata := make([]byte, 32)
		copy(calldata, buf)
		callsBlock3 = append(callsBlock3, types.TransactionArgs{
			From: &sender, To: &reader, Input: (*hexutil.Bytes)(&calldata),
		})
	}

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{ // block 1: install reader code + fund sender
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender: {"balance": balance},
					reader: {"code": blockhashReaderBytecode},
				},
				"calls": []types.TransactionArgs{},
			},
			{ // block 2: no-op
				"calls": []types.TransactionArgs{},
			},
			{ // block 3: BLOCKHASH reads
				"calls": callsBlock3,
			},
		},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 3)

	block3 := results[2]
	block3Calls := block3["calls"].([]interface{})
	suite.Require().Len(block3Calls, 3, "three BLOCKHASH calls in block 3")

	// Call 0 asks for BLOCKHASH(base.Number) — newSimGetHashFn delegates
	// to k.GetHashFn(ctx), which covers ctx.HeaderHash AND the
	// tendermint-header-recompute fallback. Compute the expected value
	// with the same function so the test stays robust to how the test
	// ctx populates its header (ctx.HeaderHash may be empty in
	// suite-level contexts).
	expectedBase := "0x" + common.Bytes2Hex(
		suite.app.EvmKeeper.GetHashFn(suite.ctx)(uint64(baseHeight)).Bytes(), //nolint:gosec // baseHeight >= 0 by construction
	)
	suite.Require().Equal(expectedBase, block3Calls[0].(map[string]interface{})["returnData"],
		"BLOCKHASH(base.Number) must match k.GetHashFn(ctx)(base.Number)")

	// Calls 1 and 2 ask for simulated-sibling hashes (base+1, base+2).
	// These must match the hashes reported in the block envelope.
	expectedHash1 := results[0]["hash"].(string)
	expectedHash2 := results[1]["hash"].(string)
	suite.Require().Equal(expectedHash1, block3Calls[1].(map[string]interface{})["returnData"],
		"BLOCKHASH(base+1) must match block 1's envelope hash")
	suite.Require().Equal(expectedHash2, block3Calls[2].(map[string]interface{})["returnData"],
		"BLOCKHASH(base+2) must match block 2's envelope hash")
}

// TestSimulateV1_MultiBlock_PrecompileStateChains is the load-bearing
// regression for the shared-StateDB design: it exercises a *custom
// precompile* whose Cosmos-side writes live in the StateDB's cached
// ctx — the layer that a naive per-block StateDB would silently drop.
// Block 1 calls btctoken.transfer (routes through bankKeeper, mutates
// s.cachedCtx); block 2 calls btctoken.balanceOf and must observe the
// transferred amount. If s.cachedCtx were dropped between blocks,
// balanceOf would return zero and the test would fail loud.
//
// This complements the pure-EVM TestSimulateV1_MultiBlock_StateChains
// which only covers the EVM journal half of the shared StateDB. Both
// halves must survive the block boundary for the design to hold.
func (suite *KeeperTestSuite) TestSimulateV1_MultiBlock_PrecompileStateChains() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000077")
	btcToken := common.HexToAddress(types.BTCTokenPrecompileAddress)

	// Fund sender via bank directly — StateDB balance overrides only
	// touch the EVM state object and do not propagate to bankKeeper,
	// which is what btctoken.transfer actually reads.
	initialBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	transferAmount := big.NewInt(777_000_000_000_000_000)
	coin := sdk.NewCoin(types.DefaultEVMDenom, initialBalance)
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(
		suite.ctx, types.ModuleName, sdk.NewCoins(coin)))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(
		suite.ctx, types.ModuleName, sender.Bytes(), sdk.NewCoins(coin)))

	// ABI: transfer(address,uint256) — selector 0xa9059cbb; balanceOf(address)
	// — selector 0x70a08231. Encoded by hand to avoid dragging ABI binding
	// generation into keeper tests.
	transferSelector := []byte{0xa9, 0x05, 0x9c, 0xbb}
	balanceOfSelector := []byte{0x70, 0xa0, 0x82, 0x31}

	padAddr := func(addr common.Address) []byte {
		buf := make([]byte, 32)
		copy(buf[12:], addr.Bytes())
		return buf
	}
	padUint := func(v *big.Int) []byte {
		buf := make([]byte, 32)
		v.FillBytes(buf)
		return buf
	}

	transferData := append([]byte{}, transferSelector...)
	transferData = append(transferData, padAddr(recipient)...)
	transferData = append(transferData, padUint(transferAmount)...)

	balanceOfData := append([]byte{}, balanceOfSelector...)
	balanceOfData = append(balanceOfData, padAddr(recipient)...)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"calls": []types.TransactionArgs{
					{From: &sender, To: &btcToken, Input: (*hexutil.Bytes)(&transferData)},
				},
			},
			{
				"calls": []types.TransactionArgs{
					{From: &sender, To: &btcToken, Input: (*hexutil.Bytes)(&balanceOfData)},
				},
			},
		},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 2)

	block1Calls := results[0]["calls"].([]interface{})
	suite.Require().Len(block1Calls, 1)
	suite.Require().Equal("0x1", block1Calls[0].(map[string]interface{})["status"],
		"block 1 btctoken.transfer must succeed")

	block2Calls := results[1]["calls"].([]interface{})
	suite.Require().Len(block2Calls, 1)
	block2Call := block2Calls[0].(map[string]interface{})
	suite.Require().Equal("0x1", block2Call["status"],
		"block 2 btctoken.balanceOf must succeed")

	// Decode the 32-byte returnData and verify it matches transferAmount.
	// If s.cachedCtx were dropped between blocks, bankKeeper would read
	// canonical state (pre-transfer) and return 0 here.
	returnData := block2Call["returnData"].(string)
	got := new(big.Int)
	got.SetString(returnData[2:], 16)
	suite.Require().Equal(0, got.Cmp(transferAmount),
		"block 2 balanceOf must return transferAmount; got %s (precompile cachedCtx did not survive block boundary)", got.String())
}

// Precompile-emitted logs stamp log.BlockNumber from sdkCtx.BlockHeight()
// in EmitEvent; the driver anchors sdkCtx at the parent height for
// fork-gated reads, so without the back-stamp normalization those logs
// would report parent-height instead of the simulated height.
func (suite *KeeperTestSuite) TestSimulateV1_LogBlockNumber_MatchesSimulatedHeader() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000088")
	btcToken := common.HexToAddress(types.BTCTokenPrecompileAddress)

	// Fund sender via bank — btctoken.transfer reads bankKeeper, not
	// the EVM state-object balance, so a stateOverride wouldn't help.
	initialBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	transferAmount1 := big.NewInt(500_000_000_000_000_000)
	transferAmount2 := big.NewInt(123_000_000_000_000_000)
	coin := sdk.NewCoin(types.DefaultEVMDenom, initialBalance)
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(
		suite.ctx, types.ModuleName, sdk.NewCoins(coin)))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(
		suite.ctx, types.ModuleName, sender.Bytes(), sdk.NewCoins(coin)))

	// transfer(address,uint256) — selector 0xa9059cbb. Encoded by hand
	// to avoid pulling ABI binding generation into keeper tests.
	transferSelector := []byte{0xa9, 0x05, 0x9c, 0xbb}
	padAddr := func(addr common.Address) []byte {
		buf := make([]byte, 32)
		copy(buf[12:], addr.Bytes())
		return buf
	}
	padUint := func(v *big.Int) []byte {
		buf := make([]byte, 32)
		v.FillBytes(buf)
		return buf
	}
	encodeTransfer := func(to common.Address, amount *big.Int) []byte {
		out := append([]byte{}, transferSelector...)
		out = append(out, padAddr(to)...)
		out = append(out, padUint(amount)...)
		return out
	}

	transferData1 := encodeTransfer(recipient, transferAmount1)
	transferData2 := encodeTransfer(recipient, transferAmount2)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"calls": []types.TransactionArgs{
					{From: &sender, To: &btcToken, Input: (*hexutil.Bytes)(&transferData1)},
					{From: &sender, To: &btcToken, Input: (*hexutil.Bytes)(&transferData2)},
				},
			},
		},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	blockNumber := results[0]["number"].(string)
	blockHash := results[0]["hash"].(string)
	suite.Require().NotEmpty(blockNumber)
	suite.Require().NotEmpty(blockHash)

	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 2)

	totalLogs := 0
	for callIdx, raw := range calls {
		call := raw.(map[string]interface{})
		suite.Require().Equal("0x1", call["status"],
			"call %d btctoken.transfer must succeed", callIdx)

		logsRaw, ok := call["logs"].([]interface{})
		suite.Require().True(ok, "call %d logs must be present as a JSON array", callIdx)

		for logIdx, lraw := range logsRaw {
			log := lraw.(map[string]interface{})
			suite.Require().Equal(blockNumber, log["blockNumber"],
				"call %d log %d blockNumber must equal simulated header number", callIdx, logIdx)
			suite.Require().Equal(blockHash, log["blockHash"],
				"call %d log %d blockHash must equal simulated header hash", callIdx, logIdx)
			totalLogs++
		}
	}

	// Guard against a future change that silently drops the precompile
	// Transfer event — the per-log assertions above are vacuous on an
	// empty logs array.
	suite.Require().Greater(totalLogs, 0,
		"at least one log must be emitted across the two btctoken.transfer calls")
}

// First call exhausts the per-block gas; the second omits args.Gas and
// trips -38015 in resolveSimCallGas as a request-level fatal.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_ImplicitGasAfterExhaustedBudget() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000201")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	value := (*hexutil.Big)(big.NewInt(1))

	tightGasLimit := hexutil.Uint64(21_000)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"blockOverrides": map[string]interface{}{
				"gasLimit": tightGasLimit,
			},
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipient, Value: value},
				// Omit Gas; remaining=0 after call 1 trips the preflight.
				{From: &sender, To: &recipient, Value: value},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err, "the request must NOT collapse to gRPC Internal")
	suite.Require().NotNil(resp.Error, "exhausted-budget preflight must surface a top-level fatal SimError")
	suite.Require().Equal(int32(types.SimErrCodeBlockGasLimitReached), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// args.Gas below intrinsic surfaces as a request-level -38013.
func (suite *KeeperTestSuite) TestSimulateV1_ExplicitGasBelowIntrinsic() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000202")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	value := (*hexutil.Big)(big.NewInt(1))
	tooLowGas := hexutil.Uint64(20_999) // intrinsic-gas baseline is 21000

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipient, Value: value, Gas: &tooLowGas},
			},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err, "the request must NOT collapse to gRPC Internal")
	suite.Require().NotNil(resp.Error, "below-intrinsic call must surface a top-level fatal SimError")
	suite.Require().Equal(int32(types.SimErrCodeIntrinsicGas), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_MultiBlock_ParentHashChainCoherent — three blocks, each
// with one successful transfer, each with a non-trivial cumulative
// gasUsed. The response envelope's `block[i+1].parentHash` must equal
// `block[i].hash`, and `block[0].parentHash` must equal the request's
// BaseBlockHash.
func (suite *KeeperTestSuite) TestSimulateV1_MultiBlock_ParentHashChainCoherent() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000200")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	value := (*hexutil.Big)(big.NewInt(1))

	transfer := types.TransactionArgs{From: &sender, To: &recipient, Value: value}

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender: {"balance": balance},
				},
				"calls": []types.TransactionArgs{transfer},
			},
			{"calls": []types.TransactionArgs{transfer}},
			{"calls": []types.TransactionArgs{transfer}},
		},
	})
	suite.Require().NoError(err)

	canonical := common.HexToHash("0x65fdad50586258b80fdeec1e9d108e975d20a1a34ab3dfadd97eeedffa0727cc")
	req := suite.simulateV1Request(optsJSON)
	req.BaseBlockHash = canonical.Bytes()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, req)
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 3)

	suite.Require().Equal(canonical.Hex(), results[0]["parentHash"],
		"block[0].parentHash must echo BaseBlockHash")

	for i := 0; i < len(results)-1; i++ {
		suite.Require().Equal(
			results[i]["hash"], results[i+1]["parentHash"],
			"block[%d].parentHash must equal block[%d].hash", i+1, i,
		)
	}
}

// log0Runtime is 6 bytes: PUSH1 0; PUSH1 0; LOG0; STOP. A call to a
// contract carrying this runtime emits one empty LOG0 (zero data, zero
// topics) whose Address is the contract address.
const log0Runtime = "0x60006000A000"

// TestSimulateV1_MultiCall_CumulativeLogIndex — two calls in one block
// each emitting a single LOG0 must produce log indices 0 and 1
// respectively (geth-aligned per-block monotonicity). A second block
// with a single LOG0 call must reset the counter to 0.
func (suite *KeeperTestSuite) TestSimulateV1_MultiCall_CumulativeLogIndex() {
	suite.SetupTest()

	sender := suite.address
	emitter := common.HexToAddress("0xbbbb000000000000000000000000000000000203")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	emit := types.TransactionArgs{From: &sender, To: &emitter}

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender:  {"balance": balance},
					emitter: {"code": log0Runtime},
				},
				"calls": []types.TransactionArgs{emit, emit},
			},
			{"calls": []types.TransactionArgs{emit}},
		},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 2)

	block0Calls := results[0]["calls"].([]interface{})
	suite.Require().Len(block0Calls, 2)

	logs0 := block0Calls[0].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Len(logs0, 1)
	suite.Require().Equal("0x0", logs0[0].(map[string]interface{})["logIndex"],
		"first call's LOG0 must land at logIndex 0")

	logs1 := block0Calls[1].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Len(logs1, 1)
	suite.Require().Equal("0x1", logs1[0].(map[string]interface{})["logIndex"],
		"second call's LOG0 must land at logIndex 1 (cumulative within block)")

	// Block 2: counter must reset.
	block1Calls := results[1]["calls"].([]interface{})
	suite.Require().Len(block1Calls, 1)
	logs2 := block1Calls[0].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Len(logs2, 1)
	suite.Require().Equal("0x0", logs2[0].(map[string]interface{})["logIndex"],
		"block 2's first log must reset to logIndex 0")
}

// -----------------------------------------------------------------------------
// SimulateV1 — DoS guards
// -----------------------------------------------------------------------------

// 257 sequential blocks must be rejected by sanitizeSimChain.
func (suite *KeeperTestSuite) TestSimulateV1_DoS_BlockCap() {
	suite.SetupTest()

	base := suite.ctx.BlockHeight()
	blocks := make([]map[string]interface{}, types.MaxSimulateBlocks+1)
	for i := range blocks {
		blocks[i] = map[string]interface{}{
			"blockOverrides": map[string]interface{}{
				"number": hexutil.EncodeUint64(uint64(base + int64(i+1))), //nolint:gosec
			},
		}
	}
	optsJSON, err := json.Marshal(map[string]interface{}{"blockStateCalls": blocks})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeClientLimitExceeded), resp.Error.Code)
}

// 1009 calls (> MaxSimulateCalls) within the block-count cap: keeper
// rejects with -38026.
func (suite *KeeperTestSuite) TestSimulateV1_DoS_CallCap() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	v := (*hexutil.Big)(big.NewInt(1))

	const wide = 251
	const widePerBlock = 4
	const narrow = 5
	const narrowPerBlock = 1
	totalCalls := wide*widePerBlock + narrow*narrowPerBlock // 1009 > 1000

	base := suite.ctx.BlockHeight()
	blocks := make([]map[string]interface{}, 0, wide+narrow)
	bn := uint64(base) //nolint:gosec
	appendBlocks := func(count, callsPerBlock int) {
		for b := 0; b < count; b++ {
			bn++
			callsList := make([]types.TransactionArgs, callsPerBlock)
			for i := range callsList {
				callsList[i] = types.TransactionArgs{From: &sender, To: &recipient, Value: v}
			}
			blocks = append(blocks, map[string]interface{}{
				"blockOverrides": map[string]interface{}{
					"number": hexutil.EncodeUint64(bn),
				},
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender: {"balance": balance},
				},
				"calls": callsList,
			})
		}
	}
	appendBlocks(wide, widePerBlock)
	appendBlocks(narrow, narrowPerBlock)

	optsJSON, err := json.Marshal(map[string]interface{}{"blockStateCalls": blocks})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeClientLimitExceeded), resp.Error.Code)
	suite.Require().Contains(resp.Error.Message, "calls")
	suite.Require().Contains(resp.Error.Message, fmt.Sprintf("%d", totalCalls))
}

// TestSimulateV1_DoS_GasPool_ClampsCallGas: with GasCap small relative to
// the request's 30M-default block gas, an explicit args.Gas above GasCap
// must be clamped down by the budget. The call still succeeds (intrinsic
// gas is 21k, well under the clamp) and gasUsed reflects the clamp via
// minGasMultiplier inflation.
func (suite *KeeperTestSuite) TestSimulateV1_DoS_GasPool_ClampsCallGas() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	// Explicit args.Gas larger than GasCap.
	gas := hexutil.Uint64(5_000_000)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipient, Value: value, Gas: &gas},
			},
		}},
	})
	suite.Require().NoError(err)

	req := suite.simulateV1Request(optsJSON)
	req.GasCap = 1_000_000

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, req)
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 1)
	call := calls[0].(map[string]interface{})
	suite.Require().Equal("0x1", call["status"])

	gasUsedHex := call["gasUsed"].(string)
	gasUsed, err := hexutil.DecodeUint64(gasUsedHex)
	suite.Require().NoError(err)
	suite.Require().LessOrEqual(gasUsed, uint64(1_000_000), "gasUsed must respect the request-wide gas cap")
}

// gasCap=0 is unlimited.
func (suite *KeeperTestSuite) TestSimulateV1_DoS_GasPool_ZeroIsUnlimited() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
	})
	suite.Require().NoError(err)

	req := suite.simulateV1Request(optsJSON)
	req.GasCap = 0

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, req)
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 1)
	suite.Require().Equal("0x1", calls[0].(map[string]interface{})["status"])
}

// Pre-canceled request context surfaces -32016 from the loop-top check.
func (suite *KeeperTestSuite) TestSimulateV1_Timeout_LoopCheck() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
	})
	suite.Require().NoError(err)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	sdkCtx := suite.ctx.WithContext(canceledCtx)

	req := suite.simulateV1Request(optsJSON)
	req.TimeoutMs = 1_000 // shape the message text; cancellation is what fires

	resp, err := suite.app.EvmKeeper.SimulateV1(sdkCtx, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeTimeout), resp.Error.Code)
}

// infiniteJumpBytecode is a tight EVM loop that runs until canceled or
// out of gas:
//
//	@00  JUMPDEST
//	@01  PUSH1 0x00
//	@03  JUMP            → unconditional jump back to @00
//
// Each iteration costs 1 (JUMPDEST) + 3 (PUSH1) + 8 (JUMP) = 12 gas, so
// even at the keeper's 21M GasCap the loop runs ~1.75M iterations before
// gas runs out. With a 200ms request timeout the watcher's evm.Cancel()
// fires long before that.
const infiniteJumpBytecode = "0x5B600056"

// Mid-call request-context cancellation surfaces -32016 from the
// post-call timeout check. Drives a tight EVM loop (infiniteJumpBytecode)
// against a small TimeoutMs so the cancel-watcher's evm.Cancel() fires
// while applyMessageWithConfig is actually running, exercising a path
// distinct from TestSimulateV1_Timeout_LoopCheck (which only covers the
// pre-call ctx.Err() guard at the top of the per-block loop).
func (suite *KeeperTestSuite) TestSimulateV1_Timeout_MidCall() {
	suite.SetupTest()

	sender := suite.address
	contract := common.HexToAddress("0xaaaa000000000000000000000000000000000300")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:   {"balance": balance},
				contract: {"code": infiniteJumpBytecode},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &contract}},
		}},
	})
	suite.Require().NoError(err)

	req := suite.simulateV1Request(optsJSON)
	// Gas budget high enough that the loop must be bounded by the
	// 200ms timeout, not by gas exhaustion.
	req.GasCap = 1_000_000_000
	req.TimeoutMs = 200

	start := time.Now()
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, req)
	elapsed := time.Since(start)

	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error, "expected request-level SimError on mid-call timeout")
	suite.Require().Equal(int32(types.SimErrCodeTimeout), resp.Error.Code)
	suite.Require().Empty(resp.Result, "request-level fatal must not emit per-call entries")
	// Wall-clock bound confirms the deadline fired and the test isn't
	// running the bytecode to gas exhaustion.
	suite.Require().Less(elapsed, 2*time.Second,
		"timeout did not fire mid-call; elapsed=%s", elapsed)
}

// -----------------------------------------------------------------------------
// SimulateV1 — TraceTransfers (ERC-7528 synthetic Transfer logs)
// -----------------------------------------------------------------------------

// erc7528Address is the canonical ERC-7528 pseudo-address used as the
// emitter of synthetic native-value Transfer logs.
const erc7528Address = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"

// erc20TransferTopic is keccak256("Transfer(address,address,uint256)").
const erc20TransferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// valueForwarderRuntime forwards CALLVALUE to the address packed (left-
// padded) into the first 32 bytes of calldata. Two value-bearing call
// edges land per top-level invocation: outer (caller -> forwarder) and
// inner (forwarder -> recipient). Used by the nested-CALL synthetic-log
// test.
//
//	@00  PUSH1 0x00          ; retSize
//	@02  PUSH1 0x00          ; retOffset
//	@04  PUSH1 0x00          ; argsSize
//	@06  PUSH1 0x00          ; argsOffset
//	@08  CALLVALUE           ; value
//	@09  PUSH1 0x00          ; CALLDATALOAD offset
//	@0B  CALLDATALOAD        ; recipient (low 20 bytes of calldata[0..32])
//	@0C  GAS                 ; forward all remaining gas
//	@0D  CALL
//	@0E  STOP
const valueForwarderRuntime = "0x6000600060006000346000355AF100"

// log0RevertRuntime emits an empty LOG0 then reverts. Used to pin the
// "tracer drops logs from a reverted frame" path.
//
//	@00  PUSH1 0x00          ; LOG0 size
//	@02  PUSH1 0x00          ; LOG0 offset
//	@04  LOG0
//	@05  PUSH1 0x00          ; REVERT size
//	@07  PUSH1 0x00          ; REVERT offset
//	@09  REVERT
const log0RevertRuntime = "0x60006000A060006000FD"

// erc20TransferCalldata builds calldata for ERC-20 transfer(to, amount).
// Selector is keccak256("transfer(address,uint256)")[:4] = 0xa9059cbb.
func erc20TransferCalldata(to common.Address, amount *big.Int) []byte {
	const transferSelector = "a9059cbb"
	out := common.Hex2Bytes(transferSelector)
	out = append(out, common.LeftPadBytes(to.Bytes(), 32)...)
	out = append(out, common.LeftPadBytes(amount.Bytes(), 32)...)
	return out
}

// syntheticTransferLogs filters a logs JSON array down to entries whose
// `address` matches the ERC-7528 pseudo-address (case-insensitive). The
// JSON encoding of ethtypes.Log lower-cases the address, but the helper
// stays case-insensitive to survive future encoding tweaks.
func syntheticTransferLogs(logs []interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(logs))
	for _, raw := range logs {
		entry := raw.(map[string]interface{})
		addr, ok := entry["address"].(string)
		if !ok {
			continue
		}
		if common.HexToAddress(addr) == common.HexToAddress(erc7528Address) {
			out = append(out, entry)
		}
	}
	return out
}

// TestSimulateV1_TraceTransfers_On_NativeTransfer_OneSyntheticLog —
// happy path: a single native value-transfer call with traceTransfers
// enabled produces exactly one synthetic ERC-7528 Transfer log carrying
// the canonical Transfer topic, indexed sender / recipient, and the
// value as 32-byte data.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_NativeTransfer_OneSyntheticLog() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 1)

	call := calls[0].(map[string]interface{})
	suite.Require().Equal("0x1", call["status"])

	logs := call["logs"].([]interface{})
	synthetic := syntheticTransferLogs(logs)
	suite.Require().Len(synthetic, 1, "value transfer must produce exactly one synthetic ERC-7528 log")

	log := synthetic[0]
	topics := log["topics"].([]interface{})
	suite.Require().Len(topics, 3)
	suite.Require().Equal(erc20TransferTopic, topics[0].(string))

	// Indexed sender / recipient: 12 zero bytes + 20 address bytes.
	expectedFrom := "0x" + common.Bytes2Hex(common.LeftPadBytes(sender.Bytes(), 32))
	expectedTo := "0x" + common.Bytes2Hex(common.LeftPadBytes(recipient.Bytes(), 32))
	suite.Require().Equal(expectedFrom, topics[1].(string))
	suite.Require().Equal(expectedTo, topics[2].(string))

	// Data: 32-byte big-endian value.
	expectedData := "0x" + common.Bytes2Hex(common.LeftPadBytes(big.NewInt(1_000_000).Bytes(), 32))
	suite.Require().Equal(expectedData, log["data"].(string))
}

// TestSimulateV1_TraceTransfers_Off_NoSyntheticLogs — explicit
// traceTransfers=false produces no synthetic logs.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_Off_NoSyntheticLogs() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
		"traceTransfers": false,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 1)

	logs := calls[0].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Empty(syntheticTransferLogs(logs),
		"traceTransfers=false must not produce synthetic ERC-7528 logs")
}

// TestSimulateV1_TraceTransfers_OmittedDefaultsToFalse — when the
// option is absent the driver behaves as if it were false: no synthetic
// logs.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_OmittedDefaultsToFalse() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	logs := calls[0].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Empty(syntheticTransferLogs(logs),
		"omitted traceTransfers must default to false")
}

// TestSimulateV1_TraceTransfers_On_ZeroValueCall_NoSyntheticLog —
// traceTransfers=true with a zero-value call must not emit a synthetic
// log. The deny-list and DELEGATECALL guards both sit behind the value
// > 0 check; this test pins the value-sign branch.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_ZeroValueCall_NoSyntheticLog() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient}},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	logs := calls[0].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Empty(logs, "zero-value call must produce no logs at all")
}

// TestSimulateV1_TraceTransfers_On_RealLogStillCaptured — a real EVM
// log emitted by user bytecode (LOG0 here) is captured into the call's
// log list when traceTransfers is on, even with no synthetic emission.
// Pins the StateDB.AddLog -> tracer.OnLog -> captured-log path.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_RealLogStillCaptured() {
	suite.SetupTest()

	sender := suite.address
	emitter := common.HexToAddress("0xbbbb000000000000000000000000000000000301")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:  {"balance": balance},
				emitter: {"code": log0Runtime},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &emitter}},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	logs := calls[0].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Len(logs, 1, "real LOG0 must surface even when no synthetic log fires")

	log := logs[0].(map[string]interface{})
	suite.Require().Equal(emitter, common.HexToAddress(log["address"].(string)),
		"captured real log must carry the emitter contract's address")
	suite.Require().Empty(syntheticTransferLogs(logs),
		"zero-value call must not produce any synthetic log")
}

// TestSimulateV1_TraceTransfers_On_NativeTransfer_BlockHashBackStamped —
// captured logs carry the per-block hash that the response envelope
// surfaces under `hash`. Only observable end-to-end because the driver
// records logs with a zero block hash and patches BlockHash + Index in
// a post-call back-stamp.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_NativeTransfer_BlockHashBackStamped() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	blockHash := results[0]["hash"].(string)
	suite.Require().NotEmpty(blockHash)

	calls := results[0]["calls"].([]interface{})
	logs := calls[0].(map[string]interface{})["logs"].([]interface{})
	synthetic := syntheticTransferLogs(logs)
	suite.Require().Len(synthetic, 1)
	suite.Require().Equal(blockHash, synthetic[0]["blockHash"].(string),
		"synthetic log blockHash must equal the assembled block's hash post-back-stamp")
}

// TestSimulateV1_TraceTransfers_On_RevertingCall_LogsDropped — a call
// that emits a real log and then reverts surfaces with `error.code = 3`
// and an empty logs list (the revert must drop both real and synthetic
// frame contents).
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_RevertingCall_LogsDropped() {
	suite.SetupTest()

	sender := suite.address
	contract := common.HexToAddress("0xbbbb000000000000000000000000000000000302")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:   {"balance": balance},
				contract: {"code": log0RevertRuntime},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &contract}},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	call := calls[0].(map[string]interface{})

	suite.Require().Equal("0x0", call["status"])
	errObj := call["error"].(map[string]interface{})
	suite.Require().EqualValues(float64(types.SimErrCodeReverted), errObj["code"])

	logs := call["logs"].([]interface{})
	suite.Require().Empty(logs, "revert must drop the real LOG0 the contract emitted before REVERT")
}

// TestSimulateV1_TraceTransfers_On_NestedCalls_OneSyntheticPerEdge — a
// top-level value-bearing CALL into a forwarder contract that
// re-CALLs the recipient with the same value produces TWO synthetic
// logs: outer (sender -> forwarder), inner (forwarder -> recipient).
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_NestedCalls_OneSyntheticPerEdge() {
	suite.SetupTest()

	sender := suite.address
	forwarder := common.HexToAddress("0xbbbb000000000000000000000000000000000303")
	recipient := common.HexToAddress("0x2222222222222222222222222222222222222222")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	value := (*hexutil.Big)(big.NewInt(123_456))
	calldata := common.LeftPadBytes(recipient.Bytes(), 32)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:    {"balance": balance},
				forwarder: {"code": valueForwarderRuntime},
			},
			"calls": []types.TransactionArgs{{
				From:  &sender,
				To:    &forwarder,
				Value: value,
				Input: (*hexutil.Bytes)(&calldata),
			}},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Equal("0x1", calls[0].(map[string]interface{})["status"])

	logs := calls[0].(map[string]interface{})["logs"].([]interface{})
	synthetic := syntheticTransferLogs(logs)
	suite.Require().Len(synthetic, 2,
		"forwarder pattern produces two value-transfer call edges (outer + inner)")

	outerTopics := synthetic[0]["topics"].([]interface{})
	innerTopics := synthetic[1]["topics"].([]interface{})

	expectedSender := "0x" + common.Bytes2Hex(common.LeftPadBytes(sender.Bytes(), 32))
	expectedForwarder := "0x" + common.Bytes2Hex(common.LeftPadBytes(forwarder.Bytes(), 32))
	expectedRecipient := "0x" + common.Bytes2Hex(common.LeftPadBytes(recipient.Bytes(), 32))

	suite.Require().Equal(expectedSender, outerTopics[1].(string))
	suite.Require().Equal(expectedForwarder, outerTopics[2].(string))
	suite.Require().Equal(expectedForwarder, innerTopics[1].(string))
	suite.Require().Equal(expectedRecipient, innerTopics[2].(string))
}

// TestSimulateV1_TraceTransfers_On_MultiCall_PerCallLogsIsolated — two
// value-transfer calls in the same block each surface their own
// synthetic log on the corresponding call result; neither call's log
// list contains the other's entries.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_MultiCall_PerCallLogsIsolated() {
	suite.SetupTest()

	sender := suite.address
	recipientA := common.HexToAddress("0xaaaa000000000000000000000000000000000401")
	recipientB := common.HexToAddress("0xbbbb000000000000000000000000000000000402")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipientA, Value: value},
				{From: &sender, To: &recipientB, Value: value},
			},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 2)

	logsA := calls[0].(map[string]interface{})["logs"].([]interface{})
	logsB := calls[1].(map[string]interface{})["logs"].([]interface{})

	syntheticA := syntheticTransferLogs(logsA)
	syntheticB := syntheticTransferLogs(logsB)
	suite.Require().Len(syntheticA, 1, "call 0's logs must contain only its own synthetic")
	suite.Require().Len(syntheticB, 1, "call 1's logs must contain only its own synthetic")

	expectedToA := "0x" + common.Bytes2Hex(common.LeftPadBytes(recipientA.Bytes(), 32))
	expectedToB := "0x" + common.Bytes2Hex(common.LeftPadBytes(recipientB.Bytes(), 32))
	suite.Require().Equal(expectedToA, syntheticA[0]["topics"].([]interface{})[2].(string))
	suite.Require().Equal(expectedToB, syntheticB[0]["topics"].([]interface{})[2].(string))
}

// TestSimulateV1_TraceTransfers_On_MultiCall_TxIndexMatchesCallPosition
// — transactionIndex on each captured synthetic log equals that call's
// position in the block's calls array.
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_MultiCall_TxIndexMatchesCallPosition() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xaaaa000000000000000000000000000000000403")
	value := (*hexutil.Big)(big.NewInt(1))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &sender, To: &recipient, Value: value},
				{From: &sender, To: &recipient, Value: value},
				{From: &sender, To: &recipient, Value: value},
			},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 3)

	for i, expected := range []string{"0x0", "0x1", "0x2"} {
		logs := calls[i].(map[string]interface{})["logs"].([]interface{})
		synthetic := syntheticTransferLogs(logs)
		suite.Require().Len(synthetic, 1)
		suite.Require().Equal(expected, synthetic[0]["transactionIndex"].(string),
			"call %d's synthetic log must carry transactionIndex=%s", i, expected)
	}
}

// TestSimulateV1_TraceTransfers_On_MultiBlock_LogIndexResetsPerBlock —
// per-block log index counter must reset between blocks. Block 0
// carries two value transfers (logIndex 0, 1); block 1 carries one
// (logIndex 0).
func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_MultiBlock_LogIndexResetsPerBlock() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xaaaa000000000000000000000000000000000404")
	value := (*hexutil.Big)(big.NewInt(1))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	transfer := types.TransactionArgs{From: &sender, To: &recipient, Value: value}

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender: {"balance": balance},
				},
				"calls": []types.TransactionArgs{transfer, transfer},
			},
			{"calls": []types.TransactionArgs{transfer}},
		},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 2)

	block0 := results[0]["calls"].([]interface{})
	block1 := results[1]["calls"].([]interface{})

	logs00 := syntheticTransferLogs(block0[0].(map[string]interface{})["logs"].([]interface{}))
	logs01 := syntheticTransferLogs(block0[1].(map[string]interface{})["logs"].([]interface{}))
	logs10 := syntheticTransferLogs(block1[0].(map[string]interface{})["logs"].([]interface{}))

	suite.Require().Len(logs00, 1)
	suite.Require().Len(logs01, 1)
	suite.Require().Len(logs10, 1)

	suite.Require().Equal("0x0", logs00[0]["logIndex"].(string),
		"block 0 call 0's synthetic log must be at logIndex 0")
	suite.Require().Equal("0x1", logs01[0]["logIndex"].(string),
		"block 0 call 1's synthetic log must be at logIndex 1 (cumulative within block)")
	suite.Require().Equal("0x0", logs10[0]["logIndex"].(string),
		"block 1's first synthetic log must reset the counter to 0")
}

func (suite *KeeperTestSuite) TestSimulateV1_TraceTransfers_On_BTCTokenSkipped() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	btcToken := common.HexToAddress(types.BTCTokenPrecompileAddress)

	// Fund the sender via bank directly — StateDB balance overrides
	// only touch the EVM state object and do not propagate to
	// bankKeeper, which is what btctoken.transfer actually reads.
	initial := sdkmath.NewInt(1_000_000_000_000_000_000)
	coin := sdk.NewCoin(types.DefaultEVMDenom, initial)
	suite.Require().NoError(suite.app.BankKeeper.MintCoins(
		suite.ctx, types.ModuleName, sdk.NewCoins(coin)))
	suite.Require().NoError(suite.app.BankKeeper.SendCoinsFromModuleToAccount(
		suite.ctx, types.ModuleName, sender.Bytes(), sdk.NewCoins(coin)))

	calldata := erc20TransferCalldata(recipient, big.NewInt(1))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"calls": []types.TransactionArgs{{
				From:  &sender,
				To:    &btcToken,
				Input: (*hexutil.Bytes)(&calldata),
			}},
		}},
		"traceTransfers": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	calls := results[0]["calls"].([]interface{})
	suite.Require().Equal("0x1", calls[0].(map[string]interface{})["status"],
		"btctoken.transfer must succeed when the sender's bank balance is funded")

	logs := calls[0].(map[string]interface{})["logs"].([]interface{})
	suite.Require().Empty(syntheticTransferLogs(logs),
		"BTC token precompile address must not produce a synthetic ERC-7528 log")

	// Real precompile-emitted Transfer event must still be present and
	// flow through the tracer. The address must be the precompile, the
	// topic must be the canonical ERC-20 Transfer signature.
	require := suite.Require()
	require.NotEmpty(logs, "BTC precompile transfer must surface its own real Transfer event")
	realLog := logs[0].(map[string]interface{})
	require.Equal(btcToken, common.HexToAddress(realLog["address"].(string)),
		"real Transfer event must carry the BTC precompile address as emitter")
	topics := realLog["topics"].([]interface{})
	require.NotEmpty(topics)
	require.Equal(erc20TransferTopic, topics[0].(string),
		"first topic must be the canonical ERC-20 Transfer signature")
}

// -----------------------------------------------------------------------------
// SimulateV1 — validation=true mode
// -----------------------------------------------------------------------------

func (suite *KeeperTestSuite) validatedSimulateRequest(
	state map[common.Address]map[string]interface{},
	calls []types.TransactionArgs,
	valFlag bool,
) []byte {
	suite.T().Helper()
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": state,
			"calls":          calls,
		}},
		"validation": valFlag,
	})
	suite.Require().NoError(err)
	return optsJSON
}

// validationMaxFeePerGas is large enough to clear the eip1559 baseFee
// floor that validation=true derives from the parent header in every
// fixture below; calls with this fee never hit the fee-cap gate so a
// test failure points at the gate it actually means to exercise.
var validationMaxFeePerGas = (*hexutil.Big)(big.NewInt(1_000_000_000_000)) // 1e12

// validationFundedBalance is large enough to cover any gasLimit *
// validationMaxFeePerGas + small `value` for every fixture. The few
// tests that need an under-funded sender override this field
// explicitly.
var validationFundedBalance = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(1_000_000_000), big.NewInt(1_000_000_000_000_000_000))) // 1e9 ether

// TestSimulateV1_Validation_HappyPath — funded sender, correct nonce,
// fee-cap above the chain-computed baseFee floor: the call must clear
// every gate.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_HappyPath() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": (*hexutil.Uint64)(nil)},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &recipient,
		Value:        (*hexutil.Big)(big.NewInt(1)),
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)
	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	c := results[0]["calls"].([]interface{})[0].(map[string]interface{})
	suite.Require().Equal("0x1", c["status"])
}

// TestSimulateV1_Validation_NonceLow — state.nonce=5, call.nonce=4 →
// top-level -38010.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_NonceLow() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	stateNonce := hexutil.Uint64(5)
	callNonce := hexutil.Uint64(4)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &recipient,
		Value:        (*hexutil.Big)(big.NewInt(1)),
		Nonce:        &callNonce,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeNonceTooLow), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_NonceHigh — state.nonce=5, call.nonce=9 →
// top-level -38011.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_NonceHigh() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	stateNonce := hexutil.Uint64(5)
	callNonce := hexutil.Uint64(9)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &recipient,
		Value:        (*hexutil.Big)(big.NewInt(1)),
		Nonce:        &callNonce,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeNonceTooHigh), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_InsufficientFunds — sender has zero balance,
// call carries non-zero value → -38014.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_InsufficientFunds() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	zero := (*hexutil.Big)(big.NewInt(0))
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": zero},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &recipient,
		Value:        (*hexutil.Big)(big.NewInt(1)),
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeInsufficientFunds), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_IntrinsicGas — call carries gas below the
// 21k pure-transfer intrinsic floor → -38013. Same surface as the
// existing TestSimulateV1_ExplicitGasBelowIntrinsic, lifted under
// validation=true to confirm the gate fires there too rather than
// changing the request-level fatal code.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_IntrinsicGas() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	tooLowGas := hexutil.Uint64(20_999)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &recipient,
		Value:        (*hexutil.Big)(big.NewInt(1)),
		Gas:          &tooLowGas,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeIntrinsicGas), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_InitCodeTooLarge — CREATE call (To=nil)
// with init-code 1 byte over MaxInitCodeSize → -38025. Per EIP-3860 the
// gate is Shanghai-active in the test app's default chain config.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_InitCodeTooLarge() {
	suite.SetupTest()

	sender := suite.address
	overSize := make([]byte, ethparams.MaxInitCodeSize+1)
	// 1M gas covers the intrinsic-gas + EIP-3860 word cost for an
	// all-zero init-code payload at MaxInitCodeSize+1 bytes (~232k).
	gas := hexutil.Uint64(1_000_000)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance},
	}
	data := hexutil.Bytes(overSize)
	calls := []types.TransactionArgs{{
		From:         &sender,
		Data:         &data,
		Gas:          &gas,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeMaxInitCodeSizeExceeded), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_FeeCapBelowBaseFee — explicit MaxFeePerGas=0
// while validation=true forces a non-zero baseFee on the synthesized
// header → -32005.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_FeeCapBelowBaseFee() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	zeroFee := (*hexutil.Big)(big.NewInt(0))
	// Ensure header.BaseFee > 0 by overriding the block's baseFee. The
	// override goes through the eip1559 floor check, which requires the
	// override to be >= the chain-computed floor; setting the override
	// to a small positive value above the floor (1 wei is below, so we
	// pick a larger number) keeps that gate happy while exercising the
	// per-call fee-cap gate.
	baseFee := (*hexutil.Big)(big.NewInt(1_000_000_000))
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"blockOverrides": map[string]interface{}{"baseFeePerGas": baseFee},
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": validationFundedBalance},
			},
			"calls": []types.TransactionArgs{{
				From:                 &sender,
				To:                   &recipient,
				Value:                (*hexutil.Big)(big.NewInt(1)),
				MaxFeePerGas:         zeroFee,
				MaxPriorityFeePerGas: zeroFee,
			}},
		}},
		"validation": true,
	})
	suite.Require().NoError(err)
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeFeeCapTooLow), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_BaseFeeOverrideTooLow — caller-supplied
// blockOverrides.baseFeePerGas BELOW the chain-computed eip1559 floor
// → -38012. Per the spec, the per-block check fires before any per-call
// gate even runs.
//
// Setup note: the keeper test app defaults to feemarket=disabled, which
// pins the chain's BaseFee at 0; with parent.BaseFee=0 the eip1559 floor
// formula collapses to 0, leaving no room to assert "override below
// floor". We flip feemarket on for this case so the parent header
// surfaces a real InitialBaseFee (1 gwei), which the floor formula then
// adjusts. Restored to default afterwards.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_BaseFeeOverrideTooLow() {
	suite.enableFeemarket = true
	defer func() { suite.enableFeemarket = false }()
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	// 1 wei sits far below any feemarket-derived floor (which starts at
	// 1 gwei minus an elasticity adjustment, tens of millions of wei
	// even after a single decrease).
	farTooLow := (*hexutil.Big)(big.NewInt(1))
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"blockOverrides": map[string]interface{}{"baseFeePerGas": farTooLow},
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": validationFundedBalance},
			},
			"calls": []types.TransactionArgs{{
				From:         &sender,
				To:           &recipient,
				Value:        (*hexutil.Big)(big.NewInt(1)),
				MaxFeePerGas: validationMaxFeePerGas,
			}},
		}},
		"validation": true,
	})
	suite.Require().NoError(err)
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeBaseFeeTooLow), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_NodeNoBaseFeeIgnored — chain-level
// feemarket.NoBaseFee=true does NOT relax the validation gate's fee-cap
// check. Per spec, validation=true is an authoritative override; node
// config cannot opt out of the realistic preflight.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_NodeNoBaseFeeIgnored() {
	suite.enableFeemarket = false
	defer func() { suite.enableFeemarket = false }()
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	zeroFee := (*hexutil.Big)(big.NewInt(0))
	baseFee := (*hexutil.Big)(big.NewInt(1_000_000_000))
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"blockOverrides": map[string]interface{}{"baseFeePerGas": baseFee},
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": validationFundedBalance},
			},
			"calls": []types.TransactionArgs{{
				From:                 &sender,
				To:                   &recipient,
				Value:                (*hexutil.Big)(big.NewInt(1)),
				MaxFeePerGas:         zeroFee,
				MaxPriorityFeePerGas: zeroFee,
			}},
		}},
		"validation": true,
	})
	suite.Require().NoError(err)
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeFeeCapTooLow), resp.Error.Code,
		"validation=true must enforce base fee even when node-level NoBaseFee is set")
}

// TestSimulateV1_Validation_RevertStaysPerCall — a reverting call is a
// per-call failure (CallResultFailure.error.code = 3), NOT a fatal abort.
// Load-bearing: revert is the most common per-call outcome a caller
// observes, and the validation gate must not promote it.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_RevertStaysPerCall() {
	suite.SetupTest()

	sender := suite.address
	// Bytecode: PUSH1 0x00 PUSH1 0x00 REVERT — unconditional revert
	// with empty return data.
	revertCode := common.Hex2Bytes("60006000FD")
	revertContract := common.HexToAddress("0xddee000000000000000000000000000000000000")

	state := map[common.Address]map[string]interface{}{
		sender:         {"balance": validationFundedBalance},
		revertContract: {"code": hexutil.Bytes(revertCode)},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &revertContract,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error, "reverting call must NOT abort the request under validation=true")
	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	c := results[0]["calls"].([]interface{})[0].(map[string]interface{})
	suite.Require().Equal("0x0", c["status"])
	cErr := c["error"].(map[string]interface{})
	suite.Require().Equal(float64(types.SimErrCodeReverted), cErr["code"])
}

// TestSimulateV1_Validation_VMErrorStaysPerCall — a non-revert VM
// failure (e.g. invalid opcode) is per-call (-32015), not a fatal
// abort.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_VMErrorStaysPerCall() {
	suite.SetupTest()

	sender := suite.address
	// 0xFE is the canonical INVALID opcode — runs OOG with all gas
	// consumed.
	invalidCode := common.Hex2Bytes("FE")
	invalidContract := common.HexToAddress("0xddff000000000000000000000000000000000000")

	state := map[common.Address]map[string]interface{}{
		sender:          {"balance": validationFundedBalance},
		invalidContract: {"code": hexutil.Bytes(invalidCode)},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &invalidContract,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error, "VM error must NOT abort the request under validation=true")
	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	c := results[0]["calls"].([]interface{})[0].(map[string]interface{})
	suite.Require().Equal("0x0", c["status"])
	cErr := c["error"].(map[string]interface{})
	suite.Require().Equal(float64(types.SimErrCodeVMError), cErr["code"])
}

// TestSimulateV1_Validation_AbortsOnFirstCallSecondNotRun — call[0]
// fails the gate; call[1] would succeed if reached. The fatal abort
// must short-circuit so call[1] is never executed (verified by
// observing that the response is bare-error, not a partial result
// envelope).
func (suite *KeeperTestSuite) TestSimulateV1_Validation_AbortsOnFirstCallSecondNotRun() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	stateNonce := hexutil.Uint64(5)
	tooLowNonce := hexutil.Uint64(0)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
	}
	calls := []types.TransactionArgs{
		// call[0] — nonce too low → fatal -38010
		{From: &sender, To: &recipient, Nonce: &tooLowNonce, MaxFeePerGas: validationMaxFeePerGas},
		// call[1] — would otherwise succeed
		{From: &sender, To: &recipient, MaxFeePerGas: validationMaxFeePerGas},
	}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeNonceTooLow), resp.Error.Code)
	suite.Require().Empty(resp.Result, "fatal abort must not emit a partial result envelope")
}

// TestSimulateV1_Validation_AbortsOnSecondBlockFirstAlreadyExecuted —
// block[0] runs to completion; block[1] fails the gate on its first
// call. The whole request aborts; no result envelope is emitted (the
// driver does not partially commit prior blocks on a fatal). Pinned
// because partial commits would be a correctness regression — callers
// rely on simulate being all-or-nothing.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_AbortsOnSecondBlockFirstAlreadyExecuted() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	stateNonce := hexutil.Uint64(5)
	tooLowNonce := hexutil.Uint64(0)
	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
				},
				"calls": []types.TransactionArgs{{
					From: &sender, To: &recipient, MaxFeePerGas: validationMaxFeePerGas,
				}},
			},
			{
				"calls": []types.TransactionArgs{{
					// Nonce reset to a value below the post-block-0 state
					// nonce (which advanced from 5 -> 6 after the CALL).
					From: &sender, To: &recipient, Nonce: &tooLowNonce,
					MaxFeePerGas: validationMaxFeePerGas,
				}},
			},
		},
		"validation": true,
	})
	suite.Require().NoError(err)
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeNonceTooLow), resp.Error.Code)
	suite.Require().Empty(resp.Result)
}

// TestSimulateV1_Validation_DeterminismRepeated — same opts run twice
// must surface the same fatal SimError byte-for-byte. Pinned to guard
// against time-dependent or PRNG-tainted error messages slipping into
// the gate path.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_DeterminismRepeated() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	stateNonce := hexutil.Uint64(5)
	tooLowNonce := hexutil.Uint64(0)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
	}
	calls := []types.TransactionArgs{{
		From: &sender, To: &recipient, Nonce: &tooLowNonce,
		MaxFeePerGas: validationMaxFeePerGas,
	}}

	resp1, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp1.Error)
	resp2, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp2.Error)
	suite.Require().Equal(resp1.Error.Code, resp2.Error.Code)
	suite.Require().Equal(resp1.Error.Message, resp2.Error.Message)
}

// --- validation=false (default) — every gate must allow the call ---------

func (suite *KeeperTestSuite) TestSimulateV1_NoValidation_NonceLowSucceeds() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	stateNonce := hexutil.Uint64(5)
	tooLowNonce := hexutil.Uint64(0)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
	}
	calls := []types.TransactionArgs{{
		From: &sender, To: &recipient, Nonce: &tooLowNonce,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, false)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)
	results := suite.simulateV1BlockResults(resp)
	c := results[0]["calls"].([]interface{})[0].(map[string]interface{})
	suite.Require().Equal("0x1", c["status"])
}

// validation=false must not promote insufficient-funds to -38014.
// Mezod's EVM still rejects value transfers via CanTransfer, so the
// call surfaces as a per-call failure — the request-level pin is the
// absence of a top-level fatal.
func (suite *KeeperTestSuite) TestSimulateV1_NoValidation_InsufficientFundsSucceeds() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	zero := (*hexutil.Big)(big.NewInt(0))
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": zero},
	}
	calls := []types.TransactionArgs{{
		From: &sender, To: &recipient, Value: (*hexutil.Big)(big.NewInt(1)),
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, false)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error,
		"validation=false must NOT promote insufficient-funds to a top-level -38014 fatal")
	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	suite.Require().Len(results[0]["calls"].([]interface{}), 1,
		"the call must show up in the per-call list, not collapse into the request error")
}

func (suite *KeeperTestSuite) TestSimulateV1_NoValidation_FeeCapBelowBaseFeeSucceeds() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	zero := (*hexutil.Big)(big.NewInt(0))
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance},
	}
	calls := []types.TransactionArgs{{
		From: &sender, To: &recipient,
		Value:                (*hexutil.Big)(big.NewInt(1)),
		MaxFeePerGas:         zero,
		MaxPriorityFeePerGas: zero,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, false)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)
	results := suite.simulateV1BlockResults(resp)
	c := results[0]["calls"].([]interface{})[0].(map[string]interface{})
	suite.Require().Equal("0x1", c["status"])
}

// TestSimulateV1_NoValidation_InitCodeOverLimitSucceeds — CREATE with
// init-code 1 byte over MaxInitCodeSize MUST NOT trip the request-level
// -38025 fatal under validation=false. The per-call result is whatever
// the EVM emits (in practice geth's CREATE-handler EIP-3860 enforcement
// surfaces as a per-call ErrMaxInitCodeSizeExceeded → the SimError
// formatter routes it to per-call code 3 reverted-equivalent or
// -32015 VM-error). Either is a *per-call* outcome, not a top-level
// fatal — that distinction is what this test pins.
func (suite *KeeperTestSuite) TestSimulateV1_NoValidation_InitCodeOverLimitSucceeds() {
	suite.SetupTest()

	sender := suite.address
	overSize := make([]byte, ethparams.MaxInitCodeSize+1)
	// 1M gas covers the intrinsic-gas + EIP-3860 word cost for an
	// all-zero init-code payload at MaxInitCodeSize+1 bytes (~232k).
	gas := hexutil.Uint64(1_000_000)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance},
	}
	data := hexutil.Bytes(overSize)
	calls := []types.TransactionArgs{{
		From: &sender, Data: &data, Gas: &gas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, false)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error,
		"validation=false must NOT promote oversize init-code to a top-level -38025 fatal")
	// The per-call result still surfaces — exact code depends on the
	// EVM's CREATE handling, but it is NOT a top-level fatal.
	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	suite.Require().Len(results[0]["calls"].([]interface{}), 1,
		"the call must show up in the per-call list, not collapse into the request error")
}

// TestSimulateV1_Validation_BoundaryEqual_NoFatal — every gate sitting
// exactly on the pass-side of its boundary at the same time must clear
// the gate. Single-fail tests verify each gate's "one-step-over"
// failure; this one asserts the gate's pass-side composition holds
// when all five are evaluated at their respective edges.
func (suite *KeeperTestSuite) TestSimulateV1_Validation_BoundaryEqual_NoFatal() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0x1111111111111111111111111111111111111111")
	gas := hexutil.Uint64(21_000) // intrinsic floor for pure transfer
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance},
	}
	calls := []types.TransactionArgs{{
		From:         &sender,
		To:           &recipient,
		Value:        (*hexutil.Big)(big.NewInt(1)),
		Gas:          &gas,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error, "every gate at boundary-equal must clear")
	results := suite.simulateV1BlockResults(resp)
	c := results[0]["calls"].([]interface{})[0].(map[string]interface{})
	suite.Require().Equal("0x1", c["status"])
}

func (suite *KeeperTestSuite) TestSimulateV1_Validation_NonceLow_StateOverrideNonceAborts() {
	suite.SetupTest()

	sender := common.HexToAddress("0xc100000000000000000000000000000000000000")
	recipient := common.HexToAddress("0xc100000000000000000000000000000000000000")
	stateNonce := hexutil.Uint64(0xa)
	callNonce := hexutil.Uint64(0)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
	}
	calls := []types.TransactionArgs{{
		From: &sender, To: &recipient, Nonce: &callNonce,
		MaxFeePerGas: validationMaxFeePerGas,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error)
	suite.Require().Equal(int32(types.SimErrCodeNonceTooLow), resp.Error.Code)
}

func (suite *KeeperTestSuite) TestSimulateV1_Validation_NonceLow_StateOverrideNonceSucceedsWhenOff() {
	suite.SetupTest()

	sender := common.HexToAddress("0xc100000000000000000000000000000000000000")
	recipient := common.HexToAddress("0xc100000000000000000000000000000000000000")
	stateNonce := hexutil.Uint64(0xa)
	callNonce := hexutil.Uint64(0)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
	}
	calls := []types.TransactionArgs{{
		From: &sender, To: &recipient, Nonce: &callNonce,
	}}
	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, false)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)
	results := suite.simulateV1BlockResults(resp)
	c := results[0]["calls"].([]interface{})[0].(map[string]interface{})
	suite.Require().Equal("0x1", c["status"])
}

// -----------------------------------------------------------------------------
// SimulateV1 — block envelope
// -----------------------------------------------------------------------------

// envelopeStandardOpts builds the simulate opts JSON for a single
// successful native transfer — the workhorse fixture for envelope
// assertions where the per-call shape is uninteresting.
func (suite *KeeperTestSuite) envelopeStandardOpts() []byte {
	suite.T().Helper()
	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000500")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
	})
	suite.Require().NoError(err)
	return optsJSON
}

// TestSimulateV1_BlockEnvelope_PopulatesAllHeaderFields — every
// canonical RPC envelope key surfaces on the response, sourced from
// types.NewSimBlockResult applied to the assembled *ethtypes.Block.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_PopulatesAllHeaderFields() {
	suite.SetupTest()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.envelopeStandardOpts()))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	for _, key := range []string{
		"number", "hash", "parentHash", "logsBloom", "stateRoot",
		"miner", "difficulty", "extraData", "gasLimit", "gasUsed",
		"timestamp", "transactionsRoot", "receiptsRoot", "size",
		"transactions", "uncles", "calls",
	} {
		suite.Require().Contains(results[0], key, "envelope must contain %q", key)
	}
}

// TestSimulateV1_BlockEnvelope_TxRootNonEmpty — block carrying a
// successful tx surfaces a non-empty transactionsRoot, computed via
// DeriveSha over the synthetic txs.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_TxRootNonEmpty() {
	suite.SetupTest()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.envelopeStandardOpts()))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	root := results[0]["transactionsRoot"].(string)
	suite.Require().NotEqual(ethtypes.EmptyTxsHash.Hex(), root,
		"transactionsRoot for a non-empty block must not be the empty-trie root")
}

// TestSimulateV1_BlockEnvelope_ReceiptsRootNonEmpty — analogous to
// TxRootNonEmpty but on the receipts side. A successful call yields a
// non-empty receipts root.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_ReceiptsRootNonEmpty() {
	suite.SetupTest()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.envelopeStandardOpts()))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	root := results[0]["receiptsRoot"].(string)
	suite.Require().NotEqual(ethtypes.EmptyReceiptsHash.Hex(), root,
		"receiptsRoot for a non-empty block must not be the empty-trie root")
}

// TestSimulateV1_BlockEnvelope_BloomMatchesLogs — when a call emits a
// real LOG with a known topic, the assembled block's logsBloom must
// affirmative-test that topic via BloomLookup.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_BloomMatchesLogs() {
	suite.SetupTest()

	sender := suite.address
	emitter := common.HexToAddress("0xbbbb000000000000000000000000000000000501")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	// log0Runtime emits LOG0 (no topics) — but we need a topic for the
	// bloom check. Build a contract that emits LOG1 with a known topic.
	//
	//   PUSH32 <topic>     ; topic
	//   PUSH1  0x00        ; size
	//   PUSH1  0x00        ; offset
	//   LOG1
	//   STOP
	topic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	log1Runtime := "0x7f" + topic.Hex()[2:] + "60006000A100"

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:  {"balance": balance},
				emitter: {"code": log1Runtime},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &emitter}},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	bloomStr := results[0]["logsBloom"].(string)
	bloom := ethtypes.BytesToBloom(common.FromHex(bloomStr))
	suite.Require().True(bloom.Test(topic.Bytes()),
		"block bloom must affirmative-test the LOG1 topic")
}

// TestSimulateV1_BlockEnvelope_SizeNonZero — `size` is non-zero on
// any sealed envelope (header alone is several dozen bytes).
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_SizeNonZero() {
	suite.SetupTest()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.envelopeStandardOpts()))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	sizeStr := results[0]["size"].(string)
	sizeVal := new(big.Int)
	_, ok := sizeVal.SetString(sizeStr[2:], 16)
	suite.Require().True(ok)
	suite.Require().Equal(1, sizeVal.Sign(), "size must be positive")
}

// TestSimulateV1_BlockEnvelope_StateRootZero — pinned divergence:
// stateRoot is always the zero hash on a Mezo simulate envelope
// (no MPT to derive an intermediate root from).
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_StateRootZero() {
	suite.SetupTest()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.envelopeStandardOpts()))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	suite.Require().Equal(common.Hash{}.Hex(), results[0]["stateRoot"].(string),
		"stateRoot must be the zero hash (Mezo divergence)")
}

// TestSimulateV1_BlockEnvelope_EmptyBlockUsesEmptyRoots — a gap-fill
// block (one with no calls inserted by sanitizeSimChain to honor a
// non-contiguous block-number override) surfaces the canonical
// transactionsRoot/receiptsRoot for an empty trie.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_EmptyBlockUsesEmptyRoots() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000502")
	value := (*hexutil.Big)(big.NewInt(1))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	// Block 1 is implicit (number = base+1, has the call). Block 2
	// has no override so it's also a normal sim block. The gap-fill
	// is triggered by an explicit jump to base+5; sanitizeSimChain
	// inserts empty blocks at base+2..base+4.
	jumpNumber := (*hexutil.Big)(new(big.Int).Add(big.NewInt(suite.ctx.BlockHeight()), big.NewInt(5)))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender: {"balance": balance},
				},
				"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
			},
			{
				"blockOverrides": map[string]interface{}{"number": jumpNumber},
				"calls":          []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
			},
		},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	// 1 explicit + 3 gap-fills + 1 explicit = 5 results.
	suite.Require().Len(results, 5)

	// Inspect a gap-fill (results[1] through results[3]). They must
	// have empty calls + empty roots.
	for i := 1; i <= 3; i++ {
		gap := results[i]
		calls := gap["calls"].([]interface{})
		suite.Require().Empty(calls, "gap-fill block %d must have empty calls", i)
		suite.Require().Equal(ethtypes.EmptyTxsHash.Hex(), gap["transactionsRoot"].(string),
			"gap-fill block %d must use the empty txs root", i)
		suite.Require().Equal(ethtypes.EmptyReceiptsHash.Hex(), gap["receiptsRoot"].(string),
			"gap-fill block %d must use the empty receipts root", i)
		txs := gap["transactions"].([]interface{})
		suite.Require().Empty(txs, "gap-fill block %d transactions must be []", i)
	}
}

// TestSimulateV1_BlockEnvelope_HashStable — running the same opts
// twice yields the same block hash on the corresponding envelope.
// Pinned because the simulate driver synthesizes hashes from header
// fields; any non-determinism (timestamp drift, randomness leaking in)
// would break this.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_HashStable() {
	suite.SetupTest()

	opts := suite.envelopeStandardOpts()

	resp1, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(opts))
	suite.Require().NoError(err)
	resp2, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(opts))
	suite.Require().NoError(err)

	r1 := suite.simulateV1BlockResults(resp1)
	r2 := suite.simulateV1BlockResults(resp2)
	suite.Require().Len(r1, 1)
	suite.Require().Len(r2, 1)
	suite.Require().Equal(r1[0]["hash"], r2[0]["hash"],
		"identical opts must produce identical block hash (deterministic envelope)")
}

// TestSimulateV1_BlockEnvelope_GasUsedAggregates — block.gasUsed
// equals the sum of per-call gasUsed values.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_GasUsedAggregates() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000503")
	value := (*hexutil.Big)(big.NewInt(1))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	transfer := types.TransactionArgs{From: &sender, To: &recipient, Value: value}

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{transfer, transfer, transfer},
		}},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	calls := results[0]["calls"].([]interface{})
	var sum uint64
	for _, c := range calls {
		gasUsed := c.(map[string]interface{})["gasUsed"].(string)
		v, ok := new(big.Int).SetString(gasUsed[2:], 16)
		suite.Require().True(ok)
		sum += v.Uint64()
	}
	blockGasUsedStr := results[0]["gasUsed"].(string)
	blockGasUsed, ok := new(big.Int).SetString(blockGasUsedStr[2:], 16)
	suite.Require().True(ok)
	suite.Require().Equal(sum, blockGasUsed.Uint64(),
		"block.gasUsed must equal sum of per-call gasUsed")
}

// TestSimulateV1_BlockEnvelope_MultiBlock_PerBlockBloomIsolated — a
// log emitted in block 0 must not leak into block 1's bloom. Pins the
// per-block isolation contract.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_MultiBlock_PerBlockBloomIsolated() {
	suite.SetupTest()

	sender := suite.address
	emitter := common.HexToAddress("0xbbbb000000000000000000000000000000000504")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	topic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	log1Runtime := "0x7f" + topic.Hex()[2:] + "60006000A100"

	emit := types.TransactionArgs{From: &sender, To: &emitter}
	recipient := common.HexToAddress("0xcccc000000000000000000000000000000000505")
	plain := types.TransactionArgs{From: &sender, To: &recipient, Value: (*hexutil.Big)(big.NewInt(1))}

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{
			{
				"stateOverrides": map[common.Address]map[string]interface{}{
					sender:  {"balance": balance},
					emitter: {"code": log1Runtime},
				},
				"calls": []types.TransactionArgs{emit},
			},
			{"calls": []types.TransactionArgs{plain}},
		},
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 2)

	bloom0 := ethtypes.BytesToBloom(common.FromHex(results[0]["logsBloom"].(string)))
	bloom1 := ethtypes.BytesToBloom(common.FromHex(results[1]["logsBloom"].(string)))

	suite.Require().True(bloom0.Test(topic.Bytes()), "block 0 bloom must carry the LOG1 topic")
	suite.Require().False(bloom1.Test(topic.Bytes()),
		"block 1 bloom must NOT carry block 0's topic — per-block bloom is isolated")
}

// TestSimulateV1_FullTx_HashOnly_DefaultBehavior — opts without
// returnFullTransactions surface `transactions` as a list of hash
// strings (66 hex chars each).
func (suite *KeeperTestSuite) TestSimulateV1_FullTx_HashOnly_DefaultBehavior() {
	suite.SetupTest()

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(suite.envelopeStandardOpts()))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 1)

	hashStr, ok := txs[0].(string)
	suite.Require().True(ok, "default returnFullTransactions must yield hash strings, got %T", txs[0])
	suite.Require().Len(hashStr, 66, "tx hash must be 66 hex chars (0x + 32 bytes)")
}

// TestSimulateV1_FullTx_FromPatched — returnFullTransactions=true
// emits transaction objects whose `from` is patched from the Senders
// map. Without the patch the unsigned-tx ECRECOVER returns the zero
// address.
func (suite *KeeperTestSuite) TestSimulateV1_FullTx_FromPatched() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000506")
	value := (*hexutil.Big)(big.NewInt(1_000_000))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
		"returnFullTransactions": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 1)

	tx, ok := txs[0].(map[string]interface{})
	suite.Require().True(ok, "returnFullTransactions=true must yield tx objects, got %T", txs[0])
	suite.Require().Equal(sender.Hex(), common.HexToAddress(tx["from"].(string)).Hex(),
		"transactions[0].from must be patched from Senders map")
}

// TestSimulateV1_FullTx_FromPatched_MultipleSendersInOneBlock — two
// calls from two different senders in one block; each tx's `from`
// must resolve to its own sender. Match by tx hash to guard against
// index-based confusion.
//
// Distinct values are required: synthetic txs are unsigned legacy
// txs whose hash does not depend on the sender, so two identical
// (nonce, to, value, data, gas, gasPrice) tuples from different
// senders would collide in the Senders-by-hash map. The values
// below are deliberately distinct to keep the assertion meaningful.
func (suite *KeeperTestSuite) TestSimulateV1_FullTx_FromPatched_MultipleSendersInOneBlock() {
	suite.SetupTest()

	senderA := suite.address
	senderB := common.HexToAddress("0xa2b30000000000000000000000000000000a2b30")
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000507")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				senderA: {"balance": balance},
				senderB: {"balance": balance},
			},
			"calls": []types.TransactionArgs{
				{From: &senderA, To: &recipient, Value: (*hexutil.Big)(big.NewInt(111))},
				{From: &senderB, To: &recipient, Value: (*hexutil.Big)(big.NewInt(222))},
			},
		}},
		"returnFullTransactions": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 2)

	gotFroms := make(map[string]struct{})
	for _, raw := range txs {
		tx := raw.(map[string]interface{})
		gotFroms[common.HexToAddress(tx["from"].(string)).Hex()] = struct{}{}
	}
	suite.Require().Contains(gotFroms, senderA.Hex())
	suite.Require().Contains(gotFroms, senderB.Hex())
}

// TestSimulateV1_FullTx_RevertedCallStillIncludedAsTx — a reverted
// call still appears in `transactions[]` (the synthetic tx is
// emitted regardless of execution outcome) and its `from` is
// patched.
func (suite *KeeperTestSuite) TestSimulateV1_FullTx_RevertedCallStillIncludedAsTx() {
	suite.SetupTest()

	sender := suite.address
	revertContract := common.HexToAddress("0xbbbb000000000000000000000000000000000508")
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	// PUSH1 0x00 PUSH1 0x00 REVERT.
	revertCode := common.Hex2Bytes("60006000FD")

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender:         {"balance": balance},
				revertContract: {"code": hexutil.Bytes(revertCode)},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &revertContract}},
		}},
		"returnFullTransactions": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	calls := results[0]["calls"].([]interface{})
	suite.Require().Len(calls, 1)
	suite.Require().Equal("0x0", calls[0].(map[string]interface{})["status"],
		"reverted call must surface status 0")

	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 1, "reverted call's synthetic tx must still appear in transactions[]")
	tx := txs[0].(map[string]interface{})
	suite.Require().Equal(sender.Hex(), common.HexToAddress(tx["from"].(string)).Hex())
}

// TestSimulateV1_FullTx_PreflightFailedCallNotInTransactions — a call
// that fails preflight (e.g. validation=true with nonce-too-low)
// aborts the entire request before assembling the block; no
// transactions[] envelope is emitted at all.
func (suite *KeeperTestSuite) TestSimulateV1_FullTx_PreflightFailedCallNotInTransactions() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb000000000000000000000000000000000509")
	stateNonce := hexutil.Uint64(5)
	tooLowNonce := hexutil.Uint64(0)

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": validationFundedBalance, "nonce": &stateNonce},
			},
			"calls": []types.TransactionArgs{{
				From: &sender, To: &recipient, Nonce: &tooLowNonce,
				MaxFeePerGas: validationMaxFeePerGas,
			}},
		}},
		"validation":             true,
		"returnFullTransactions": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Error,
		"preflight-failed validation must abort the request as a top-level fatal")
	suite.Require().Empty(resp.Result,
		"fatal abort must not emit any result envelope (so nothing claims the call ran)")
}

// recomputeTxRoot independently computes the transactions trie root
// for a single legacy unsigned tx whose fields mirror what
// simulate's buildSimTx synthesizes — used to cross-check the
// driver's transactionsRoot against an out-of-band derivation.
func recomputeTxRoot(tx *ethtypes.Transaction) common.Hash {
	return ethtypes.DeriveSha(ethtypes.Transactions{tx}, trie.NewStackTrie(nil))
}

// TestSimulateV1_BlockEnvelope_TxRootMatchesIndependentRecompute —
// pin the transactionsRoot against a freshly-built DeriveSha over
// an equivalent synthetic tx. Anchors the driver to the upstream
// hashing algorithm so any drift surfaces immediately.
//
// A bare (no fee fields) request resolves to a DynamicFeeTx envelope
// per geth's eth_simulateV1 default-type rule — see buildSimTx for the
// shape selection. Earlier revisions of this test pinned a LegacyTx;
// the move to DynamicFeeTx happened together with the buildSimTx fix
// that emits typed envelopes per request shape.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_TxRootMatchesIndependentRecompute() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb00000000000000000000000000000000050a")
	value := (*hexutil.Big)(big.NewInt(1))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value}},
		}},
		"returnFullTransactions": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)

	// Lift the actual tx fields from the response so the recomputed
	// trie covers the exact tx the driver synthesized — buildSimTx's
	// nonce default (StateDB.GetNonce) and gas resolver are too
	// coupled to the keeper to safely reconstruct here.
	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 1)
	txObj := txs[0].(map[string]interface{})

	suite.Require().Equal("0x2", txObj["type"].(string),
		"bare-default request must surface as a DynamicFeeTx envelope (type 2)")

	wantTx := ethtypes.NewTx(&ethtypes.DynamicFeeTx{
		ChainID:    mustHexBig(txObj["chainId"].(string)),
		Nonce:      mustHexUint64(txObj["nonce"].(string)),
		GasTipCap:  mustHexBig(txObj["maxPriorityFeePerGas"].(string)),
		GasFeeCap:  mustHexBig(txObj["maxFeePerGas"].(string)),
		Gas:        mustHexUint64(txObj["gas"].(string)),
		To:         &recipient,
		Value:      mustHexBig(txObj["value"].(string)),
		Data:       common.FromHex(txObj["input"].(string)),
		AccessList: ethtypes.AccessList{},
	})

	// Sanity-check: our reconstructed tx hash matches what the
	// driver published — if it doesn't, the trie comparison below
	// covers the wrong thing.
	suite.Require().Equal(txObj["hash"].(string), wantTx.Hash().Hex(),
		"reconstructed tx must hash to what the response carries")

	want := recomputeTxRoot(wantTx)
	suite.Require().Equal(want.Hex(), results[0]["transactionsRoot"].(string),
		"transactionsRoot must equal DeriveSha over the synthetic tx")
}

// TestSimulateV1_BlockEnvelope_LegacyShapeWhenGasPriceSet — when a
// caller specifies the legacy `gasPrice` field, geth's ToTransaction
// force-overrides the envelope back to LegacyTx regardless of the
// simulator's defaultType. Pin that behavior so a future buildSimTx
// rewrite can't silently drop it.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_LegacyShapeWhenGasPriceSet() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb00000000000000000000000000000000050b")
	value := (*hexutil.Big)(big.NewInt(1))
	gasPrice := (*hexutil.Big)(big.NewInt(0))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{From: &sender, To: &recipient, Value: value, GasPrice: gasPrice}},
		}},
		"returnFullTransactions": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 1)
	txObj := txs[0].(map[string]interface{})

	suite.Require().Equal("0x0", txObj["type"].(string),
		"explicit gasPrice must force a LegacyTx envelope")
	suite.Require().Nil(txObj["maxFeePerGas"],
		"LegacyTx must not carry maxFeePerGas")
	suite.Require().Nil(txObj["accessList"],
		"LegacyTx must not carry accessList")

	wantTx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    mustHexUint64(txObj["nonce"].(string)),
		GasPrice: mustHexBig(txObj["gasPrice"].(string)),
		Gas:      mustHexUint64(txObj["gas"].(string)),
		To:       &recipient,
		Value:    mustHexBig(txObj["value"].(string)),
		Data:     common.FromHex(txObj["input"].(string)),
	})
	suite.Require().Equal(txObj["hash"].(string), wantTx.Hash().Hex(),
		"reconstructed legacy tx must match the published hash")
	suite.Require().Equal(recomputeTxRoot(wantTx).Hex(), results[0]["transactionsRoot"].(string))
}

// TestSimulateV1_BlockEnvelope_DynamicFeeShapeWithAccessList —
// access-list and dynamic-fee fields combined produce a DynamicFeeTx
// envelope that nests the access list. Mirrors geth eth_simulateV1's
// behavior under defaultType=DynamicFeeTxType, where the AccessList
// case in ToTransaction's switch is unreachable except via the
// default-type fall-through.
func (suite *KeeperTestSuite) TestSimulateV1_BlockEnvelope_DynamicFeeShapeWithAccessList() {
	suite.SetupTest()

	sender := suite.address
	recipient := common.HexToAddress("0xbbbb00000000000000000000000000000000050c")
	value := (*hexutil.Big)(big.NewInt(1))
	maxFee := (*hexutil.Big)(big.NewInt(100))
	maxPrio := (*hexutil.Big)(big.NewInt(2))
	balance := (*hexutil.Big)(big.NewInt(1_000_000_000_000_000_000))
	emptyAL := ethtypes.AccessList{}

	optsJSON, err := json.Marshal(map[string]interface{}{
		"blockStateCalls": []map[string]interface{}{{
			"stateOverrides": map[common.Address]map[string]interface{}{
				sender: {"balance": balance},
			},
			"calls": []types.TransactionArgs{{
				From:                 &sender,
				To:                   &recipient,
				Value:                value,
				MaxFeePerGas:         maxFee,
				MaxPriorityFeePerGas: maxPrio,
				AccessList:           &emptyAL,
			}},
		}},
		"returnFullTransactions": true,
	})
	suite.Require().NoError(err)

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, suite.simulateV1Request(optsJSON))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error)

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	txs := results[0]["transactions"].([]interface{})
	suite.Require().Len(txs, 1)
	txObj := txs[0].(map[string]interface{})

	suite.Require().Equal("0x2", txObj["type"].(string),
		"dynamic-fee shape must surface as type 2")
	suite.Require().Equal("0x64", txObj["maxFeePerGas"].(string))
	suite.Require().Equal("0x2", txObj["maxPriorityFeePerGas"].(string))
	suite.Require().NotNil(txObj["accessList"], "type-2 tx must carry an accessList field")

	wantTx := ethtypes.NewTx(&ethtypes.DynamicFeeTx{
		ChainID:    mustHexBig(txObj["chainId"].(string)),
		Nonce:      mustHexUint64(txObj["nonce"].(string)),
		GasTipCap:  mustHexBig(txObj["maxPriorityFeePerGas"].(string)),
		GasFeeCap:  mustHexBig(txObj["maxFeePerGas"].(string)),
		Gas:        mustHexUint64(txObj["gas"].(string)),
		To:         &recipient,
		Value:      mustHexBig(txObj["value"].(string)),
		Data:       common.FromHex(txObj["input"].(string)),
		AccessList: emptyAL,
	})
	suite.Require().Equal(txObj["hash"].(string), wantTx.Hash().Hex(),
		"reconstructed dynamic-fee tx must match the published hash")
	suite.Require().Equal(recomputeTxRoot(wantTx).Hex(), results[0]["transactionsRoot"].(string))
}

// mustHexUint64 parses an `0x`-prefixed hex string into a uint64.
// Panics on malformed input — only used by envelope tests against
// well-formed JSON-RPC envelopes.
func mustHexUint64(s string) uint64 {
	v := new(big.Int)
	_, ok := v.SetString(s[2:], 16)
	if !ok {
		panic("mustHexUint64: " + s)
	}
	return v.Uint64()
}

// mustHexBig parses an `0x`-prefixed hex string into *big.Int.
func mustHexBig(s string) *big.Int {
	v := new(big.Int)
	_, ok := v.SetString(s[2:], 16)
	if !ok {
		panic("mustHexBig: " + s)
	}
	return v
}
