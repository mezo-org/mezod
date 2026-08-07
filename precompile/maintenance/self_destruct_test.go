package maintenance_test

import evmtypes "github.com/mezo-org/mezod/x/evm/types"

func (s *PrecompileTestSuite) TestSetSelfDestructDisabled() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid disabled argument type",
			run: func() []interface{} {
				return []interface{}{
					"true",
				}
			},
			errContains: "cannot use string as type bool as argument",
		},
		{
			name: "sender is not owner",
			run: func() []interface{} {
				return []interface{}{
					false,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "valid call - enable selfdestruct",
			run: func() []interface{} {
				return []interface{}{
					false,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().NotContains(params.ExtraEIPs, evmtypes.SelfdestructDisableEIP)
			},
		},
		{
			name: "valid call - disable selfdestruct",
			run: func() []interface{} {
				return []interface{}{
					true,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().Contains(params.ExtraEIPs, evmtypes.SelfdestructDisableEIP)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setSelfDestructDisabled")
}

func (s *PrecompileTestSuite) TestGetSelfDestructDisabled() {
	// Enable selfdestruct (remove the EIP) so the getter has a known state.
	setupTestCase := TestCase{
		name: "set up: enable selfdestruct",
		run: func() []interface{} {
			return []interface{}{
				false,
			}
		},
		as:        s.account1.EvmAddr,
		basicPass: true,
	}
	s.RunMethodTestCases([]TestCase{setupTestCase}, "setSelfDestructDisabled")

	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, // Additional argument
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name:      "valid call - reports enabled",
			run:       func() []interface{} { return nil },
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{false},
		},
	}

	s.RunMethodTestCases(testcases, "getSelfDestructDisabled")
}
