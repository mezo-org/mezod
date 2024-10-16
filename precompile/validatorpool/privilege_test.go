package validatorpool_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile/validatorpool"
)

func (s *PrecompileTestSuite) TestAddPrivilege() {
	testcases := []TestCase{
		{
			name:        "empty args",
			run:         func() []interface{} { return nil },
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "unknown privilege id",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr, s.account3.EvmAddr},
					uint8(100),
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unknown privilege id",
		},
		{
			name: "keeper returns error",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr, s.account3.EvmAddr},
					uint8(1),
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
					[]common.Address{s.account2.EvmAddr, s.account3.EvmAddr},
					uint8(1),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				operators := s.keeper.GetValidatorsOperatorsByPrivilege(
					s.ctx,
					"bridge",
				)

				s.Require().Equal(
					[]sdk.ValAddress{
						sdk.ValAddress(s.account2.SdkAddr),
						sdk.ValAddress(s.account3.SdkAddr),
					},
					operators,
					"privilege was not updated in keeper",
				)
			},
		},
	}

	s.RunMethodTestCases(testcases, "addPrivilege")
}

func (s *PrecompileTestSuite) TestRemovePrivilege() {
	testcases := []TestCase{
		{
			name:        "empty args",
			run:         func() []interface{} { return nil },
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "unknown privilege id",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account4.EvmAddr},
					uint8(100),
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unknown privilege id",
		},
		{
			name: "keeper returns error",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account4.EvmAddr},
					uint8(1),
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
					[]common.Address{s.account4.EvmAddr},
					uint8(1),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// There is one operator with the privilege "bridge"
				// at the beginning of the test (account4). After the
				// test, there should be none.
				operators := s.keeper.GetValidatorsOperatorsByPrivilege(
					s.ctx,
					"bridge",
				)

				s.Require().Equal(
					0,
					len(operators),
					"privilege was not updated in keeper",
				)
			},
		},
	}

	s.RunMethodTestCases(testcases, "removePrivilege")
}

func (s *PrecompileTestSuite) TestValidatorsByPrivilege() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "unknown privilege id",
			run: func() []interface{} {
				return []interface{}{
					uint8(100),
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unknown privilege id",
		},
		{
			name: "valid call",
			run: func() []interface{} {
				return []interface{}{
					uint8(1),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{[]common.Address{s.account4.EvmAddr}},
		},
	}

	s.RunMethodTestCases(testcases, "validatorsByPrivilege")
}

type privilegeDescriptor = struct {
	//nolint:revive,stylecheck
	Id   uint8  "json:\"id\""
	Name string "json:\"name\""
}

func (s *PrecompileTestSuite) TestPrivileges() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name:      "valid call",
			run:       func() []interface{} { return nil },
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{[]privilegeDescriptor{{1, "bridge"}}},
		},
	}

	s.RunMethodTestCases(testcases, "privileges")
}

func (s *PrecompileTestSuite) TestEmitPrivilegeAddedEvent() {
	e := validatorpool.NewPrivilegeAddedEvent(s.account1.EvmAddr, uint8(1))
	args := e.Arguments()

	s.Require().Len(args, 2)

	// Check the first argument
	s.Require().True(args[0].Indexed)
	s.Require().Equal(s.account1.EvmAddr, args[0].Value)

	// Check the second argument
	s.Require().True(args[1].Indexed)
	s.Require().Equal(uint8(1), args[1].Value)
}

func (s *PrecompileTestSuite) TestEmitPrivilegeRemovedEvent() {
	e := validatorpool.NewPrivilegeRemovedEvent(s.account1.EvmAddr, uint8(1))
	args := e.Arguments()

	s.Require().Len(args, 2)

	// Check the first argument
	s.Require().True(args[0].Indexed)
	s.Require().Equal(s.account1.EvmAddr, args[0].Value)

	// Check the second argument
	s.Require().True(args[1].Indexed)
	s.Require().Equal(uint8(1), args[1].Value)
}
