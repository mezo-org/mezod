package validatorpool_test

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/mezo-org/mezod/precompile/validatorpool"
)

func (s *PrecompileTestSuite) TestTransferOwnership() {
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
			name: "valid call",
			run: func() []interface{} {
				return []interface{}{
					s.account2.EvmAddr,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				candidateOwner := s.keeper.GetCandidateOwner(s.ctx)
				s.Require().Equal(candidateOwner, s.account2.SdkAddr, "candidateOwner was not updated in keeper")
			},
		},
	}

	s.RunMethodTestCases(testcases, "transferOwnership")
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
	testcases := []TestCase{
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
				return nil
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not candidateOwner",
		},
		{
			name: "valid call",
			run: func() []interface{} {
				return nil
			},
			as:        s.account4.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
	}

	s.RunMethodTestCases(testcases, "acceptOwnership")
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
	testcases := []TestCase{
		{
			name:      "valid call",
			run:       func() []interface{} { return nil },
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{s.account1.EvmAddr},
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

	s.RunMethodTestCases(testcases, "owner")
}

func (s *PrecompileTestSuite) TestCandidateOwner() {
	testcases := []TestCase{
		{
			name:      "valid call",
			run:       func() []interface{} { return nil },
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{s.account4.EvmAddr},
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

	s.RunMethodTestCases(testcases, "candidateOwner")
}
