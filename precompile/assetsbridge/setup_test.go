package assetsbridge_test

import (
	"bytes"
	"errors"
	"math/big"
	"slices"
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/mezo-org/mezod/precompile/assetsbridge"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/testutil"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/stretchr/testify/suite"
)

var testTBTCAddress = common.HexToAddress("0x517f2982701695D4E52f1ECFBEf3ba31Df470161")

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

	poaKeeper    *FakePoaKeeper
	bridgeKeeper *FakeBridgeKeeper
	app          *app.Mezo
	ctx          sdk.Context

	account1, account2 Key

	assetsBridgePrecompile *precompile.Contract
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

	s.poaKeeper = NewFakePoaKeeper(s.account1.SdkAddr)
	s.bridgeKeeper = NewFakeBridgeKeeper(testTBTCAddress.Bytes())

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

			// Create dummy keepers for the test - these are minimal implementations
			// For bridge_out tests, use the more complete implementations in bridge_out_test.go
			evmKeeper := &FakeEvmKeeper{}
			authzKeeper := &FakeAuthzKeeper{}

			assetsBridgePrecompile, err := assetsbridge.NewPrecompile(
				s.poaKeeper,
				s.bridgeKeeper,
				evmKeeper,
				authzKeeper,
				&assetsbridge.Settings{
					Observability:   true,
					BTCManagement:   true,
					ERC20Management: true,
					SequenceTipView: true,
				},
			)
			s.Require().NoError(err)
			s.assetsBridgePrecompile = assetsBridgePrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.assetsBridgePrecompile.Abi.Methods[methodName]
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

			output, err := s.assetsBridgePrecompile.Run(evm, vmContract, false)
			if tc.revert {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			for i, expected := range tc.output {
				if expected, ok := expected.(*big.Int); ok {
					out, _ := out[i].(*big.Int)
					// if this is big.Int, compare them differently
					s.Require().True(expected.Cmp(out) == 0, "expected different value")
					continue
				}
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

type FakeBridgeKeeper struct {
	sourceBTCToken      []byte
	erc20TokensMappings []*bridgetypes.ERC20TokenMapping
	currentSequenceTip  math.Int
}

func NewFakeBridgeKeeper(sourceBTCToken []byte) *FakeBridgeKeeper {
	return &FakeBridgeKeeper{
		sourceBTCToken:      sourceBTCToken,
		erc20TokensMappings: make([]*bridgetypes.ERC20TokenMapping, 0),
		currentSequenceTip:  math.NewIntFromBigInt(big.NewInt(0)),
	}
}

func (k *FakeBridgeKeeper) GetSourceBTCToken(_ sdk.Context) []byte {
	return k.sourceBTCToken
}

func (k *FakeBridgeKeeper) BurnBTC(ctx sdk.Context, fromAddr []byte, amount math.Int) error {
	return nil
}

func (k *FakeBridgeKeeper) CreateERC20TokenMapping(
	_ sdk.Context,
	sourceToken, mezoToken []byte,
) error {
	k.erc20TokensMappings = append(k.erc20TokensMappings, &bridgetypes.ERC20TokenMapping{
		SourceToken: common.BytesToAddress(sourceToken).Hex(),
		MezoToken:   common.BytesToAddress(mezoToken).Hex(),
	})

	return nil
}

func (k *FakeBridgeKeeper) DeleteERC20TokenMapping(
	_ sdk.Context,
	sourceToken []byte,
) error {
	k.erc20TokensMappings = slices.DeleteFunc(
		k.erc20TokensMappings,
		func(m *bridgetypes.ERC20TokenMapping) bool {
			return bytes.Equal(m.SourceTokenBytes(), sourceToken)
		},
	)

	return nil
}

func (k *FakeBridgeKeeper) GetERC20TokensMappings(_ sdk.Context) []*bridgetypes.ERC20TokenMapping {
	return k.erc20TokensMappings
}

func (k *FakeBridgeKeeper) GetAssetsLockedSequenceTip(_ sdk.Context) math.Int {
	return k.currentSequenceTip
}

func (k *FakeBridgeKeeper) setAssetsLockedSequenceTip(newValue math.Int) {
	k.currentSequenceTip = newValue
}

func (k *FakeBridgeKeeper) GetERC20TokenMapping(
	_ sdk.Context,
	sourceToken []byte,
) (*bridgetypes.ERC20TokenMapping, bool) {
	index := slices.IndexFunc(
		k.erc20TokensMappings,
		func(m *bridgetypes.ERC20TokenMapping) bool {
			return bytes.Equal(m.SourceTokenBytes(), sourceToken)
		},
	)
	if index == -1 {
		return nil, false
	}

	return k.erc20TokensMappings[index], true
}

func (k *FakeBridgeKeeper) GetParams(_ sdk.Context) bridgetypes.Params {
	return bridgetypes.DefaultParams()
}

func (k *FakeBridgeKeeper) SaveAssetsUnlocked(
	_ sdk.Context,
	_ []byte,
	_ math.Int,
	_ uint8,
	_ []byte,
) (*bridgetypes.AssetsUnlockedEvent, error) {
	return nil, errors.New("unimplemented")
}
