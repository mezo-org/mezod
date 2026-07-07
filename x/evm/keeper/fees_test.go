package keeper_test

import (
	"bytes"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (suite *KeeperTestSuite) TestCheckSenderBalance() {
	hundredInt := sdkmath.NewInt(100)
	zeroInt := sdkmath.ZeroInt()
	oneInt := sdkmath.OneInt()
	fiveInt := sdkmath.NewInt(5)
	fiftyInt := sdkmath.NewInt(50)
	negInt := sdkmath.NewInt(-10)

	testCases := []struct {
		name            string
		to              string
		gasLimit        uint64
		gasPrice        *sdkmath.Int
		gasFeeCap       *big.Int
		gasTipCap       *big.Int
		cost            *sdkmath.Int
		from            string
		accessList      *ethtypes.AccessList
		expectPass      bool
		enableFeemarket bool
	}{
		{
			name:       "Enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Equal balance",
			to:         suite.address.String(),
			gasLimit:   99,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "negative cost",
			to:         suite.address.String(),
			gasLimit:   1,
			gasPrice:   &oneInt,
			cost:       &negInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas limit, not enough balance",
			to:         suite.address.String(),
			gasLimit:   100,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas price, enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher gas price, not enough balance",
			to:         suite.address.String(),
			gasLimit:   20,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher cost, enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &fiftyInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher cost, not enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &hundredInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:            "Enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Equal balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        99,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "negative cost w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        1,
			gasFeeCap:       big.NewInt(1),
			cost:            &negInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas limit, not enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        100,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas price, enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas price, not enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        20,
			gasFeeCap:       big.NewInt(5),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher cost, enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &fiftyInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Higher cost, not enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &hundredInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
	}

	vmdb := suite.StateDB()
	value := uint256.NewInt(0)
	value.SetFromBig(hundredInt.BigInt())
	vmdb.AddBalance(suite.address, value, tracing.BalanceChangeUnspecified)
	balance := vmdb.GetBalance(suite.address)
	suite.Require().Equal(balance, value)
	err := vmdb.Commit()
	suite.Require().NoError(err, "Unexpected error while committing to vmdb: %d", err)

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			to := common.HexToAddress(tc.from)

			var amount, gasPrice, gasFeeCap, gasTipCap *big.Int
			if tc.cost != nil {
				amount = tc.cost.BigInt()
			}

			if tc.enableFeemarket {
				gasFeeCap = tc.gasFeeCap
				if tc.gasTipCap == nil {
					gasTipCap = oneInt.BigInt()
				} else {
					gasTipCap = tc.gasTipCap
				}
			} else if tc.gasPrice != nil {
				gasPrice = tc.gasPrice.BigInt()
			}

			ethTxParams := &evmtypes.EvmTxArgs{
				ChainID:   zeroInt.BigInt(),
				Nonce:     1,
				To:        &to,
				Amount:    amount,
				GasLimit:  tc.gasLimit,
				GasPrice:  gasPrice,
				GasFeeCap: gasFeeCap,
				GasTipCap: gasTipCap,
				Accesses:  tc.accessList,
			}
			tx := evmtypes.NewTx(ethTxParams)
			tx.From = tc.from

			txData, _ := evmtypes.UnpackTxData(tx.Data)

			acct := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, suite.address)
			err := keeper.CheckSenderBalance(
				sdkmath.NewIntFromBigInt(acct.Balance.ToBig()),
				txData,
			)

			if tc.expectPass {
				suite.Require().NoError(err, "valid test %d failed", i)
			} else {
				suite.Require().Error(err, "invalid test %d passed", i)
			}
		})
	}
}

// TestVerifyFeeAndDeductTxCostsFromUserBalance is a test method for both the VerifyFee
// function and the DeductTxCostsFromUserBalance method.
//
// NOTE: This method combines testing for both functions, because these used to be
// in one function and share a lot of the same setup.
// In practice, the two tested functions will also be sequentially executed.
func (suite *KeeperTestSuite) TestVerifyFeeAndDeductTxCostsFromUserBalance() {
	hundredInt := sdkmath.NewInt(100)
	zeroInt := sdkmath.ZeroInt()
	oneInt := sdkmath.NewInt(1)
	fiveInt := sdkmath.NewInt(5)
	fiftyInt := sdkmath.NewInt(50)

	// should be enough to cover all test cases
	initBalance := sdkmath.NewInt((ethparams.InitialBaseFee + 10) * 105)

	testCases := []struct {
		name             string
		gasLimit         uint64
		gasPrice         *sdkmath.Int
		gasFeeCap        *big.Int
		gasTipCap        *big.Int
		cost             *sdkmath.Int
		accessList       *ethtypes.AccessList
		expectPassVerify bool
		expectPassDeduct bool
		enableFeemarket  bool
		from             string
		malleate         func()
	}{
		{
			name:             "Enough balance",
			gasLimit:         10,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "Equal balance",
			gasLimit:         100,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "Higher gas limit, not enough balance",
			gasLimit:         105,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             suite.address.String(),
		},
		{
			name:             "Higher gas price, enough balance",
			gasLimit:         20,
			gasPrice:         &fiveInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "Higher gas price, not enough balance",
			gasLimit:         20,
			gasPrice:         &fiftyInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             suite.address.String(),
		},
		// This case is expected to be true because the fees can be deducted, but the tx
		// execution is going to fail because there is no more balance to pay the cost
		{
			name:             "Higher cost, enough balance",
			gasLimit:         100,
			gasPrice:         &oneInt,
			cost:             &fiftyInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		//  testcases with enableFeemarket enabled.
		{
			name:             "Invalid gasFeeCap w/ enableFeemarket",
			gasLimit:         10,
			gasFeeCap:        big.NewInt(1),
			gasTipCap:        big.NewInt(1),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: false,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:             "empty tip fee is valid to deduct",
			gasLimit:         10,
			gasFeeCap:        big.NewInt(ethparams.InitialBaseFee),
			gasTipCap:        big.NewInt(1),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:             "effectiveTip equal to gasTipCap",
			gasLimit:         100,
			gasFeeCap:        big.NewInt(ethparams.InitialBaseFee + 2),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:             "effectiveTip equal to (gasFeeCap - baseFee)",
			gasLimit:         105,
			gasFeeCap:        big.NewInt(ethparams.InitialBaseFee + 1),
			gasTipCap:        big.NewInt(2),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:             "Invalid from address",
			gasLimit:         10,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             "abcdef",
		},
		{
			name:     "Enough balance - with access list",
			gasLimit: 10,
			gasPrice: &oneInt,
			cost:     &oneInt,
			accessList: &ethtypes.AccessList{
				ethtypes.AccessTuple{
					Address:     suite.address,
					StorageKeys: []common.Hash{},
				},
			},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "gasLimit < intrinsicGas during IsCheckTx",
			gasLimit:         1,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: false,
			expectPassDeduct: true,
			from:             suite.address.String(),
			malleate: func() {
				suite.ctx = suite.ctx.WithIsCheckTx(true)
			},
		},
	}

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.SetupTest()
			vmdb := suite.StateDB()

			if tc.malleate != nil {
				tc.malleate()
			}
			var amount, gasPrice, gasFeeCap, gasTipCap *big.Int
			if tc.cost != nil {
				amount = tc.cost.BigInt()
			}

			if suite.enableFeemarket {
				if tc.gasFeeCap != nil {
					gasFeeCap = tc.gasFeeCap
				}
				if tc.gasTipCap == nil {
					gasTipCap = oneInt.BigInt()
				} else {
					gasTipCap = tc.gasTipCap
				}
				value := uint256.NewInt(0)
				value.SetFromBig(initBalance.BigInt())
				vmdb.AddBalance(suite.address, value, tracing.BalanceChangeUnspecified)
				balance := vmdb.GetBalance(suite.address)
				suite.Require().Equal(balance, value)
			} else {
				if tc.gasPrice != nil {
					gasPrice = tc.gasPrice.BigInt()
				}
				value := uint256.NewInt(0)
				value.SetFromBig(hundredInt.BigInt())
				vmdb.AddBalance(suite.address, value, tracing.BalanceChangeUnspecified)
				balance := vmdb.GetBalance(suite.address)
				suite.Require().Equal(balance, value)
			}
			err := vmdb.Commit()
			suite.Require().NoError(err, "Unexpected error while committing to vmdb: %d", err)

			ethTxParams := &evmtypes.EvmTxArgs{
				ChainID:   zeroInt.BigInt(),
				Nonce:     1,
				To:        &suite.address,
				Amount:    amount,
				GasLimit:  tc.gasLimit,
				GasPrice:  gasPrice,
				GasFeeCap: gasFeeCap,
				GasTipCap: gasTipCap,
				Accesses:  tc.accessList,
			}
			tx := evmtypes.NewTx(ethTxParams)
			tx.From = tc.from

			txData, _ := evmtypes.UnpackTxData(tx.Data)

			evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
			ethCfg := evmParams.GetChainConfig().EthereumConfig(nil)
			baseFee := suite.app.EvmKeeper.GetBaseFee(suite.ctx, ethCfg)
			priority := evmtypes.GetTxPriority(txData, baseFee)

			fees, err := keeper.VerifyFee(txData, evmtypes.DefaultEVMDenom, baseFee, false, false, false, false, suite.ctx.IsCheckTx())
			if tc.expectPassVerify {
				suite.Require().NoError(err, "valid test %d failed - '%s'", i, tc.name)
				if tc.enableFeemarket {
					baseFee := suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx)
					suite.Require().Equal(
						fees,
						sdk.NewCoins(
							sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromBigInt(txData.EffectiveFee(baseFee))),
						),
						"valid test %d failed, fee value is wrong  - '%s'", i, tc.name,
					)
					suite.Require().Equal(int64(0), priority)
				} else {
					suite.Require().Equal(
						fees,
						sdk.NewCoins(
							sdk.NewCoin(evmtypes.DefaultEVMDenom, tc.gasPrice.Mul(sdkmath.NewIntFromUint64(tc.gasLimit))),
						),
						"valid test %d failed, fee value is wrong  - '%s'", i, tc.name,
					)
				}
			} else {
				suite.Require().Error(err, "invalid test %d passed - '%s'", i, tc.name)
				suite.Require().Nil(fees, "invalid test %d passed. fees value must be nil - '%s'", i, tc.name)
			}

			err = suite.app.EvmKeeper.DeductTxCostsFromUserBalance(suite.ctx, fees, common.HexToAddress(tx.From))
			if tc.expectPassDeduct {
				suite.Require().NoError(err, "valid test %d failed - '%s'", i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test %d passed - '%s'", i, tc.name)
			}
		})
	}
	suite.enableFeemarket = false // reset flag
}

// TestVerifyFeeSetCodeAuthList confirms the VerifyFee path threads
// txData.GetAuthorizationList() into core.IntrinsicGas — three tuples
// add 3 * params.CallNewAccountGas to intrinsic gas. We pin the boundary:
// gas one tuple short must reject, exact threshold must accept, and any
// surplus must also accept.
func (suite *KeeperTestSuite) TestVerifyFeeSetCodeAuthList() {
	suite.SetupTest()
	suite.ctx = suite.ctx.WithIsCheckTx(true)

	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0x0000000000000000000000000000000000005678")

	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	auth, err := ethtypes.SignSetCode(priv, ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(chainID),
		Address: target,
		Nonce:   0,
	})
	suite.Require().NoError(err)

	authList := []ethtypes.SetCodeAuthorization{auth, auth, auth}

	testCases := []struct {
		name    string
		gas     uint64
		wantErr bool
	}{
		{
			// One CallNewAccountGas short of the 3-tuple threshold —
			// CheckTx must reject as "intrinsic gas too low".
			name:    "below threshold",
			gas:     ethparams.TxGas + 2*ethparams.CallNewAccountGas,
			wantErr: true,
		},
		{
			// Exact intrinsic gas: TxGas + 3 * CallNewAccountGas. Must
			// accept; a regression that miscounted tuples would fail here.
			name:    "exact threshold",
			gas:     ethparams.TxGas + 3*ethparams.CallNewAccountGas,
			wantErr: false,
		},
		{
			// Surplus over threshold must also accept.
			name:    "above threshold",
			gas:     ethparams.TxGas + 4*ethparams.CallNewAccountGas,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tx := ethtypes.NewTx(&ethtypes.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     1,
				GasTipCap: uint256.NewInt(1),
				GasFeeCap: uint256.NewInt(1),
				Gas:       tc.gas,
				To:        target,
				Value:     uint256.NewInt(0),
				Data:      nil,
				AuthList:  authList,
			})
			txData, err := evmtypes.NewSetCodeTx(tx)
			suite.Require().NoError(err)

			// Non-nil baseFee so EffectiveFee — reached only on the
			// accept paths after the intrinsic-gas check passes — does
			// not nil-deref inside EffectiveGasPrice.
			_, err = keeper.VerifyFee(txData, evmtypes.DefaultEVMDenom, big.NewInt(0), false, false, false, false, suite.ctx.IsCheckTx())
			if tc.wantErr {
				suite.Require().Error(err, "must reject when gas limit < intrinsic gas")
			} else {
				suite.Require().NoError(err, "must accept when gas limit >= intrinsic gas")
			}
		})
	}
}

// TestVerifyFeeFloorDataGas pins the EIP-7623 floor admission gate.
// Heavy zero-byte calldata sits in the gap where intrinsic gas alone
// is well below the floor, so the floor check is the only rejection
// path for the under-floor cases.
func (suite *KeeperTestSuite) TestVerifyFeeFloorDataGas() {
	const dataLen = 4096
	data := bytes.Repeat([]byte{0}, dataLen)
	intrinsic := ethparams.TxGas + dataLen*ethparams.TxDataZeroGas
	floor := ethparams.TxGas + dataLen*ethparams.TxCostFloorPerToken

	testCases := []struct {
		name      string
		gas       uint64
		isPrague  bool
		isCheckTx bool
		wantErr   bool
	}{
		{
			name:      "pre-Prague: gasLimit = intrinsic accepted",
			gas:       intrinsic,
			isPrague:  false,
			isCheckTx: true,
			wantErr:   false,
		},
		{
			name:      "Prague CheckTx: gasLimit = intrinsic (< floor) rejected",
			gas:       intrinsic,
			isPrague:  true,
			isCheckTx: true,
			wantErr:   true,
		},
		{
			name:      "Prague CheckTx: gasLimit = floor accepted",
			gas:       floor,
			isPrague:  true,
			isCheckTx: true,
			wantErr:   false,
		},
		{
			name:      "Prague CheckTx: gasLimit = floor-1 rejected",
			gas:       floor - 1,
			isPrague:  true,
			isCheckTx: true,
			wantErr:   true,
		},
		{
			name:      "Prague DeliverTx: gasLimit < floor accepted (deferred to apply-message)",
			gas:       intrinsic,
			isPrague:  true,
			isCheckTx: false,
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tx := ethtypes.NewTx(&ethtypes.DynamicFeeTx{
				Nonce:     1,
				GasTipCap: big.NewInt(1),
				GasFeeCap: big.NewInt(1),
				Gas:       tc.gas,
				To:        &common.Address{},
				Value:     big.NewInt(0),
				Data:      data,
			})
			txData, err := evmtypes.NewDynamicFeeTx(tx)
			suite.Require().NoError(err)

			// Non-zero baseFee would fail the later gasFeeCap < baseFee
			// check on the accept paths; mirror TestVerifyFeeSetCodeAuthList.
			_, err = keeper.VerifyFee(
				txData,
				evmtypes.DefaultEVMDenom,
				big.NewInt(0),
				false, false, false,
				tc.isPrague,
				tc.isCheckTx,
			)
			if tc.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestSetCodeTx_AuthListRoundTripCount pins the invariant that VerifyFee's
// txData.GetAuthorizationList() (Cosmos-encoded path) and the apply loop's
// core.Message.SetCodeAuthorizations (geth path) report the same number of
// tuples for the same SetCodeTx. A future filter on either route that
// silently dropped tuples would diverge intrinsic-gas accounting from
// authorization processing — N=0 catches regressions that always emit one
// tuple, N=3 catches off-by-one truncation.
func (suite *KeeperTestSuite) TestSetCodeTx_AuthListRoundTripCount() {
	suite.SetupTest()

	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0x000000000000000000000000000000000000abcd")

	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	auth, err := ethtypes.SignSetCode(priv, ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(chainID),
		Address: target,
		Nonce:   0,
	})
	suite.Require().NoError(err)

	testCases := []struct {
		name string
		n    int
	}{
		{"empty auth list", 0},
		{"three tuples", 3},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			authList := make([]ethtypes.SetCodeAuthorization, tc.n)
			for i := range authList {
				authList[i] = auth
			}

			signer := ethtypes.LatestSignerForChainID(chainID)
			tx, err := ethtypes.SignNewTx(priv, signer, &ethtypes.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     1,
				GasTipCap: uint256.NewInt(1),
				GasFeeCap: uint256.NewInt(1),
				Gas:       ethparams.TxGas + uint64(tc.n)*ethparams.CallNewAccountGas, //nolint:gosec
				To:        target,
				Value:     uint256.NewInt(0),
				AuthList:  authList,
			})
			suite.Require().NoError(err)

			txData, err := evmtypes.NewSetCodeTx(tx)
			suite.Require().NoError(err)

			coreMsg, err := core.TransactionToMessage(tx, signer, nil)
			suite.Require().NoError(err)

			suite.Require().Equal(
				len(txData.GetAuthorizationList()),
				len(coreMsg.SetCodeAuthorizations),
				"round-trip auth count must agree",
			)
			suite.Require().Equal(tc.n, len(txData.GetAuthorizationList()))
		})
	}
}
