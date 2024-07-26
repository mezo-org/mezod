package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/x/evm/statedb"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"
)

func (s *PrecompileTestSuite) TestSubmitApplication() {
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
					1, 2,
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

			method := s.validatorpoolPrecompile.Abi.Methods["submitApplication"]
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

func (s *PrecompileTestSuite) TestEmitApplicationSubmittedEvent() {
	testcases := []struct {
		name       string
		operator   common.Address
		consPubKey [32]byte
	}{
		{
			name:       "pass",
			operator:   s.account1.EvmAddr,
			consPubKey: [32]byte(s.account1.ConsPubKey.Bytes()),
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			description := poatypes.NewDescription("moniker", "identity", "website", "securityContact", "details")
			e := newApplicationSubmittedEvent(tc.operator, tc.consPubKey, description)
			args := e.Arguments()

			s.Require().Len(args, 3)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.operator, args[0].Value)

			// Check the second argument
			s.Require().True(args[1].Indexed)
			s.Require().Equal(tc.consPubKey, args[1].Value)

			// Check the second argument
			s.Require().False(args[2].Indexed)
			s.Require().Equal(description, args[2].Value)
		})
	}
}

func (s *PrecompileTestSuite) TestApproveApplication() {
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

			validatorpoolPrecompile, err := NewPrecompile(s.keeper)
			s.Require().NoError(err)
			s.validatorpoolPrecompile = validatorpoolPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.validatorpoolPrecompile.Abi.Methods["approveApplication"]
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

func (s *PrecompileTestSuite) TestEmitApplicationApprovedEvent() {
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
			e := newApplicationApprovedEvent(tc.operator)
			args := e.Arguments()

			s.Require().Len(args, 1)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.operator, args[0].Value)
		})
	}
}

func (s *PrecompileTestSuite) TestEmitValidatorJoinedEvent() {
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
			e := newValidatorJoinedEvent(tc.operator)
			args := e.Arguments()

			s.Require().Len(args, 1)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.operator, args[0].Value)
		})
	}
}
