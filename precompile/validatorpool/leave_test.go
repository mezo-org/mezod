package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestLeave() {
	testcases := []struct {
		name        string
		run         func() []interface{}
		postCheck   func()
		basicPass   bool
		errContains string
	}{
		{
			name:      "empty args",
			run:       func() []interface{} { return nil },
			basicPass: true,
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
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}

			validatorpoolPrecompile, err := NewPrecompile(s.keeper)
			s.Require().NoError(err)
			s.validatorpoolPrecompile = validatorpoolPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.validatorpoolPrecompile.Abi.Methods["leave"]
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
			vmContract.Input = append(vmContract.Input, method.ID...)
			vmContract.Input = append(vmContract.Input, methodInputArgs...)
			vmContract.CallerAddress = s.account2.EvmAddr

			output, err := s.validatorpoolPrecompile.Run(evm, vmContract, false)
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

func (s *PrecompileTestSuite) TestEmitValidatorLeftEvent() {
	testcases := []struct {
		name     string
		operator common.Address
	}{
		{
			name:     "pass",
			operator: s.account1.EvmAddr,
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			e := newValidatorLeftEvent(tc.operator)
			args := e.Arguments()

			s.Require().Len(args, 1)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.operator, args[0].Value)
		})
	}
}
