package priceoracle_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/mezo-org/mezod/precompile/priceoracle"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

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

	oracleQueryServer *FakeOracleQueryServer
	app               *app.Mezo
	ctx               sdk.Context

	account1, account2 Key

	priceOraclePrecompile *precompile.Contract
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

	s.oracleQueryServer = NewFakeOracleQueryServer()

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

			priceOraclePrecompile, err := priceoracle.NewPrecompile(s.oracleQueryServer)
			s.Require().NoError(err)
			s.priceOraclePrecompile = priceOraclePrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.priceOraclePrecompile.Abi.Methods[methodName]
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

			output, err := s.priceOraclePrecompile.Run(evm, vmContract, false)
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

type FakeOracleQueryServer struct {
	price          sdkmath.Int
	blockTimestamp time.Time
	nonce          uint64
	decimals       uint64
	id             uint64
}

func NewFakeOracleQueryServer() *FakeOracleQueryServer {
	return &FakeOracleQueryServer{}
}

func (k *FakeOracleQueryServer) GetPrice(
	_ context.Context,
	req *oracletypes.GetPriceRequest,
) (*oracletypes.GetPriceResponse, error) {
	if req.CurrencyPair != "BTC/USD" {
		return nil, fmt.Errorf("invalid currency pair")
	}

	return &oracletypes.GetPriceResponse{
		Price: &oracletypes.QuotePrice{
			Price:          k.price,
			BlockTimestamp: k.blockTimestamp,
			BlockHeight:    100,
		},
		Nonce:    k.nonce,
		Decimals: k.decimals,
		Id:       k.id,
	}, nil
}

func (k *FakeOracleQueryServer) SetPrice(
	price sdkmath.Int,
	blockTimestamp time.Time,
	nonce uint64,
	decimals uint64,
	id uint64,
) {
	k.price = price
	k.blockTimestamp = blockTimestamp
	k.nonce = nonce
	k.decimals = decimals
	k.id = id
}
