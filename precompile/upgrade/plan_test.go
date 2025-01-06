package upgrade_test

func (s *PrecompileTestSuite) TestPlan() {
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
			output:    []interface{}{string("")}, // No plan set
		},
	}

	s.RunMethodTestCases(testcases, "plan")
}

func (s *PrecompileTestSuite) TestSubmitPlan() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3, 4,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "keeper returns error",
			run: func() []interface{} {
				return []interface{}{
					"v2.0.0",
					"100",
					"{...}",
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
					"v2.0.0",
					"1000",
					"{...}",
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				plan, err := s.upgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().NoError(err, "expected no error")
				s.Require().Equal(plan.Name, "v2.0.0", "upgrade plan name was not updated in keeper")
				s.Require().Equal(plan.Height, int64(1000), "upgrade plan height was not updated in keeper")
				s.Require().Equal(plan.Info, "{...}", "upgrade plan info was not updated in keeper")
			},
		},
		{
			name: "valid call then cancel",
			run: func() []interface{} {
				return []interface{}{
					"v2.0.0",
					"1000",
					"{...}",
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				plan, err := s.upgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().NoError(err, "expected no error")
				s.Require().Equal(plan.Name, "v2.0.0", "upgrade plan name was not updated in keeper")
				s.Require().Equal(plan.Height, int64(1000), "upgrade plan height was not updated in keeper")
				s.Require().Equal(plan.Info, "{...}", "upgrade plan info was not updated in keeper")
				err = s.upgradeKeeper.ClearUpgradePlan(s.ctx)
				s.Require().NoError(err, "expected no error")
				plan, err = s.upgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().NoError(err, "expected no error")
				s.Require().Empty(plan, "upgrade plan is not empty")
			},
		},
	}

	s.RunMethodTestCases(testcases, "submitPlan")
}

func (s *PrecompileTestSuite) TestCancelPlan() {
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
			name:        "keeper returns error",
			run:         func() []interface{} { return nil },
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "valid call",
			run: func() []interface{} {
				return nil
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				plan, err := s.upgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().NoError(err, "expected no error")
				s.Require().Empty(plan, "upgrade plan is not empty")
			},
		},
	}

	s.RunMethodTestCases(testcases, "cancelPlan")
}
