package maintenance_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile/maintenance"
)

func (s *PrecompileTestSuite) TestSetPrecompileByteCode() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "sender is not owner",
			run: func() []interface{} {
				return []interface{}{
					common.HexToAddress(maintenance.EvmAddress),
					[]byte{},
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "not a precompile",
			run: func() []interface{} {
				return []interface{}{
					common.HexToAddress("0x0000000000000000000000000000000000000001"),
					[]byte{},
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "address is not a precompile",
		},
		{
			name: "valid call",
			run: func() []interface{} {
				return []interface{}{
					common.HexToAddress(maintenance.EvmAddress),
					common.Hex2Bytes("0xffffffff"),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				account := s.evmKeeper.GetAccount(s.ctx, common.HexToAddress(maintenance.EvmAddress))
				codeHash := account.CodeHash
				code := s.evmKeeper.GetCode(s.ctx, common.BytesToHash(codeHash))
				s.Require().Equal(code, "0xffffffff")
			},
		},
	}

	s.RunMethodTestCases(testcases, "setPrecompileByteCode")
}

// argument count mismatch
// sender not owner
// address is not precompile
// valid call
