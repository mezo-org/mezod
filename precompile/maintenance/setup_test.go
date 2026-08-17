package maintenance_test

import (
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/maintenance"
	"github.com/mezo-org/mezod/testutil"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	feemarkettypes "github.com/mezo-org/mezod/x/feemarket/types"
	"github.com/stretchr/testify/suite"
)

type TestCase struct {
	// name of test
	name string
	// run function to determine inputs
	run func() []interface{}
	// address to execute method as (msg.sender)
	as common.Address
	// function to perform any post checks
	postCheck func()
	// true if expected good inputs, false if expect an input related error (set errContains)
	basicPass bool
	// set true if expecting an execution error (set errContains)
	revert bool
	// define expect error
	errContains string
	// define expected outputs
	output []interface{}
}

type Key struct {
	EvmAddr    common.Address
	SdkAddr    sdk.AccAddress
	ConsPubKey cryptotypes.PubKey
	Priv       cryptotypes.PrivKey
}

type PrecompileTestSuite struct {
	suite.Suite

	app             *app.Mezo
	poaKeeper       *FakePoaKeeper
	evmKeeper       *FakeEvmKeeper
	feeMarketKeeper *FakeFeeMarketKeeper
	bridgeKeeper    *FakeBridgeKeeper
	ctx             sdk.Context

	account1, account2, account3 Key

	maintenancePrecompile *precompile.Contract
	stateDB               *statedb.StateDB
}

// latestSettings returns the settings of the latest maintenance precompile
// version.
func latestSettings() *maintenance.Settings {
	return &maintenance.Settings{
		EVM:                 true,
		Precompiles:         true,
		ChainFeeSplitter:    true,
		GasPrice:            true,
		MaxPrecompilesCalls: true,
		SelfDestruct:        true,
		EmergencyControls:   true,
	}
}

func NewKey() Key {
	addr, privKey := utiltx.NewAddrKey()
	// Generate a consPubKey
	tmpk := ed25519.GenPrivKey().PubKey()
	consPubKey, err := cryptocdc.FromCmtPubKeyInterface(tmpk)
	if err != nil {
		panic(err)
	}

	sdkAddr := sdk.AccAddress(addr.Bytes())

	return Key{
		EvmAddr:    addr,
		SdkAddr:    sdkAddr,
		Priv:       privKey,
		ConsPubKey: consPubKey,
	}
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	// accounts
	s.account1 = NewKey() // owner account
	s.account2 = NewKey() // non-owner account
	s.account3 = NewKey() // emergency team account

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	// init fake keeper
	s.evmKeeper = NewFakeEvmKeeper()
	s.poaKeeper = NewFakePoaKeeper(s.account1.SdkAddr)
	s.feeMarketKeeper = NewFakeFeeMarketKeeper()
	s.bridgeKeeper = NewFakeBridgeKeeper()

	// init app
	s.app = app.Setup(false, nil)
	header := testutil.NewHeader(
		1, time.Now().UTC(), "mezo_31612-1", consAddress, nil, nil,
	)
	s.ctx = s.app.BaseApp.NewContextLegacy(false, header)
}

func (s *PrecompileTestSuite) RunMethodTestCases(testcases []TestCase, methodName string) {
	s.RunMethodTestCasesWithSettings(testcases, methodName, latestSettings())
}

func (s *PrecompileTestSuite) RunMethodTestCasesWithSettings(
	testcases []TestCase,
	methodName string,
	settings *maintenance.Settings,
) {
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			stateDB := statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{})
			s.stateDB = stateDB
			evm := &vm.EVM{
				StateDB: stateDB,
			}
			maintenancePrecompile, err := maintenance.NewPrecompile(
				s.poaKeeper,
				s.evmKeeper,
				s.feeMarketKeeper,
				s.bridgeKeeper,
				settings,
			)
			s.Require().NoError(err)
			s.maintenancePrecompile = maintenancePrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.maintenancePrecompile.Abi.Methods[methodName]
			var methodInputArgs []byte
			methodInputArgs, err = method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewPrecompile(tc.as, common.Address{}, nil, 0)
			vmContract.Input = append(vmContract.Input, method.ID...)
			vmContract.Input = append(vmContract.Input, methodInputArgs...)

			output, err := s.maintenancePrecompile.Run(evm, vmContract, false)
			if tc.revert {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")

				if tc.postCheck != nil {
					tc.postCheck()
				}

				return
			}
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			for i, expected := range tc.output {
				s.Require().Equal(expected, out[i], "expected different value")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

// requireEmittedEvent asserts that the last precompile call emitted exactly one
// event with the given name. It returns the indexed argument topics and the
// unpacked non-indexed arguments.
func (s *PrecompileTestSuite) requireEmittedEvent(
	eventName string,
) ([]common.Hash, []interface{}) {
	logs := s.stateDB.Logs()
	s.Require().Len(logs, 1, "expected a single emitted event")

	log := logs[0]
	s.Require().Equal(common.HexToAddress(maintenance.EvmAddress), log.Address)

	event, ok := s.maintenancePrecompile.Abi.Events[eventName]
	s.Require().True(ok, "event %s not found in the ABI", eventName)
	s.Require().Equal(event.ID, log.Topics[0])

	arguments, err := event.Inputs.NonIndexed().Unpack(log.Data)
	s.Require().NoError(err)

	return log.Topics[1:], arguments
}

// requireNoEmittedEvent asserts that the last precompile call emitted no events.
func (s *PrecompileTestSuite) requireNoEmittedEvent() {
	s.Require().Empty(s.stateDB.Logs())
}

type FakePoaKeeper struct {
	owner         sdk.AccAddress
	emergencyTeam sdk.AccAddress
}

type FakeEvmKeeper struct {
	params     evmtypes.Params
	accountMap map[common.Address]statedb.Account
	codeMap    map[common.Hash][]byte
}

func NewFakePoaKeeper(owner sdk.AccAddress) *FakePoaKeeper {
	return &FakePoaKeeper{
		owner: owner,
	}
}

func (k *FakePoaKeeper) CheckOwner(_ sdk.Context, sender sdk.AccAddress) error {
	if !sender.Equals(k.owner) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not owner",
		)
	}
	return nil
}

func (k *FakePoaKeeper) CheckOwnerOrEmergencyTeam(
	ctx sdk.Context,
	sender sdk.AccAddress,
) error {
	if k.CheckOwner(ctx, sender) == nil {
		return nil
	}

	if k.emergencyTeam.Empty() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"emergency team address is empty",
		)
	}

	if !sender.Equals(k.emergencyTeam) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not the emergency team",
		)
	}

	return nil
}

func (k *FakePoaKeeper) GetEmergencyTeam(_ sdk.Context) sdk.AccAddress {
	return k.emergencyTeam
}

func (k *FakePoaKeeper) SetEmergencyTeam(
	ctx sdk.Context,
	sender sdk.AccAddress,
	emergencyTeam sdk.AccAddress,
) error {
	if err := k.CheckOwner(ctx, sender); err != nil {
		return err
	}

	k.setEmergencyTeam(emergencyTeam)

	return nil
}

// setEmergencyTeam mirrors the keeper's handling of the zero address: it
// revokes the role.
func (k *FakePoaKeeper) setEmergencyTeam(emergencyTeam sdk.AccAddress) {
	if emergencyTeam.Empty() || emergencyTeam.Equals(sdk.AccAddress(make([]byte, 20))) {
		k.emergencyTeam = nil
		return
	}

	k.emergencyTeam = emergencyTeam
}

type FakeBridgeKeeper struct {
	bridgeInPaused  bool
	bridgeOutPaused bool
}

func NewFakeBridgeKeeper() *FakeBridgeKeeper {
	return &FakeBridgeKeeper{}
}

func (k *FakeBridgeKeeper) IsBridgeInPaused(_ sdk.Context) bool {
	return k.bridgeInPaused
}

func (k *FakeBridgeKeeper) SetBridgeInPaused(_ sdk.Context, isPaused bool) {
	k.bridgeInPaused = isPaused
}

func (k *FakeBridgeKeeper) IsBridgeOutPaused(_ sdk.Context) bool {
	return k.bridgeOutPaused
}

func (k *FakeBridgeKeeper) SetBridgeOutPaused(_ sdk.Context, isPaused bool) {
	k.bridgeOutPaused = isPaused
}

func NewFakeEvmKeeper() *FakeEvmKeeper {
	return &FakeEvmKeeper{
		params:     evmtypes.DefaultParams(),
		accountMap: make(map[common.Address]statedb.Account),
		codeMap:    make(map[common.Hash][]byte),
	}
}

func (k *FakeEvmKeeper) GetAccount(_ sdk.Context, addr common.Address) *statedb.Account {
	account, ok := k.accountMap[addr]
	if ok {
		return &account
	}
	return nil
}

func (k *FakeEvmKeeper) SetAccount(_ sdk.Context, addr common.Address, account statedb.Account) error {
	k.accountMap[addr] = account
	return nil
}

func (k *FakeEvmKeeper) SetCode(_ sdk.Context, codeHash []byte, code []byte) {
	if len(code) > 0 {
		k.codeMap[common.BytesToHash(codeHash)] = code
	} else {
		delete(k.codeMap, common.BytesToHash(codeHash))
	}
}

func (k *FakeEvmKeeper) GetCode(_ sdk.Context, codeHash common.Hash) []byte {
	code, ok := k.codeMap[codeHash]
	if !ok {
		return []byte{}
	}
	return code
}

func (k *FakeEvmKeeper) GetParams(_ sdk.Context) (params evmtypes.Params) {
	return k.params
}

// SetParams mirrors the real keeper and rejects parameters that do not pass
// validation.
func (k *FakeEvmKeeper) SetParams(_ sdk.Context, params evmtypes.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	k.params = params
	return nil
}

func (k *FakeEvmKeeper) IsCustomPrecompile(address common.Address) bool {
	return address == common.HexToAddress(maintenance.EvmAddress)
}

type FakeFeeMarketKeeper struct {
	params feemarkettypes.Params
}

func NewFakeFeeMarketKeeper() *FakeFeeMarketKeeper {
	return &FakeFeeMarketKeeper{
		params: feemarkettypes.DefaultParams(),
	}
}

func (k *FakeFeeMarketKeeper) GetParams(_ sdk.Context) (params feemarkettypes.Params) {
	return k.params
}

func (k *FakeFeeMarketKeeper) SetParams(_ sdk.Context, params feemarkettypes.Params) error {
	k.params = params
	return nil
}
