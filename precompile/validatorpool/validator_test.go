package validatorpool_test

import (
	"github.com/ethereum/go-ethereum/common"
)

func (s *PrecompileTestSuite) TestValidator() {
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
					s.account2.EvmAddr,
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "validator does not exist",
		},
		{
			name: "valid call",
			run: func() []interface{} {
				return []interface{}{
					s.account3.EvmAddr,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{s.account3.ConsPubKeyBytes32(), s.account3.Description},
		},
	}

	s.RunMethodTestCases(testcases, "validator")
}

func (s *PrecompileTestSuite) TestValidators() {
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
				[]common.Address{s.account3.EvmAddr},
			},
		},
	}

	s.RunMethodTestCases(testcases, "validators")
}
