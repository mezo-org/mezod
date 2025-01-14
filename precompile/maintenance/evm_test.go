package maintenance_test

func (s *PrecompileTestSuite) TestSetSupportNonEIP155Txs() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "sender is not owner",
			run: func() []interface{} {
				return []interface{}{
					true,
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
					true,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				value := s.evmKeeper.GetParams(s.ctx)
				s.Require().Equal(true, value.AllowUnprotectedTxs)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setSupportNonEIP155Txs")
}

func (s *PrecompileTestSuite) TestGetSupportNonEIP155Txs() {
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
			name:      "valid call",
			run:       func() []interface{} { return nil },
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{false},
		},
	}

	s.RunMethodTestCases(testcases, "getSupportNonEIP155Txs")
}
