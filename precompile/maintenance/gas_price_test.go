package maintenance_test

import (
	"math/big"
)

func (s *PrecompileTestSuite) TestSetMinGasPrice() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid min gas price argument type",
			run: func() []interface{} {
				return []interface{}{
					"100",
				}
			},
			errContains: "cannot use string as type ptr as argument",
		},
		{
			name: "sender is not owner",
			run: func() []interface{} {
				return []interface{}{
					big.NewInt(123456789),
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			// We only check for zero value as go-ethereum argument packing
			// casts the value to uint256 and any negative value passed
			// here will underflow to a very large positive value.
			name: "min gas price is zero",
			run: func() []interface{} {
				return []interface{}{
					big.NewInt(0),
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "min gas price must be positive",
		},

		{
			name: "valid call",
			run: func() []interface{} {
				return []interface{}{
					big.NewInt(123456789),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.feeMarketKeeper.GetParams(s.ctx)
				s.Require().Equal("123456789.000000000000000000", params.MinGasPrice.String())
			},
		},
	}

	s.RunMethodTestCases(testcases, "setMinGasPrice")
}

func (s *PrecompileTestSuite) TestGetMinGasPrice() {
	// Set the min gas price through setMinGasPrice
	setupTestCase := TestCase{
		name: "set up: valid set call",
		run: func() []interface{} {
			return []interface{}{
				big.NewInt(123456789),
			}
		},
		as:        s.account1.EvmAddr,
		basicPass: true,
	}
	s.RunMethodTestCases([]TestCase{setupTestCase}, "setMinGasPrice")

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
			name:      "valid call",
			run:       func() []interface{} { return nil },
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{big.NewInt(123456789)},
		},
	}

	s.RunMethodTestCases(testcases, "getMinGasPrice")
}
