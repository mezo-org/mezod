package validatorpool_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/precompile/validatorpool"
	"github.com/evmos/evmos/v12/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestTransferOwnership() {
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
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}

			validatorpoolPrecompile, err := validatorpool.NewPrecompile(s.keeper)
			s.Require().NoError(err)
			s.validatorpoolPrecompile = validatorpoolPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.validatorpoolPrecompile.Abi.Methods["transferOwnership"]
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

func (s *PrecompileTestSuite) TestEmitOwnershipTransferStartedEvent() {
	testcases := []struct {
		name     string
		owner    common.Address
		newOwner common.Address
	}{
		{
			name:     "pass",
			owner:    s.account1.EvmAddr,
			newOwner: s.account2.EvmAddr,
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			e := validatorpool.NewOwnershipTransferStartedEvent(tc.owner, tc.newOwner)
			args := e.Arguments()

			s.Require().Len(args, 2)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.owner, args[0].Value)

			// Check the second argument
			s.Require().True(args[1].Indexed)
			s.Require().Equal(tc.newOwner, args[1].Value)
		})
	}
}

func (s *PrecompileTestSuite) TestAcceptOwnership() {
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

			validatorpoolPrecompile, err := validatorpool.NewPrecompile(s.keeper)
			s.Require().NoError(err)
			s.validatorpoolPrecompile = validatorpoolPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.validatorpoolPrecompile.Abi.Methods["acceptOwnership"]
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

func (s *PrecompileTestSuite) TestEmitOwnershipTransferredEvent() {
	testcases := []struct {
		name     string
		owner    common.Address
		newOwner common.Address
	}{
		{
			name:     "pass",
			owner:    s.account1.EvmAddr,
			newOwner: s.account2.EvmAddr,
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			e := validatorpool.NewOwnershipTransferredEvent(tc.owner, tc.newOwner)
			args := e.Arguments()

			s.Require().Len(args, 2)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.owner, args[0].Value)

			// Check the second argument
			s.Require().True(args[1].Indexed)
			s.Require().Equal(tc.newOwner, args[1].Value)
		})
	}
}

func (s *PrecompileTestSuite) TestOwner() {
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

			validatorpoolPrecompile, err := validatorpool.NewPrecompile(s.keeper)
			s.Require().NoError(err)
			s.validatorpoolPrecompile = validatorpoolPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.validatorpoolPrecompile.Abi.Methods["owner"]
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
			s.Require().Equal(s.account1.EvmAddr, out[0], "expected different value")

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCandidateOwner() {
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

			validatorpoolPrecompile, err := validatorpool.NewPrecompile(s.keeper)
			s.Require().NoError(err)
			s.validatorpoolPrecompile = validatorpoolPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.validatorpoolPrecompile.Abi.Methods["candidateOwner"]
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

			var emptyAddress common.Address

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(emptyAddress, out[0], "expected different value")

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
