package mezotoken_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/precompile"
	erc20testsuite "github.com/mezo-org/mezod/precompile/erc20/testsuite"
	"github.com/mezo-org/mezod/precompile/mezotoken"
	"github.com/mezo-org/mezod/testutil"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/stretchr/testify/suite"
)

const (
	Denom    = "amezo"
	Name     = "MEZO"
	Symbol   = "MEZO"
	Decimals = uint8(18)
)

// This domain separator was computed using the following Solidity code:
//
//	keccak256(
//		abi.encode(
//			keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
//			keccak256("MEZO"),
//			keccak256("1"),
//			31612,
//			0x7B7c000000000000000000000000000000000001
//		)
//	);
//
// Same in hex: 43d4db585ef58130660466546b06e8802f4479372409cf0636169cfd80f6a186
var DomainSeparator = []byte{
	0x43, 0xd4, 0xdb, 0x58, 0x5e, 0xf5, 0x81, 0x30, 0x66, 0x04, 0x66, 0x54, 0x6b, 0x06, 0xe8, 0x80,
	0x2f, 0x44, 0x79, 0x37, 0x24, 0x09, 0xcf, 0x06, 0x36, 0x16, 0x9c, 0xfd, 0x80, 0xf6, 0xa1, 0x86,
}

type PrecompileTestSuite struct {
	suite.Suite

	app                 *app.Mezo
	ctx                 sdk.Context
	mezoPrecompile      *precompile.Contract
	poaOwner            common.Address
	poaOwnerSDK         sdk.AccAddress
	minter              common.Address
	minterSDK           sdk.AccAddress
	recipient           common.Address
	recipientSDK        sdk.AccAddress
	unauthorizedAddr    common.Address
	unauthorizedAddrSDK sdk.AccAddress
}

func TestMEZOPrecompile(t *testing.T) {
	precompileFactoryFn := func(app *app.Mezo) (*precompile.Contract, error) {
		return mezotoken.NewPrecompile(
			app.BankKeeper,
			app.AuthzKeeper,
			*app.EvmKeeper,
			app.PoaKeeper,
			"mezo_31612-1",
			&mezotoken.Settings{
				Minting: false,
			},
		)
	}

	erc20TestSuite := erc20testsuite.New(
		Denom,
		Name,
		Symbol,
		Decimals,
		DomainSeparator,
		precompileFactoryFn,
		false,
	)

	// Run the test suite for the common ERC20 functionality
	suite.Run(t, erc20TestSuite)
	// Run the test suite for custom MEZO functionality
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	// Consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	// Init app
	s.app = app.Setup(false, nil)

	header := testutil.NewHeader(
		1, time.Now().UTC(), "mezo_31612-1", consAddress, nil, nil,
	)
	s.ctx = s.app.BaseApp.NewContextLegacy(false, header)

	// Get POA owner from genesis (set by app.Setup)
	s.poaOwnerSDK = s.app.PoaKeeper.GetOwner(s.ctx)
	s.poaOwner = common.BytesToAddress(s.poaOwnerSDK.Bytes())

	// Create minter account
	minterAddr, _ := utiltx.NewAddrKey()
	s.minter = minterAddr
	s.minterSDK = sdk.AccAddress(minterAddr.Bytes())

	// Create recipient account
	recipientAddr, _ := utiltx.NewAddrKey()
	s.recipient = recipientAddr
	s.recipientSDK = sdk.AccAddress(recipientAddr.Bytes())

	// Create unauthorized account
	unauthorizedAddr, _ := utiltx.NewAddrKey()
	s.unauthorizedAddr = unauthorizedAddr
	s.unauthorizedAddrSDK = sdk.AccAddress(unauthorizedAddr.Bytes())

	// Create precompile
	s.mezoPrecompile, err = mezotoken.NewPrecompile(
		s.app.BankKeeper,
		s.app.AuthzKeeper,
		*s.app.EvmKeeper,
		s.app.PoaKeeper,
		"mezo_31612-1",
		&mezotoken.Settings{
			Minting: true,
		},
	)
	s.Require().NoError(err)
}
