package testsuite

import (
	"crypto/ecdsa"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/testutil"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/stretchr/testify/suite"
)

type Key struct {
	EvmAddr common.Address
	SdkAddr sdk.AccAddress
	Priv    *ecdsa.PrivateKey
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

type TestSuite struct {
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

	// Must be true only for the precompile tied to the EVM native gas token.
	ensureEVMBalanceChange bool

	erc20Precompile *precompile.Contract
}

func New(
	denom, name, symbol string,
	decimals uint8,
	domainSeparator []byte,
	precompileFactoryFn func(*app.Mezo) (*precompile.Contract, error),
	ensureEVMBalanceChange bool,
) *TestSuite {
	return &TestSuite{
		denom:                  denom,
		name:                   name,
		symbol:                 symbol,
		decimals:               decimals,
		domainSeparator:        domainSeparator,
		precompileFactoryFn:    precompileFactoryFn,
		ensureEVMBalanceChange: ensureEVMBalanceChange,
	}
}

func (s *TestSuite) SetupTest() {
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
