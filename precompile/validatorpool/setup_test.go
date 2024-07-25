package validatorpool

import (
	"testing"
	"time"

	errorsmod "cosmossdk.io/errors"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/app"
	"github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/testutil"
	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type Key struct {
	EvmAddr    common.Address
	SdkAddr    sdk.AccAddress
	ConsPubKey cryptotypes.PubKey
	Priv       cryptotypes.PrivKey
}

type PrecompileTestSuite struct {
	suite.Suite

	app    *app.Evmos
	keeper *FakePoaKeeper
	ctx    sdk.Context

	account1, account2 Key

	validatorpoolPrecompile *precompile.Contract
}

func NewKey() Key {
	addr, privKey := utiltx.NewAddrKey()
	// Generate a consPubKey
	tmpk := ed25519.GenPrivKey().PubKey()
	consPubKey, err := cryptocdc.FromTmPubKeyInterface(tmpk)
	if err != nil {
		panic(err)
	}
	return Key{
		EvmAddr:    addr,
		SdkAddr:    sdk.AccAddress(addr.Bytes()),
		Priv:       privKey,
		ConsPubKey: consPubKey,
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

	// init fake keeper
	s.keeper = NewFakePoaKeeper((s.account1.SdkAddr))

	// init app
	s.app = app.Setup(false, nil)
	header := testutil.NewHeader(
		1, time.Now().UTC(), "mezo_31612-1", consAddress, nil, nil,
	)
	s.ctx = s.app.BaseApp.NewContext(false, header)
}

type FakePoaKeeper struct {
	owner          sdk.AccAddress
	candidateOwner sdk.AccAddress
}

func NewFakePoaKeeper(owner sdk.AccAddress) *FakePoaKeeper {
	return &FakePoaKeeper{
		owner: owner,
	}
}

func (k *FakePoaKeeper) GetOwner(sdk.Context) sdk.AccAddress {
	return k.owner
}

func (k *FakePoaKeeper) GetCandidateOwner(sdk.Context) sdk.AccAddress {
	return k.candidateOwner
}

func (k *FakePoaKeeper) TransferOwnership(_ sdk.Context, sender sdk.AccAddress, newOwner sdk.AccAddress) error {
	if sender.Empty() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"sender address is empty",
		)
	}

	k.candidateOwner = newOwner
	return nil
}

func (k *FakePoaKeeper) AcceptOwnership(_ sdk.Context, sender sdk.AccAddress) error {
	if sender.Empty() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"sender address is empty",
		)
	}
	k.owner = k.candidateOwner
	k.candidateOwner = sdk.AccAddress{}
	return nil
}

func (k *FakePoaKeeper) SubmitApplication(sdk.Context, sdk.AccAddress, poatypes.Validator) error {
	return nil
}

func (k *FakePoaKeeper) ApproveApplication(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}

func (k *FakePoaKeeper) Leave(sdk.Context, sdk.AccAddress) error {
	return nil
}

func (k *FakePoaKeeper) Kick(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}
