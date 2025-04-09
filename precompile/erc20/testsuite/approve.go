package testsuite

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/encoding"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/erc20"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

func (s *TestSuite) TestApprove() {
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
					s.account1.EvmAddr, "invalid amount",
				}
			},
			errContains: "cannot use string as type ptr as argument",
		},
		{
			name: "approve without existing authorization",
			run: func() []interface{} {
				return []interface{}{
					s.account1.EvmAddr, big.NewInt(100),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 100)),
				)
			},
		},
		{
			name: "approve more with existing single-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 1)),
				)

				return []interface{}{
					s.account1.EvmAddr, big.NewInt(100),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 100)),
				)
			},
		},
		{
			name: "approve less with existing single-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 1000)),
				)

				return []interface{}{
					s.account1.EvmAddr, big.NewInt(100),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 100)),
				)
			},
		},
		{
			name: "approve same with existing single-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 100)),
				)

				return []interface{}{
					s.account1.EvmAddr, big.NewInt(100),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 100)),
				)
			},
		},
		{
			name: "approve more with existing multi-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(
						sdk.NewInt64Coin(s.denom, 1),
						sdk.NewInt64Coin("otherdenom", 2),
					).Sort(),
				)

				return []interface{}{
					s.account1.EvmAddr, big.NewInt(100),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(
						sdk.NewInt64Coin(s.denom, 100),
						sdk.NewInt64Coin("otherdenom", 2),
					).Sort(),
				)
			},
		},
		{
			name: "approve less with existing multi-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(
						sdk.NewInt64Coin(s.denom, 1000),
						sdk.NewInt64Coin("otherdenom", 2),
					).Sort(),
				)

				return []interface{}{
					s.account1.EvmAddr, big.NewInt(100),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(
						sdk.NewInt64Coin(s.denom, 100),
						sdk.NewInt64Coin("otherdenom", 2),
					).Sort(),
				)
			},
		},
		{
			name: "approve same with existing multi-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(
						sdk.NewInt64Coin(s.denom, 100),
						sdk.NewInt64Coin("otherdenom", 2),
					).Sort(),
				)

				return []interface{}{
					s.account1.EvmAddr, big.NewInt(100),
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(
						sdk.NewInt64Coin(s.denom, 100),
						sdk.NewInt64Coin("otherdenom", 2),
					).Sort(),
				)
			},
		},
		{
			name: "delete existing single-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin(s.denom, 100)),
				)

				return []interface{}{
					s.account1.EvmAddr, common.Big0,
				}
			},
			basicPass: true,
			postCheck: func() {
				grants, err := s.app.AuthzKeeper.GranteeGrants(s.ctx, &authz.QueryGranteeGrantsRequest{
					Grantee: s.account1.SdkAddr.String(),
				})
				s.Require().NoError(err, "expected no error querying the grants")
				authzs, err := unpackGrantAuthzs(grants.Grants)
				s.Require().NoError(err, "expected no error unpacking the authorization")
				s.Require().Len(authzs, 0, "expected grant to be deleted")
			},
		},
		{
			name: "delete existing multi-denom authorization",
			run: func() []interface{} {
				s.setupSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(
						sdk.NewInt64Coin(s.denom, 100),
						sdk.NewInt64Coin("otherdenom", 2),
					).Sort(),
				)

				return []interface{}{
					s.account1.EvmAddr, common.Big0,
				}
			},
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account1.SdkAddr,
					s.account2.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin("otherdenom", 2)),
				)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
			}

			erc20Precompile, err := s.precompileFactoryFn(s.app)
			s.Require().NoError(err)
			s.erc20Precompile = erc20Precompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.erc20Precompile.Abi.Methods["approve"]
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
			vmContract.CallerAddress = s.account2.EvmAddr

			output, err := s.erc20Precompile.Run(evm, vmContract, false)
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(true, out[0], "expected different value")

			// we call  the statedb commit here to simulate end of transaction
			// processing and flush the cache context
			s.Require().NoError(evm.StateDB.(*statedb.StateDB).Commit())

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

// Check that athorization exists for a given grantee and granter
// for a given amount.
func (s *TestSuite) requireSendAuthz(grantee, granter sdk.AccAddress, amount sdk.Coins) {
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

// Sets up a send authorization for a given grantee and granter.
func (s *TestSuite) setupSendAuthz(grantee, granter sdk.AccAddress, amount sdk.Coins) {
	authzKeeper := s.app.AuthzKeeper
	expiration := s.ctx.BlockTime().Add(time.Hour * 24 * 365)
	sendAuthz := banktypes.NewSendAuthorization(amount, nil)
	err := sendAuthz.ValidateBasic()
	s.Require().NoError(err, "expected no error validating the grant")

	err = authzKeeper.SaveGrant(s.ctx, grantee.Bytes(), granter.Bytes(), sendAuthz, &expiration)
	s.Require().NoError(err, "expected no error saving the grant")

	grants, err := authzKeeper.GranteeGrants(s.ctx, &authz.QueryGranteeGrantsRequest{
		Grantee: grantee.String(),
	})
	s.Require().NoError(err, "expected no error querying the grants")

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

func (s *TestSuite) TestEmitApprovalEvent() {
	testcases := []struct {
		name    string
		owner   common.Address
		spender common.Address
		amount  *big.Int
	}{
		{
			name:    "pass",
			owner:   s.account1.EvmAddr,
			spender: s.account2.EvmAddr,
			amount:  big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			ae := erc20.NewApprovalEvent(tc.owner, tc.spender, tc.amount)
			args := ae.Arguments()

			s.Require().Len(args, 3)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.owner, args[0].Value)

			// Check the second argument
			s.Require().True(args[1].Indexed)
			s.Require().Equal(tc.spender, args[1].Value)

			// Check the third argument
			s.Require().False(args[2].Indexed)
			s.Require().Equal(tc.amount, args[2].Value)
		})
	}
}
