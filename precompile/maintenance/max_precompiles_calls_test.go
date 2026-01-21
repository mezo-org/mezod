package maintenance_test

func (s *PrecompileTestSuite) TestSetMaxPrecompilesCallsPerExecution() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid max precompiles calls argument type",
			run: func() []interface{} {
				return []interface{}{
					"100",
				}
			},
			errContains: "cannot use string as type uint32 as argument",
		},
		{
			name: "sender is not owner",
			run: func() []interface{} {
				return []interface{}{
					uint32(20),
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "valid call - set to 20",
			run: func() []interface{} {
				return []interface{}{
					uint32(20),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().Equal(uint32(20), params.MaxPrecompilesCallsPerExecution)
			},
		},
		{
			name: "value below minimum",
			run: func() []interface{} {
				return []interface{}{
					uint32(0),
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "max precompiles calls per execution must be at least 1",
		},
		{
			name: "valid call - set to minimum value of 1",
			run: func() []interface{} {
				return []interface{}{
					uint32(1),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().Equal(uint32(1), params.MaxPrecompilesCallsPerExecution)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setMaxPrecompilesCallsPerExecution")
}

func (s *PrecompileTestSuite) TestGetMaxPrecompilesCallsPerExecution() {
	// Set the max precompiles calls through setMaxPrecompilesCallsPerExecution
	setupTestCase := TestCase{
		name: "set up: valid set call",
		run: func() []interface{} {
			return []interface{}{
				uint32(50),
			}
		},
		as:        s.account1.EvmAddr,
		basicPass: true,
	}
	s.RunMethodTestCases([]TestCase{setupTestCase}, "setMaxPrecompilesCallsPerExecution")

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
			output:    []interface{}{uint32(50)},
		},
	}

	s.RunMethodTestCases(testcases, "getMaxPrecompilesCallsPerExecution")
}
