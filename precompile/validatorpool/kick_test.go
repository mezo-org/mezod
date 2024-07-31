package validatorpool_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v12/precompile/validatorpool"
)

func (s *PrecompileTestSuite) TestKick() {
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
			name: "valid kick",
			run: func() []interface{} {
				return []interface{}{
					s.account3.EvmAddr,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// Check the keeper was updated
				_, found := s.keeper.GetValidator(s.ctx, types.ValAddress(s.account3.SdkAddr))
				s.Require().False(found, fmt.Sprintf("validator was not kicked %s\n", types.ValAddress(s.account3.SdkAddr)))
			},
		},
	}

	s.RunMethodTestCases(testcases, "kick")
}

func (s *PrecompileTestSuite) TestEmitValidatorKickedEvent() {
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
			e := validatorpool.NewValidatorKickedEvent(tc.operator)
			args := e.Arguments()

			s.Require().Len(args, 1)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.operator, args[0].Value)
		})
	}
}
