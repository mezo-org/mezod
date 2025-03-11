package maintenance_test

import "github.com/ethereum/go-ethereum/common"

func (s *PrecompileTestSuite) TestSetFeeChainSplitterAddress() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid address argument",
			run: func() []interface{} {
				return []interface{}{
					"invalid_address_string",
				}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "sender is not owner",
			run: func() []interface{} {
				return []interface{}{
					common.HexToAddress("0x1234567890AbcdEF1234567890aBcdef12345678"),
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
					common.HexToAddress("0x1234567890AbcdEF1234567890aBcdef12345678"),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				value := s.evmKeeper.GetParams(s.ctx)
				s.Require().Equal("0x1234567890AbcdEF1234567890aBcdef12345678", value.ChainFeeSplitterAddress)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setFeeChainSplitterAddress")
}

func (s *PrecompileTestSuite) TestGetFeeChainSplitterAddress() {
	// Set the address through setFeeChainSplitterAddress
	setupTestCase := TestCase{
		name: "set up: valid set call",
		run: func() []interface{} {
			return []interface{}{
				common.HexToAddress("0x1234567890AbcdEF1234567890aBcdef12345678"),
			}
		},
		as:        s.account1.EvmAddr,
		basicPass: true,
	}
	s.RunMethodTestCases([]TestCase{setupTestCase}, "setFeeChainSplitterAddress")

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
			output:    []interface{}{common.HexToAddress("0x1234567890AbcdEF1234567890aBcdef12345678")},
		},
	}

	s.RunMethodTestCases(testcases, "getFeeChainSplitterAddress")
}
