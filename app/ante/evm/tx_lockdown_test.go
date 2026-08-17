package evm_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethante "github.com/mezo-org/mezod/app/ante/evm"
	testutiltx "github.com/mezo-org/mezod/testutil/tx"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

// fakeTxLockdownEVMKeeper answers the only EVM keeper call the transaction
// lockdown decorator makes. Every other call panics on the nil interface.
type fakeTxLockdownEVMKeeper struct {
	ethante.EVMKeeper
	params evmtypes.Params
}

func (k fakeTxLockdownEVMKeeper) GetParams(_ sdk.Context) evmtypes.Params {
	return k.params
}

// fakeTxLockdownPoaKeeper serves the x/poa role addresses from memory.
type fakeTxLockdownPoaKeeper struct {
	owner         sdk.AccAddress
	emergencyTeam sdk.AccAddress
}

func (k fakeTxLockdownPoaKeeper) GetOwner(_ sdk.Context) sdk.AccAddress {
	return k.owner
}

func (k fakeTxLockdownPoaKeeper) GetEmergencyTeam(_ sdk.Context) sdk.AccAddress {
	return k.emergencyTeam
}

// newTxLockdownTx builds a signed-shape Ethereum transaction message with the
// given sender and target. A nil target means contract creation.
func newTxLockdownTx(sender common.Address, target *common.Address) sdk.Tx {
	msg := evmtypes.NewTx(&evmtypes.EvmTxArgs{
		Nonce:    1,
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
		Amount:   big.NewInt(0),
		To:       target,
	})
	// The signature verification decorator runs earlier and sets this field.
	msg.From = sender.Hex()

	return msg
}

func TestTxLockdownAllowed(t *testing.T) {
	owner := testutiltx.GenerateAddress()
	team := testutiltx.GenerateAddress()
	thirdParty := testutiltx.GenerateAddress()
	fourthParty := testutiltx.GenerateAddress()
	targetA := testutiltx.GenerateAddress()
	targetB := testutiltx.GenerateAddress()

	bothRoles := []common.Address{owner, team}
	ownerOnly := []common.Address{owner}

	testCases := []struct {
		name         string
		sender       common.Address
		target       *common.Address
		roles        []common.Address
		extraSenders []common.Address
		extraTargets []common.Address
		expAllowed   bool
	}{
		{
			name:       "the owner is allowed as sender",
			sender:     owner,
			target:     &thirdParty,
			roles:      bothRoles,
			expAllowed: true,
		},
		{
			name:       "the emergency team is allowed as sender",
			sender:     team,
			target:     &thirdParty,
			roles:      bothRoles,
			expAllowed: true,
		},
		{
			name:       "the owner is allowed as target",
			sender:     thirdParty,
			target:     &owner,
			roles:      bothRoles,
			expAllowed: true,
		},
		{
			name:       "the emergency team is allowed as target",
			sender:     thirdParty,
			target:     &team,
			roles:      bothRoles,
			expAllowed: true,
		},
		{
			name:       "an unset emergency team is not a role",
			sender:     team,
			target:     &thirdParty,
			roles:      ownerOnly,
			expAllowed: false,
		},
		{
			name:       "an unset emergency team is not a role target",
			sender:     thirdParty,
			target:     &team,
			roles:      ownerOnly,
			expAllowed: false,
		},
		{
			name:       "the owner passes while the emergency team is unset",
			sender:     owner,
			target:     &thirdParty,
			roles:      ownerOnly,
			expAllowed: true,
		},
		{
			name:       "both dimensions empty rejects a third party",
			sender:     thirdParty,
			target:     &fourthParty,
			roles:      bothRoles,
			expAllowed: false,
		},
		{
			name:         "empty senders let any sender reach a listed target",
			sender:       thirdParty,
			target:       &targetA,
			roles:        bothRoles,
			extraTargets: []common.Address{targetA},
			expAllowed:   true,
		},
		{
			name:         "empty senders reject another target",
			sender:       thirdParty,
			target:       &targetB,
			roles:        bothRoles,
			extraTargets: []common.Address{targetA},
			expAllowed:   false,
		},
		{
			name:         "empty senders reject contract creation",
			sender:       thirdParty,
			target:       nil,
			roles:        bothRoles,
			extraTargets: []common.Address{targetA},
			expAllowed:   false,
		},
		{
			name:         "empty targets let a listed sender reach any target",
			sender:       thirdParty,
			target:       &targetB,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			expAllowed:   true,
		},
		{
			name:         "empty targets let a listed sender deploy a contract",
			sender:       thirdParty,
			target:       nil,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			expAllowed:   true,
		},
		{
			name:         "empty targets reject another sender",
			sender:       fourthParty,
			target:       &targetB,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			expAllowed:   false,
		},
		{
			name:         "the listed pair passes",
			sender:       thirdParty,
			target:       &targetA,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			extraTargets: []common.Address{targetA},
			expAllowed:   true,
		},
		{
			name:         "the listed sender to another target is rejected",
			sender:       thirdParty,
			target:       &targetB,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			extraTargets: []common.Address{targetA},
			expAllowed:   false,
		},
		{
			name:         "another sender to the listed target is rejected",
			sender:       fourthParty,
			target:       &targetA,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			extraTargets: []common.Address{targetA},
			expAllowed:   false,
		},
		{
			name:         "the listed sender cannot deploy while targets are set",
			sender:       thirdParty,
			target:       nil,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			extraTargets: []common.Address{targetA},
			expAllowed:   false,
		},
		{
			name:         "a role sender deploys while targets are set",
			sender:       owner,
			target:       nil,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			extraTargets: []common.Address{targetA},
			expAllowed:   true,
		},
		{
			name:       "a third party cannot deploy without an allowlist",
			sender:     fourthParty,
			target:     nil,
			roles:      bothRoles,
			expAllowed: false,
		},
		{
			name:         "a role target passes while the allowlist rejects the pair",
			sender:       fourthParty,
			target:       &owner,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty},
			extraTargets: []common.Address{targetA},
			expAllowed:   true,
		},
		{
			name:         "several listed senders and targets pass in every pair",
			sender:       fourthParty,
			target:       &targetB,
			roles:        bothRoles,
			extraSenders: []common.Address{thirdParty, fourthParty},
			extraTargets: []common.Address{targetA, targetB},
			expAllowed:   true,
		},
		{
			name:       "no roles at all rejects everyone",
			sender:     thirdParty,
			target:     &fourthParty,
			roles:      nil,
			expAllowed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allowed := ethante.TxLockdownAllowed(
				tc.sender,
				tc.target,
				tc.roles,
				tc.extraSenders,
				tc.extraTargets,
			)

			require.Equal(t, tc.expAllowed, allowed, tc.name)
		})
	}
}

func TestEthTxLockdownDecorator(t *testing.T) {
	owner := testutiltx.GenerateAddress()
	team := testutiltx.GenerateAddress()
	thirdParty := testutiltx.GenerateAddress()
	fourthParty := testutiltx.GenerateAddress()
	targetA := testutiltx.GenerateAddress()

	// lockdownParams returns EVM parameters with the transaction lockdown in
	// the given state.
	lockdownParams := func(enabled bool, senders, targets []string) evmtypes.Params {
		params := evmtypes.DefaultParams()
		params.TxLockdownEnabled = enabled
		params.TxLockdownSenders = senders
		params.TxLockdownTargets = targets
		return params
	}

	testCases := []struct {
		name          string
		params        evmtypes.Params
		emergencyTeam sdk.AccAddress
		sender        common.Address
		target        *common.Address
		expNextCalled bool
	}{
		{
			name:          "a disabled lockdown lets every transaction through",
			params:        lockdownParams(false, nil, nil),
			sender:        thirdParty,
			target:        &fourthParty,
			expNextCalled: true,
		},
		{
			name:          "an enabled lockdown rejects a third party",
			params:        lockdownParams(true, nil, nil),
			sender:        thirdParty,
			target:        &fourthParty,
			expNextCalled: false,
		},
		{
			name:          "an enabled lockdown lets the owner through",
			params:        lockdownParams(true, nil, nil),
			sender:        owner,
			target:        &fourthParty,
			expNextCalled: true,
		},
		{
			name:          "an enabled lockdown lets the emergency team through",
			params:        lockdownParams(true, nil, nil),
			emergencyTeam: sdk.AccAddress(team.Bytes()),
			sender:        team,
			target:        &fourthParty,
			expNextCalled: true,
		},
		{
			name:          "an enabled lockdown lets a transaction to the owner through",
			params:        lockdownParams(true, nil, nil),
			sender:        thirdParty,
			target:        &owner,
			expNextCalled: true,
		},
		{
			name:          "an unset emergency team grants nothing",
			params:        lockdownParams(true, nil, nil),
			sender:        team,
			target:        &fourthParty,
			expNextCalled: false,
		},
		{
			name:          "an all-zero emergency team grants nothing",
			params:        lockdownParams(true, nil, nil),
			emergencyTeam: sdk.AccAddress(make([]byte, 20)),
			sender:        common.Address{},
			target:        &fourthParty,
			expNextCalled: false,
		},
		{
			name: "an enabled lockdown lets an allowlisted pair through",
			params: lockdownParams(
				true,
				[]string{thirdParty.Hex()},
				[]string{targetA.Hex()},
			),
			sender:        thirdParty,
			target:        &targetA,
			expNextCalled: true,
		},
		{
			name: "an enabled lockdown rejects an allowlisted sender elsewhere",
			params: lockdownParams(
				true,
				[]string{thirdParty.Hex()},
				[]string{targetA.Hex()},
			),
			sender:        thirdParty,
			target:        &fourthParty,
			expNextCalled: false,
		},
		{
			name:          "an enabled lockdown rejects a third-party deployment",
			params:        lockdownParams(true, nil, nil),
			sender:        thirdParty,
			target:        nil,
			expNextCalled: false,
		},
		{
			name:          "an enabled lockdown lets the owner deploy",
			params:        lockdownParams(true, nil, nil),
			sender:        owner,
			target:        nil,
			expNextCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decorator := ethante.NewEthTxLockdownDecorator(
				fakeTxLockdownEVMKeeper{params: tc.params},
				fakeTxLockdownPoaKeeper{
					owner:         sdk.AccAddress(owner.Bytes()),
					emergencyTeam: tc.emergencyTeam,
				},
			)

			nextCalled := false
			next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
				nextCalled = true
				return ctx, nil
			}

			_, err := decorator.AnteHandle(
				sdk.Context{},
				newTxLockdownTx(tc.sender, tc.target),
				false,
				next,
			)

			if tc.expNextCalled {
				require.NoError(t, err, tc.name)
				require.True(t, nextCalled, "expected the next handler to run")
				return
			}

			require.Error(t, err, tc.name)
			require.ErrorIs(t, err, errortypes.ErrUnauthorized)
			require.ErrorContains(t, err, "rejected by the transaction lockdown")
			require.False(t, nextCalled, "expected the next handler to be skipped")
		})
	}
}

func TestEthTxLockdownDecoratorRejectsOtherMessages(t *testing.T) {
	decorator := ethante.NewEthTxLockdownDecorator(
		fakeTxLockdownEVMKeeper{
			params: func() evmtypes.Params {
				params := evmtypes.DefaultParams()
				params.TxLockdownEnabled = true
				return params
			}(),
		},
		fakeTxLockdownPoaKeeper{},
	)

	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}

	_, err := decorator.AnteHandle(sdk.Context{}, &notAnEthTx{}, false, next)

	require.ErrorIs(t, err, errortypes.ErrUnknownRequest)
}

// notAnEthTx is a transaction that carries no Ethereum message.
type notAnEthTx struct{}

func (tx *notAnEthTx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{&evmtypes.MsgUpdateParams{}}
}

func (tx *notAnEthTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}
