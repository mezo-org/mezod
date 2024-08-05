package validatorpool_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/mezo-org/mezod/precompile/validatorpool"
	poatypes "github.com/mezo-org/mezod/x/poa/types"
)

func (s *PrecompileTestSuite) TestSubmitApplication() {
	testcases := []TestCase{
		{
			name:        "empty args",
			run:         func() []interface{} { return nil },
			as:          s.account1.EvmAddr,
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			as:          s.account1.EvmAddr,
			errContains: "argument count mismatch",
		},
		{
			name: "keeper returns error",
			run: func() []interface{} {
				return []interface{}{
					s.account2.ConsPubKeyBytes32(),
					s.account2.Description,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: poatypes.ErrAlreadyApplying.Error(),
		},
		{
			name: "valid application",
			run: func() []interface{} {
				return []interface{}{
					s.account1.ConsPubKeyBytes32(),
					s.account1.Description,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// Check the keeper was updated
				application, found := s.keeper.GetApplication(s.ctx, types.ValAddress(s.account1.SdkAddr))
				s.Require().True(found, fmt.Sprintf("application not found %s\n%v", types.ValAddress(s.account1.SdkAddr), s.keeper.GetAllApplications(s.ctx)))
				operator := types.AccAddress(application.Validator.GetOperator())
				s.Require().Equal(operator, s.account1.SdkAddr, "expected application operator to match")
				description := validatorpool.Description(application.Validator.GetDescription())
				s.Require().Equal(description.Moniker, s.account1.Description.Moniker, "expected moniker to match")
			},
		},
	}

	s.RunMethodTestCases(testcases, "submitApplication")
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
			consPubKey: s.account1.ConsPubKeyBytes32(),
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			e := validatorpool.NewApplicationSubmittedEvent(tc.operator, tc.consPubKey, s.account1.Description)
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
			s.Require().Equal(s.account1.Description, args[2].Value)
		})
	}
}

func (s *PrecompileTestSuite) TestApproveApplication() {
	testcases := []TestCase{
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
			name: "keeper returns error",
			run: func() []interface{} {
				return []interface{}{
					s.account2.EvmAddr,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "valid approval",
			run: func() []interface{} {
				return []interface{}{
					s.account2.EvmAddr,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// Check the keeper was updated
				validator, found := s.keeper.GetValidator(s.ctx, types.ValAddress(s.account2.SdkAddr))
				s.Require().True(found, fmt.Sprintf("validator not found %s\n", types.ValAddress(s.account2.SdkAddr)))
				operator := types.AccAddress(validator.GetOperator())
				s.Require().Equal(operator, s.account2.SdkAddr, "expected validator operator to match")
				description := validatorpool.Description(validator.GetDescription())
				s.Require().Equal(description.Moniker, s.account2.Description.Moniker, "expected moniker to match")
			},
		},
	}

	s.RunMethodTestCases(testcases, "approveApplication")
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
			e := validatorpool.NewApplicationApprovedEvent(tc.operator)
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
			e := validatorpool.NewValidatorJoinedEvent(tc.operator)
			args := e.Arguments()

			s.Require().Len(args, 1)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.operator, args[0].Value)
		})
	}
}

func (s *PrecompileTestSuite) TestApplication() {
	testcases := []TestCase{
		{
			name:        "empty args",
			run:         func() []interface{} { return nil },
			as:          s.account1.EvmAddr,
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			as:          s.account1.EvmAddr,
			errContains: "argument count mismatch",
		},
		{
			name: "keeper returns error",
			run: func() []interface{} {
				return []interface{}{
					s.account3.EvmAddr,
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "application does not exist",
		},
		{
			name: "valid application",
			run: func() []interface{} {
				return []interface{}{
					s.account2.EvmAddr,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{s.account2.ConsPubKeyBytes32(), s.account2.Description},
		},
	}

	s.RunMethodTestCases(testcases, "application")
}

func (s *PrecompileTestSuite) TestApplications() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			as:          s.account1.EvmAddr,
			errContains: "argument count mismatch",
		},
		{
			name:      "valid call",
			run:       func() []interface{} { return nil },
			as:        s.account1.EvmAddr,
			basicPass: true,
			output: []interface{}{
				[]common.Address{s.account2.EvmAddr},
			},
		},
	}

	s.RunMethodTestCases(testcases, "applications")
}
