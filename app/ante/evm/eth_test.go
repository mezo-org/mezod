package evm_test

import (
	"math"
	"math/big"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethante "github.com/mezo-org/mezod/app/ante/evm"
	"github.com/mezo-org/mezod/server/config"
	"github.com/mezo-org/mezod/testutil"
	testutiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/mezo-org/mezod/types"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

func (suite *AnteTestSuite) TestNewEthAccountVerificationDecorator() {
	dec := ethante.NewEthAccountVerificationDecorator(
		suite.app.AccountKeeper, suite.app.EvmKeeper,
	)

	addr := testutiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx.From = addr.Hex()

	var vmdb *statedb.StateDB

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		checkTx  bool
		expPass  bool
	}{
		{"not CheckTx", nil, func() {}, false, true},
		{"invalid transaction type", &testutiltx.InvalidTx{}, func() {}, true, false},
		{
			"sender not set to msg",
			tx,
			func() {},
			true,
			false,
		},
		{
			"sender not EOA",
			tx,
			func() {
				// set not as an EOA
				vmdb.SetCode(addr, []byte("1"))
			},
			true,
			false,
		},
		{
			"not enough balance to cover tx cost",
			tx,
			func() {
				// reset back to EOA
				vmdb.SetCode(addr, nil)
			},
			true,
			false,
		},
		{
			"success new account",
			tx,
			func() {
				vmdb.AddBalance(addr, uint256.NewInt(1000000), tracing.BalanceChangeUnspecified)
			},
			true,
			true,
		},
		{
			"success existing account",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				vmdb.AddBalance(addr, uint256.NewInt(1000000), tracing.BalanceChangeUnspecified)
			},
			true,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vmdb = testutil.NewStateDB(suite.ctx, suite.app.EvmKeeper)
			tc.malleate()
			suite.Require().NoError(vmdb.Commit())

			_, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(tc.checkTx), tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestEthNonceVerificationDecorator() {
	suite.SetupTest()
	dec := ethante.NewEthIncrementSenderSequenceDecorator(suite.app.AccountKeeper)

	addr := testutiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx.From = addr.Hex()

	testCases := []struct {
		name      string
		tx        sdk.Tx
		malleate  func()
		reCheckTx bool
		expPass   bool
	}{
		{"ReCheckTx", &testutiltx.InvalidTx{}, func() {}, true, false},
		{"invalid transaction type", &testutiltx.InvalidTx{}, func() {}, false, false},
		{"sender account not found", tx, func() {}, false, false},
		{
			"sender nonce missmatch",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
			false,
		},
		{
			"success",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.Require().NoError(acc.SetSequence(1))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()
			_, err := dec.AnteHandle(suite.ctx.WithIsReCheckTx(tc.reCheckTx), tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestEthGasConsumeDecorator() {
	chainID := suite.app.EvmKeeper.ChainID()
	dec := ethante.NewEthGasConsumeDecorator(suite.app.BankKeeper, suite.app.EvmKeeper, config.DefaultMaxTxGasWanted)

	addr := testutiltx.GenerateAddress()

	txGasLimit := uint64(1000)

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: txGasLimit,
		GasPrice: big.NewInt(1),
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx.From = addr.Hex()

	ethCfg := suite.app.EvmKeeper.GetParams(suite.ctx).
		ChainConfig.EthereumConfig(chainID)
	baseFee := suite.app.EvmKeeper.GetBaseFee(suite.ctx, ethCfg)
	suite.Require().Equal(int64(1000000000), baseFee.Int64())

	gasPrice := new(big.Int).Add(baseFee, evmtypes.DefaultPriorityReduction.BigInt())

	tx2GasLimit := uint64(1000000)
	eth2TxContractParams := &evmtypes.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: tx2GasLimit,
		GasPrice: gasPrice,
		Accesses: &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
	}
	tx2 := evmtypes.NewTx(eth2TxContractParams)
	tx2.From = addr.Hex()
	tx2Priority := int64(1)

	tx3GasLimit := types.BlockGasLimit(suite.ctx) + uint64(1)
	eth3TxContractParams := &evmtypes.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: tx3GasLimit,
		GasPrice: gasPrice,
		Accesses: &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
	}
	tx3 := evmtypes.NewTx(eth3TxContractParams)

	dynamicTxContractParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  tx2GasLimit,
		GasFeeCap: new(big.Int).Add(baseFee, big.NewInt(evmtypes.DefaultPriorityReduction.Int64()*2)),
		GasTipCap: evmtypes.DefaultPriorityReduction.BigInt(),
		Accesses:  &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
	}
	dynamicFeeTx := evmtypes.NewTx(dynamicTxContractParams)
	dynamicFeeTx.From = addr.Hex()
	dynamicFeeTxPriority := int64(1)

	var vmdb *statedb.StateDB

	testCases := []struct {
		name        string
		tx          sdk.Tx
		gasLimit    uint64
		malleate    func(ctx sdk.Context) sdk.Context
		expPass     bool
		expPanic    bool
		expPriority int64
		postCheck   func(ctx sdk.Context)
	}{
		{
			"invalid transaction type",
			&testutiltx.InvalidTx{},
			math.MaxUint64,
			func(ctx sdk.Context) sdk.Context { return ctx },
			false,
			false,
			0,
			func(_ sdk.Context) {},
		},
		{
			"sender not found",
			evmtypes.NewTx(&evmtypes.EvmTxArgs{
				ChainID:  chainID,
				Nonce:    1,
				Amount:   big.NewInt(10),
				GasLimit: 1000,
				GasPrice: big.NewInt(1),
			}),
			math.MaxUint64,
			func(ctx sdk.Context) sdk.Context { return ctx },
			false, false,
			0,
			func(_ sdk.Context) {},
		},
		{
			"gas limit too low",
			tx,
			math.MaxUint64,
			func(ctx sdk.Context) sdk.Context { return ctx },
			false, false,
			0,
			func(_ sdk.Context) {},
		},
		{
			"gas limit above block gas limit",
			tx3,
			math.MaxUint64,
			func(ctx sdk.Context) sdk.Context { return ctx },
			false, false,
			0,
			func(_ sdk.Context) {},
		},
		{
			"not enough balance for fees",
			tx2,
			math.MaxUint64,
			func(ctx sdk.Context) sdk.Context { return ctx },
			false, false,
			0,
			func(_ sdk.Context) {},
		},
		{
			"not enough tx gas",
			tx2,
			0,
			func(ctx sdk.Context) sdk.Context {
				vmdb.AddBalance(addr, uint256.NewInt(1e6), tracing.BalanceChangeUnspecified)
				return ctx
			},
			false, true,
			0,
			func(_ sdk.Context) {},
		},
		{
			"not enough block gas",
			tx2,
			0,
			func(ctx sdk.Context) sdk.Context {
				vmdb.AddBalance(addr, uint256.NewInt(1e6), tracing.BalanceChangeUnspecified)
				return ctx.WithBlockGasMeter(storetypes.NewGasMeter(1))
			},
			false, true,
			0,
			func(_ sdk.Context) {},
		},
		{
			"success - legacy tx",
			tx2,
			tx2GasLimit, // it's capped
			func(ctx sdk.Context) sdk.Context {
				vmdb.AddBalance(addr, uint256.NewInt(1e16), tracing.BalanceChangeUnspecified)
				return ctx.WithBlockGasMeter(storetypes.NewGasMeter(1e19))
			},
			true, false,
			tx2Priority,
			func(_ sdk.Context) {},
		},
		{
			"success - dynamic fee tx",
			dynamicFeeTx,
			tx2GasLimit, // it's capped
			func(ctx sdk.Context) sdk.Context {
				vmdb.AddBalance(addr, uint256.NewInt(1e16), tracing.BalanceChangeUnspecified)
				return ctx.WithBlockGasMeter(storetypes.NewGasMeter(1e19))
			},
			true, false,
			dynamicFeeTxPriority,
			func(_ sdk.Context) {},
		},
		{
			"success - gas limit on gasMeter is set on ReCheckTx mode",
			dynamicFeeTx,
			0, // for reCheckTX mode, gas limit should be set to 0
			func(ctx sdk.Context) sdk.Context {
				vmdb.AddBalance(addr, uint256.NewInt(1e15), tracing.BalanceChangeUnspecified)
				return ctx.WithIsReCheckTx(true)
			},
			true, false,
			0,
			func(_ sdk.Context) {},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			cacheCtx, _ := suite.ctx.CacheContext()
			// Create new stateDB for each test case from the cached context
			vmdb = testutil.NewStateDB(cacheCtx, suite.app.EvmKeeper)
			cacheCtx = tc.malleate(cacheCtx)
			suite.Require().NoError(vmdb.Commit())

			if tc.expPanic {
				suite.Require().Panics(func() {
					_, _ = dec.AnteHandle(cacheCtx.WithIsCheckTx(true).WithGasMeter(storetypes.NewGasMeter(1)), tc.tx, false, testutil.NextFn)
				})
				return
			}

			ctx, err := dec.AnteHandle(cacheCtx.WithIsCheckTx(true).WithGasMeter(storetypes.NewInfiniteGasMeter()), tc.tx, false, testutil.NextFn)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expPriority, ctx.Priority())
			} else {
				suite.Require().Error(err)
			}
			suite.Require().Equal(tc.gasLimit, ctx.GasMeter().Limit())

			// check state after the test case
			tc.postCheck(ctx)
		})
	}
}

func (suite *AnteTestSuite) TestCanTransferDecorator() {
	dec := ethante.NewCanTransferDecorator(suite.app.EvmKeeper)

	addr, privKey := testutiltx.NewAddrKey()

	suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))
	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  1000,
		GasPrice:  big.NewInt(1),
		GasFeeCap: big.NewInt(150),
		GasTipCap: big.NewInt(200),
		Accesses:  &ethtypes.AccessList{},
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx2 := evmtypes.NewTx(ethContractCreationTxParams)

	tx.From = addr.Hex()

	err := tx.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	var vmdb *statedb.StateDB

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
	}{
		{"invalid transaction type", &testutiltx.InvalidTx{}, func() {}, false},
		{"AsMessage failed", tx2, func() {}, false},
		{
			"evm CanTransfer failed",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
		},
		{
			"success",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				vmdb.AddBalance(addr, uint256.NewInt(1000000), tracing.BalanceChangeUnspecified)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vmdb = testutil.NewStateDB(suite.ctx, suite.app.EvmKeeper)
			tc.malleate()
			suite.Require().NoError(vmdb.Commit())

			_, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(true), tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestEthIncrementSenderSequenceDecorator() {
	dec := ethante.NewEthIncrementSenderSequenceDecorator(suite.app.AccountKeeper)
	addr, privKey := testutiltx.NewAddrKey()

	ethTxContractParamsNonce0 := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    0,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	contract := evmtypes.NewTx(ethTxContractParamsNonce0)
	contract.From = addr.Hex()
	err := contract.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	to := testutiltx.GenerateAddress()
	ethTxParamsNonce0 := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    0,
		To:       &to,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	tx := evmtypes.NewTx(ethTxParamsNonce0)
	tx.From = addr.Hex()
	err = tx.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	ethTxParamsNonce1 := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		To:       &to,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	tx2 := evmtypes.NewTx(ethTxParamsNonce1)
	tx2.From = addr.Hex()
	err = tx2.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
		expPanic bool
	}{
		{
			"invalid transaction type",
			&testutiltx.InvalidTx{},
			func() {},
			false, false,
		},
		{
			"no signers",
			evmtypes.NewTx(ethTxParamsNonce1),
			func() {},
			false, false,
		},
		{
			"account not set to store",
			tx,
			func() {},
			false, false,
		},
		{
			"success - create contract",
			contract,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			true, false,
		},
		{
			"success - call",
			tx2,
			func() {},
			true, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			if tc.expPanic {
				suite.Require().Panics(func() {
					_, _ = dec.AnteHandle(suite.ctx, tc.tx, false, testutil.NextFn)
				})
				return
			}

			_, err := dec.AnteHandle(suite.ctx, tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
				msg := tc.tx.(*evmtypes.MsgEthereumTx)

				txData, err := evmtypes.UnpackTxData(msg.Data)
				suite.Require().NoError(err)

				nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, addr)
				suite.Require().Equal(txData.GetNonce()+1, nonce)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
