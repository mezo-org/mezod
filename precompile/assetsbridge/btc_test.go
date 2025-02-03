package assetsbridge_test

func (s *PrecompileTestSuite) TestGetSourceBTCToken() {
	testcases := []TestCase{
		{
			name:      "happy path",
			run:       func() []interface{} { return nil },
			basicPass: true,
			output:    []interface{}{testTBTCAddress},
		},
	}

	s.RunMethodTestCases(testcases, "getSourceBTCToken")
}
