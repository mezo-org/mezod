package maintenance_test

func (s *PrecompileTestSuite) TestSetBridgeLockdown() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					true,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid bridgeIn argument type",
			run: func() []interface{} {
				return []interface{}{
					"true",
					false,
				}
			},
			errContains: "cannot use string as type bool as argument",
		},
		{
			name: "invalid bridgeOut argument type",
			run: func() []interface{} {
				return []interface{}{
					false,
					"true",
				}
			},
			errContains: "cannot use string as type bool as argument",
		},
		{
			name: "sender is neither the owner nor the emergency team",
			run: func() []interface{} {
				return []interface{}{
					true,
					true,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "emergency team address is empty",
			postCheck: func() {
				s.Require().False(s.bridgeKeeper.IsBridgeInPaused(s.ctx))
				s.Require().False(s.bridgeKeeper.IsBridgeOutPaused(s.ctx))
			},
		},
		{
			name: "valid call - the owner enables both directions",
			run: func() []interface{} {
				return []interface{}{
					true,
					true,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(s.bridgeKeeper.IsBridgeInPaused(s.ctx))
				s.Require().True(s.bridgeKeeper.IsBridgeOutPaused(s.ctx))

				topics, arguments := s.requireEmittedEvent("BridgeLockdownSet")
				s.Require().Empty(topics)
				s.Require().Equal([]interface{}{true, true}, arguments)
			},
		},
		{
			name: "valid call - the emergency team enables the bridge-in only",
			run: func() []interface{} {
				err := s.poaKeeper.SetEmergencyTeam(
					s.ctx,
					s.account1.SdkAddr,
					s.account3.SdkAddr,
				)
				s.Require().NoError(err)

				return []interface{}{
					true,
					false,
				}
			},
			as:        s.account3.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(s.bridgeKeeper.IsBridgeInPaused(s.ctx))
				s.Require().False(s.bridgeKeeper.IsBridgeOutPaused(s.ctx))

				_, arguments := s.requireEmittedEvent("BridgeLockdownSet")
				s.Require().Equal([]interface{}{true, false}, arguments)
			},
		},
		{
			name: "sender is not the emergency team",
			run: func() []interface{} {
				return []interface{}{
					false,
					false,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not the emergency team",
			postCheck: func() {
				// The flags keep the values set by the previous test case.
				s.Require().True(s.bridgeKeeper.IsBridgeInPaused(s.ctx))
				s.Require().False(s.bridgeKeeper.IsBridgeOutPaused(s.ctx))
			},
		},
		{
			name: "valid call - the emergency team disables the lockdown",
			run: func() []interface{} {
				return []interface{}{
					false,
					false,
				}
			},
			as:        s.account3.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().False(s.bridgeKeeper.IsBridgeInPaused(s.ctx))
				s.Require().False(s.bridgeKeeper.IsBridgeOutPaused(s.ctx))

				_, arguments := s.requireEmittedEvent("BridgeLockdownSet")
				s.Require().Equal([]interface{}{false, false}, arguments)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setBridgeLockdown")
}

func (s *PrecompileTestSuite) TestGetBridgeLockdown() {
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
			name:      "valid call - no lockdown",
			run:       func() []interface{} { return nil },
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{false, false},
		},
		{
			name: "valid call - bridge-out lockdown only",
			run: func() []interface{} {
				s.bridgeKeeper.SetBridgeOutPaused(s.ctx, true)
				return nil
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{false, true},
		},
		{
			name: "valid call - both directions",
			run: func() []interface{} {
				s.bridgeKeeper.SetBridgeInPaused(s.ctx, true)
				s.bridgeKeeper.SetBridgeOutPaused(s.ctx, true)
				return nil
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{true, true},
			postCheck: func() {
				// A read method must not emit events.
				s.requireNoEmittedEvent()
			},
		},
	}

	s.RunMethodTestCases(testcases, "getBridgeLockdown")
}
