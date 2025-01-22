package priceoracle_test

func (s *PrecompileTestSuite) TestDecimals() {
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
			name:      "happy path",
			run:       func() []interface{} { return nil },
			basicPass: true,
			output:    []interface{}{uint8(18)},
		},
	}

	s.RunMethodTestCases(testcases, "decimals")
}
