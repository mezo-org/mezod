package btctoken_test

import (
	"crypto/ecdsa"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/btctoken"
	"github.com/mezo-org/mezod/testutil"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/stretchr/testify/suite"
)

// keccak256(encode(
//
//			keccak256(
//				"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
//			),
//			keccak256(BTC),
//			keccak256(1),
//			31612,
//			0x7b7C000000000000000000000000000000000000
//		)
//	)
//
// Same in hex: f98315225d67e6d98b18f0a6b73bf711423ff8310bd350400db36e876c5cddf4
var DomainSeparator = []byte{
	249, 131, 21, 34, 93, 103, 230, 217, 139, 24, 240, 166, 183, 59, 247, 17,
	66, 63, 248, 49, 11, 211, 80, 64, 13, 179, 110, 135, 108, 92, 221, 244,
}

type Key struct {
	EvmAddr common.Address
	SdkAddr sdk.AccAddress
	Priv    *ecdsa.PrivateKey
}

type PrecompileTestSuite struct {
	suite.Suite

	app *app.Mezo
	ctx sdk.Context

	account1, account2 Key

	denom               string
	name                string
	symbol              string
	decimals            uint8
	domainSeparator     []byte
	precompileFactoryFn func(*app.Mezo) (*precompile.Contract, error)

	erc20Precompile *precompile.Contract
}

func NewKey() Key {
	addr, privKey := utiltx.NewAddrKey()
	privKeyECDSA, _ := privKey.ToECDSA()
	return Key{
		EvmAddr: addr,
		SdkAddr: sdk.AccAddress(addr.Bytes()),
		Priv:    privKeyECDSA,
	}
}

func TestPrecompileTestSuite(t *testing.T) {
	suiteInstance := new(PrecompileTestSuite)

	suiteInstance.denom = "abtc"
	suiteInstance.name = "BTC"
	suiteInstance.symbol = "BTC"
	suiteInstance.decimals = uint8(18)
	suiteInstance.domainSeparator = DomainSeparator

	suiteInstance.precompileFactoryFn = func(app *app.Mezo) (*precompile.Contract, error) {
		return btctoken.NewPrecompile(app.BankKeeper, app.AuthzKeeper, *app.EvmKeeper, "mezo_31612-1")
	}

	suite.Run(t, suiteInstance)
}

func (s *PrecompileTestSuite) SetupTest() {
	// accounts
	s.account1 = NewKey()
	s.account2 = NewKey()

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	// init app
	s.app = app.Setup(false, nil)
	header := testutil.NewHeader(
		1, time.Now().UTC(), "mezo_31612-1", consAddress, nil, nil,
	)
	s.ctx = s.app.BaseApp.NewContextLegacy(false, header)
}
