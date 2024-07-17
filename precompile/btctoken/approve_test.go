package btctoken_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/core/vm"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v12/app"
	"github.com/evmos/evmos/v12/encoding"
	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/precompile/btctoken"
	"github.com/evmos/evmos/v12/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestApprove() {
	amount := int64(100)

	testcases := []struct {
		name        string
		run         func() []interface{}
		postCheck   func()
		basicPass   bool
		errContains string
	}{
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
			name: "invalid address",
			run: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2),
				}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid amount",
			run: func() []interface{} {
				return []interface{}{
					s.account1.Addr, "invalid amount",
				}
			},
			errContains: "cannot use string as type ptr as argument",
		},

		// TODO: Needs more investigation. When using the ABI.Pack method,
		// a negative number is converted to a max uint256 and is being passed
		// as a max number to the precompiled approve function.
		// {
		// 	name: "fail - negative amount",
		// 	run: func() []interface{} {
		// 		return []interface{}{
		// 			s.account1.Addr, big.NewInt(-1),
		// 		}
		// 	},
		// 	basicPass: true,
		// 	errContains: "cannot approve negative values",
		// },

		// TODO: Needs more investigation. When using the ABI.Pack method,
		// an overflow number hits the max limit and resets.
		// {
		// 	name: "fail - approve uint256 overflow",
		// 	run: func() []interface{} {
		// 		return []interface{}{
		// 			s.account1.Addr, new(big.Int).Add(abi.MaxUint256, common.Big1),
		// 		}
		// 	},
		// 	basicPass: true,
		// 	errContains: "causes integer overflow",
		// },

		{
			name: "pass - approve without existing authorization",
			run: func() []interface{} {
				return []interface{}{
					s.account1.Addr, big.NewInt(amount),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.AccAddr,
					s.account2.AccAddr,
					sdk.NewCoins(sdk.NewInt64Coin("abtc", int64(amount))),
				)
			},
		},

		// TODO: add more tests
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {

			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}

			bankKeeper := s.app.BankKeeper
			authzKeeper := s.app.AuthzKeeper

			btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper, authzKeeper)
			s.Require().NoError(err)
			s.btcTokenPrecompile = btcTokenPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.btcTokenPrecompile.Abi.Methods["approve"]
			var methodInputArgs []byte
			methodInputArgs, err = method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			// These first 4 bytes correspond to the method ID (first 4 bytes of the
			// Keccak-256 hash of the function signature).
			// In this case a function signature is 'function approve(address spender, uint256 value)'
			vmContract.Input = append([]byte{0x09, 0x5e, 0xa7, 0xb3}, methodInputArgs...)
			vmContract.CallerAddress = s.account2.Addr

			output, err := s.btcTokenPrecompile.Run(evm, vmContract, false)
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(true, out[0], "expected different value")

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

// Check that athorization exists for a given grantee and granter
// for a given amount.
func (s *PrecompileTestSuite) requireSendAuthz(grantee, granter sdk.AccAddress, amount sdk.Coins) {
	authzKeeper := s.app.AuthzKeeper
	grants, err := authzKeeper.GranteeGrants(s.ctx, &authz.QueryGranteeGrantsRequest{
		Grantee: grantee.String(),
	})
	s.Require().NoError(err, "expected no error querying the grants")

	s.Require().Len(grants.Grants, 1, "expected one grant")
	s.Require().Equal(grantee.String(), grants.Grants[0].Grantee, "expected different grantee")
	s.Require().Equal(granter.String(), grants.Grants[0].Granter, "expected different granter")

	authzs, err := unpackGrantAuthzs(grants.Grants)
	s.Require().NoError(err, "expected no error unpacking the authorization")
	s.Require().Len(authzs, 1, "expected one authorization")

	sendAuthz, ok := authzs[0].(*banktypes.SendAuthorization)
	s.Require().True(ok, "expected send authorization")

	s.Require().Equal(amount, sendAuthz.SpendLimit, "expected different spend limit amount")
}

// Unpacks the given grant authorization.
func unpackGrantAuthzs(grantAuthzs []*authz.GrantAuthorization) ([]authz.Authorization, error) {
	encodingCfg := encoding.MakeConfig(app.ModuleBasics)

	auths := make([]authz.Authorization, 0, len(grantAuthzs))
	for _, grantAuthz := range grantAuthzs {
		var auth authz.Authorization
		err := encodingCfg.InterfaceRegistry.UnpackAny(grantAuthz.Authorization, &auth)
		if err != nil {
			return nil, err
		}

		auths = append(auths, auth)
	}

	return auths, nil
}
