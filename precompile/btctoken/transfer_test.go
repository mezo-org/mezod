package btctoken_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v12/utils"

	sdkmath "cosmossdk.io/math"
	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/precompile/btctoken"
	"github.com/evmos/evmos/v12/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

func (s *PrecompileTestSuite) TestTransfer() {
	testcases := []struct {
		name        string
		run         func() []interface{}
		postCheck   func()
		basicPass   bool
		errContains string
	}{
		{
			name:        "empty args",
			run:         func() []interface{} { return nil },
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid address",
			run: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2),
				}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid amount",
			run: func() []interface{} {
				return []interface{}{
					s.account1.EvmAddr, "invalid amount",
				}
			},
			errContains: "cannot use string as type ptr as argument",
		},
		{
			name: "not enough balance",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(utils.BaseDenom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), sdk.Coins{sdk.NewCoin(utils.BaseDenom, sdkmath.NewInt(42))})
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{
					s.account1.EvmAddr, big.NewInt(43),
				}
			},
			basicPass:   true,
			errContains: "insufficient funds",
		},
		{
			name: "successful transfer",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(utils.BaseDenom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), sdk.Coins{sdk.NewCoin(utils.BaseDenom, sdkmath.NewInt(42))})
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{
					s.account2.EvmAddr, big.NewInt(40),
				}
			},
			basicPass: true,
			postCheck: func() {
				// Check the balance of the two accounts
				acc1 := s.app.BankKeeper.GetBalance(s.ctx, s.account1.SdkAddr, utils.BaseDenom)
				acc2 := s.app.BankKeeper.GetBalance(s.ctx, s.account2.SdkAddr, utils.BaseDenom)
				s.Require().Equal(sdkmath.NewInt(2), acc1.Amount)
				s.Require().Equal(sdkmath.NewInt(40), acc2.Amount)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}

			bankKeeper := s.app.BankKeeper
			authzKeeper := s.app.AuthzKeeper

			btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper, authzKeeper)
			s.Require().NoError(err)
			s.btcTokenPrecompile = btcTokenPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.btcTokenPrecompile.Abi.Methods["transfer"]
			var methodInputArgs []byte
			methodInputArgs, err = method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			// These first 4 bytes correspond to the method ID (first 4 bytes of the
			// Keccak-256 hash of the function signature).
			// In this case a function signature is 'function transfer(address to, uint256 value)'
			vmContract.Input = append([]byte{0xa9, 0x05, 0x9c, 0xbb}, methodInputArgs...)
			vmContract.CallerAddress = s.account1.EvmAddr

			output, err := s.btcTokenPrecompile.Run(evm, vmContract, false)
			if err != nil {
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(true, out[0], "expected different value")

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

// TODO: add tests for `transferFrom`

// TODO: add tests for the `transfer` event
