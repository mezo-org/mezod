package testsuite

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/precompile"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (s *TestSuite) TestBalance() {
	testcases := []struct {
		name          string
		run           func() []interface{}
		errContains   string
		expectedValue *big.Int
		basicPass     bool
	}{
		{
			name: "invalid number of arguments",
			run: func() []interface{} {
				return []interface{}{1, 1}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "check zero balance",
			run: func() []interface{} {
				return []interface{}{s.account1.EvmAddr}
			},
			basicPass:     true,
			expectedValue: big.NewInt(0),
		},
		{
			name: "check non-zero balance",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the address
				err := s.app.BankKeeper.MintCoins(
					s.ctx,
					evmtypes.ModuleName,
					sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1e18))},
				)
				s.Require().NoError(err, "failed to mint coins")

				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(
					s.ctx,
					evmtypes.ModuleName,
					s.account1.EvmAddr.Bytes(),
					sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1000000000000))},
				)
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{s.account1.EvmAddr}
			},
			basicPass:     true,
			expectedValue: big.NewInt(1000000000000),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
			}

			erc20Precompile, err := s.precompileFactoryFn(s.app)
			s.Require().NoError(err)
			s.erc20Precompile = erc20Precompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.erc20Precompile.Abi.Methods["balanceOf"]
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
			// In this case a function signature is 'function balanceOf(address)'
			vmContract.Input = append([]byte{0x70, 0xa0, 0x82, 0x31}, methodInputArgs...)

			output, err := s.erc20Precompile.Run(evm, vmContract, false)
			if err != nil && tc.errContains != "" {
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)

			val, ok := out[0].(*big.Int)
			if !ok {
				s.Require().Equal(tc.expectedValue, out[0], "expected different value")
			} else {
				s.Require().Equal(0, tc.expectedValue.Cmp(val), "expected different value")
			}
		})
	}
}
