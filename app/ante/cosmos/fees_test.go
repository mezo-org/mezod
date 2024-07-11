package cosmos_test

import (
	"cosmossdk.io/math"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosante "github.com/evmos/evmos/v12/app/ante/cosmos"
	"github.com/evmos/evmos/v12/testutil"
	testutiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/evmos/evmos/v12/utils"
)

func (suite *AnteTestSuite) TestDeductFeeDecorator() {
	var (
		dfd cosmosante.DeductFeeDecorator
		// General setup
		addr, priv  = testutiltx.NewAccAddressAndKey()
		initBalance = sdk.NewInt(1e18)
		lowGasPrice = math.NewInt(1)
		zero        = sdk.ZeroInt()
	)

	// Testcase definitions
	testcases := []struct {
		name        string
		balance     math.Int
		rewards     math.Int
		gas         uint64
		gasPrice    *math.Int
		feeGranter  sdk.AccAddress
		checkTx     bool
		simulate    bool
		expPass     bool
		errContains string
		postCheck   func()
		malleate    func()
	}{
		{
			name:        "pass - sufficient balance to pay fees",
			balance:     initBalance,
			rewards:     zero,
			gas:         0,
			checkTx:     false,
			simulate:    true,
			expPass:     true,
			errContains: "",
		},
		{
			name:        "fail - zero gas limit in check tx mode",
			balance:     initBalance,
			rewards:     zero,
			gas:         0,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "must provide positive gas",
		},
		{
			name:        "fail - insufficient funds",
			balance:     sdk.NewInt(1e5),
			rewards:     sdk.NewInt(1e5),
			gas:         10_000_000,
			checkTx:     false,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient funds",
			postCheck: func() {
				// the balance should not have changed
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(sdk.NewInt(1e5), balance.Amount, "expected balance to be unchanged")
			},
		},
		{
			name:        "fail - sufficient balance to pay fees but provided fees < required fees",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			gasPrice:    &lowGasPrice,
			checkTx:     true,
			simulate:    false,
			expPass:     false,
			errContains: "insufficient fees",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(10_000)),
					),
				)
			},
		},
		{
			name:        "success - sufficient balance to pay fees & min gas prices is zero",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			gasPrice:    &lowGasPrice,
			checkTx:     true,
			simulate:    false,
			expPass:     true,
			errContains: "",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, zero),
					),
				)
			},
		},
		{
			name:        "success - sufficient balance to pay fees (fees > required fees)",
			balance:     initBalance,
			rewards:     zero,
			gas:         10_000_000,
			checkTx:     true,
			simulate:    false,
			expPass:     true,
			errContains: "",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(100)),
					),
				)
			},
		},
		{
			name:        "success - zero fees",
			balance:     initBalance,
			rewards:     zero,
			gas:         100,
			gasPrice:    &zero,
			checkTx:     true,
			simulate:    false,
			expPass:     true,
			errContains: "",
			malleate: func() {
				suite.ctx = suite.ctx.WithMinGasPrices(
					sdk.NewDecCoins(
						sdk.NewDecCoin(utils.BaseDenom, zero),
					),
				)
			},
			postCheck: func() {
				// the tx sender balance should not have changed
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(initBalance, balance.Amount, "expected balance to be unchanged")
			},
		},
	}

	// Test execution
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Create a new DeductFeeDecorator
			dfd = cosmosante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil, nil)

			// prepare the testcase
			var err error
			err = testutil.PrepareAccount(suite.ctx, suite.app.AccountKeeper, suite.app.BankKeeper, addr, tc.balance)
			suite.Require().NoError(err, "failed to prepare account")

			// Create an arbitrary message for testing purposes
			msg := sdktestutil.NewTestMsg(addr)

			// Set up the transaction arguments
			args := testutiltx.CosmosTxArgs{
				TxCfg:      suite.clientCtx.TxConfig,
				Priv:       priv,
				Gas:        tc.gas,
				GasPrice:   tc.gasPrice,
				FeeGranter: tc.feeGranter,
				Msgs:       []sdk.Msg{msg},
			}

			if tc.malleate != nil {
				tc.malleate()
			}
			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx)

			// Create a transaction out of the message
			tx, err := testutiltx.PrepareCosmosTx(suite.ctx, suite.app, args)
			suite.Require().NoError(err, "failed to create transaction")

			// run the ante handler
			_, err = dfd.AnteHandle(suite.ctx, tx, tc.simulate, testutil.NextFn)

			// assert the resulting error
			if tc.expPass {
				suite.Require().NoError(err, "expected no error")
			} else {
				suite.Require().Error(err, "expected error")
				suite.Require().ErrorContains(err, tc.errContains)
			}

			// run the post check
			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
