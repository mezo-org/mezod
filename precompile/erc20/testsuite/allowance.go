package testsuite

import (
	"math/big"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

func (s *TestSuite) TestAllowance() {
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
				return []interface{}{1}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid owner address",
			run: func() []interface{} {
				return []interface{}{"invalid address", s.account1.EvmAddr}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid spender address",
			run: func() []interface{} {
				return []interface{}{s.account1.EvmAddr, "invalid address"}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "no allowance authorization exist",
			run: func() []interface{} {
				return []interface{}{s.account1.EvmAddr, s.account2.EvmAddr}
			},
			basicPass:     true,
			expectedValue: common.Big0,
		},
		{
			name: "allowance exists for spender",
			run: func() []interface{} {
				s.setupSendAuthz(
					precompile.TypesConverter.Address.ToSDK(s.account1.EvmAddr),
					precompile.TypesConverter.Address.ToSDK(s.account2.EvmAddr),
					sdk.NewCoins(
						sdk.NewCoin(s.denom, sdkmath.NewInt(42)),
						sdk.NewCoin("otherdenom", sdkmath.NewInt(77)),
					).Sort(),
				)

				return []interface{}{s.account2.EvmAddr, s.account1.EvmAddr}
			},
			basicPass:     true,
			expectedValue: big.NewInt(42),
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

			method := s.erc20Precompile.Abi.Methods["allowance"]
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
			// In this case a function signature is 'function allowance(address owner, address spender)'
			vmContract.Input = append([]byte{0xdd, 0x62, 0xed, 0x3e}, methodInputArgs...)
			vmContract.CallerAddress = s.account2.EvmAddr

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
