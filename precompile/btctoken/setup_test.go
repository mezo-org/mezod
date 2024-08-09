package btctoken_test

import (
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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
	Priv    cryptotypes.PrivKey
}

type PrecompileTestSuite struct {
	suite.Suite

	app *app.Mezo
	ctx sdk.Context

	account1, account2 Key

	btcTokenPrecompile *precompile.Contract
}

func NewKey() Key {
	addr, privKey := utiltx.NewAddrKey()
	return Key{
		EvmAddr: addr,
		SdkAddr: sdk.AccAddress(addr.Bytes()),
		Priv:    privKey,
	}
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
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
