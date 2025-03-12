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

	app        *app.Mezo
	poaKeeper  *FakePoaKeeper
	evmKeeper  *FakeEvmKeeper
	bankKeeper *FakeBankKeeper
	ctx        sdk.Context

	account1, account2 Key

	maintenancePrecompile *precompile.Contract
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

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	// init fake keeper
	s.evmKeeper = NewFakeEvmKeeper()
	s.poaKeeper = NewFakePoaKeeper(s.account1.SdkAddr)

	// init app
	s.app = app.Setup(false, nil)
	header := testutil.NewHeader(
		1, time.Now().UTC(), "mezo_31612-1", consAddress, nil, nil,
	)
	s.ctx = s.app.BaseApp.NewContextLegacy(false, header)
}

func (s *PrecompileTestSuite) RunMethodTestCases(testcases []TestCase, methodName string) {
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
			}
			maintenancePrecompile, err := maintenance.NewPrecompile(
				s.poaKeeper,
				s.evmKeeper,
				s.bankKeeper,
				&maintenance.Settings{
					EVM:         true,
					Precompiles: true,
				},
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

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			vmContract.Input = append(vmContract.Input, method.ID...)
			vmContract.Input = append(vmContract.Input, methodInputArgs...)
			vmContract.CallerAddress = tc.as

			output, err := s.maintenancePrecompile.Run(evm, vmContract, false)
			if tc.revert {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
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

type FakePoaKeeper struct {
	owner sdk.AccAddress
}

type FakeEvmKeeper struct {
	params     evmtypes.Params
	accountMap map[common.Address]statedb.Account
	codeMap    map[common.Hash][]byte
}

type FakeBankKeeper struct {
	blockedAddrs map[string]bool
}

func NewFakeBankKeeper() *FakeBankKeeper {
	return &FakeBankKeeper{
		blockedAddrs: make(map[string]bool),
	}
}
func (k *FakeBankKeeper) GetBlockedAddresses() map[string]bool {
	return k.blockedAddrs
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

func (k *FakeEvmKeeper) SetParams(_ sdk.Context, params evmtypes.Params) error {
	k.params = params
	return nil
}

func (k *FakeEvmKeeper) IsCustomPrecompile(address common.Address) bool {
	return address == common.HexToAddress(maintenance.EvmAddress)
}
