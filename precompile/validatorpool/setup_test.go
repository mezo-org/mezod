package validatorpool_test

import (
	"math/rand"
	"slices"
	"strings"
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
	"github.com/mezo-org/mezod/precompile/validatorpool"
	"github.com/mezo-org/mezod/testutil"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/mezo-org/mezod/x/evm/statedb"
	poatypes "github.com/mezo-org/mezod/x/poa/types"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/maps"
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
	EvmAddr     common.Address
	SdkAddr     sdk.AccAddress
	ConsPubKey  cryptotypes.PubKey
	Priv        cryptotypes.PrivKey
	Validator   poatypes.Validator
	Description validatorpool.Description
}

func (k *Key) ConsPubKeyBytes32() [32]byte {
	var consPubKey [32]byte
	copy(consPubKey[:], k.ConsPubKey.Bytes())
	return consPubKey
}

type PrecompileTestSuite struct {
	suite.Suite

	app    *app.Mezo
	keeper *FakePoaKeeper
	ctx    sdk.Context

	account1, account2, account3, account4, account5 Key

	validatorpoolPrecompile *precompile.Contract
}

func NewKey(withValidator bool) Key {
	addr, privKey := utiltx.NewAddrKey()
	// Generate a consPubKey
	tmpk := ed25519.GenPrivKey().PubKey()
	consPubKey, err := cryptocdc.FromCmtPubKeyInterface(tmpk)
	if err != nil {
		panic(err)
	}

	sdkAddr := sdk.AccAddress(addr.Bytes())
	validator := poatypes.Validator{}
	description := validatorpool.Description{}

	if withValidator {
		// Create a validator description
		descID := addr.String()[0:8]
		description = validatorpool.Description{
			Moniker:         "moniker-" + descID,
			Identity:        "identity-" + descID,
			Website:         "website-" + descID,
			SecurityContact: "securityContact-" + descID,
			Details:         "details-" + descID,
		}

		// Create a validator
		validator, err = poatypes.NewValidator(sdk.ValAddress(sdkAddr), consPubKey, poatypes.Description(description))
		if err != nil {
			panic(err)
		}
	}

	return Key{
		EvmAddr:     addr,
		SdkAddr:     sdkAddr,
		Priv:        privKey,
		Validator:   validator,
		ConsPubKey:  consPubKey,
		Description: description,
	}
}

// returns a random string of length l
func randomString(l int) string {
	characters := []rune("ABCDEF0123456789")
	var sb strings.Builder

	for i := 0; i < l; i++ {
		randomIndex := rand.Intn(len(characters)) //nolint:all
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	// accounts
	s.account1 = NewKey(true)  // owner account
	s.account2 = NewKey(true)  // applicant account
	s.account3 = NewKey(true)  // validator account
	s.account4 = NewKey(true)  // candidateOwner account
	s.account5 = NewKey(false) // invalid validator applicant (description exceeds character)

	// set a description on account 5 that exceeds the character limit
	s.account5.Description.Moniker = randomString(101)

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	// init fake keeper
	s.keeper = NewFakePoaKeeper(
		s.account1.SdkAddr,
		s.account4.SdkAddr,
		poatypes.NewApplication(s.account2.Validator),
		s.account3.Validator,
	)

	err = s.keeper.AddPrivilege(
		s.ctx,
		s.account1.SdkAddr,
		[]sdk.ValAddress{sdk.ValAddress(s.account4.SdkAddr)},
		"bridge",
	)
	s.Require().NoError(err)

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
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}
			validatorpoolPrecompile, err := validatorpool.NewPrecompile(s.keeper)
			s.Require().NoError(err)
			s.validatorpoolPrecompile = validatorpoolPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.validatorpoolPrecompile.Abi.Methods[methodName]
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

			output, err := s.validatorpoolPrecompile.Run(evm, vmContract, false)
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
	owner          sdk.AccAddress
	candidateOwner sdk.AccAddress
	applications   map[string]poatypes.Application
	validators     map[string]poatypes.Validator
	privileges     map[string][]sdk.ValAddress
}

func NewFakePoaKeeper(
	owner sdk.AccAddress,
	candidateOwner sdk.AccAddress,
	application poatypes.Application,
	validator poatypes.Validator,
) *FakePoaKeeper {
	applications := make(map[string]poatypes.Application)
	applications[application.Validator.GetOperator().String()] = application

	validators := make(map[string]poatypes.Validator)
	validators[validator.GetOperator().String()] = validator
	return &FakePoaKeeper{
		owner:          owner,
		candidateOwner: candidateOwner,
		applications:   applications,
		validators:     validators,
		privileges:     make(map[string][]sdk.ValAddress),
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

	if !sender.Equals(k.owner) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not owner",
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

	if !sender.Equals(k.candidateOwner) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not candidateOwner",
		)
	}
	k.owner = k.candidateOwner
	k.candidateOwner = sdk.AccAddress{}
	return nil
}

func (k *FakePoaKeeper) SubmitApplication(ctx sdk.Context, _ sdk.AccAddress, validator poatypes.Validator) error {
	_, found := k.GetApplication(ctx, validator.GetOperator())
	if found == true {
		return poatypes.ErrAlreadyApplying
	}
	application := poatypes.NewApplication(validator)
	k.applications[validator.GetOperator().String()] = application
	return nil
}

func (k *FakePoaKeeper) ApproveApplication(ctx sdk.Context, sender sdk.AccAddress, operator sdk.ValAddress) error {
	if !sender.Equals(k.owner) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not owner",
		)
	}

	application, found := k.GetApplication(ctx, operator)
	if !found {
		return errorsmod.Wrap(
			sdkerrors.ErrNotFound,
			"application does not exist",
		)
	}

	validator := application.GetValidator()

	k.validators[operator.String()] = validator
	delete(k.applications, operator.String())
	return nil
}

func (k *FakePoaKeeper) Leave(ctx sdk.Context, sender sdk.AccAddress) error {
	_, found := k.GetValidator(ctx, sdk.ValAddress(sender))
	if !found {
		return errorsmod.Wrap(
			poatypes.ErrWrongValidatorState,
			"not an active validator",
		)
	}
	delete(k.validators, sdk.ValAddress(sender).String())
	return nil
}

func (k *FakePoaKeeper) Kick(_ sdk.Context, sender sdk.AccAddress, operator sdk.ValAddress) error {
	if !sender.Equals(k.owner) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not owner",
		)
	}

	delete(k.validators, operator.String())
	return nil
}

func (k *FakePoaKeeper) GetApplication(_ sdk.Context, operator sdk.ValAddress) (poatypes.Application, bool) {
	application, found := k.applications[operator.String()]
	return application, found
}

func (k *FakePoaKeeper) GetAllApplications(sdk.Context) []poatypes.Application {
	return maps.Values(k.applications)
}

func (k *FakePoaKeeper) GetValidator(_ sdk.Context, operator sdk.ValAddress) (poatypes.Validator, bool) {
	for _, validator := range k.validators {
		if operator.Equals(validator.GetOperator()) {
			return validator, true
		}
	}
	return poatypes.Validator{}, false
}

func (k *FakePoaKeeper) GetAllValidators(sdk.Context) []poatypes.Validator {
	return maps.Values(k.validators)
}

func (k *FakePoaKeeper) AddPrivilege(
	_ sdk.Context,
	sender sdk.AccAddress,
	operators []sdk.ValAddress,
	privilege string,
) error {
	if sender.Empty() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"sender address is empty",
		)
	}

	if !sender.Equals(k.owner) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not owner",
		)
	}

	k.privileges[privilege] = operators

	return nil
}

func (k *FakePoaKeeper) RemovePrivilege(
	_ sdk.Context,
	sender sdk.AccAddress,
	operators []sdk.ValAddress,
	privilege string,
) error {
	if sender.Empty() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"sender address is empty",
		)
	}

	if !sender.Equals(k.owner) {
		return errorsmod.Wrap(
			sdkerrors.ErrUnauthorized,
			"sender is not owner",
		)
	}

	existing := k.privileges[privilege]

	for _, operator := range operators {
		index := slices.IndexFunc(
			existing,
			func(o sdk.ValAddress) bool {
				return o.Equals(operator)
			},
		)
		if index >= 0 {
			existing = append(existing[:index], existing[index+1:]...)
		}
	}

	k.privileges[privilege] = existing

	return nil
}

func (k *FakePoaKeeper) GetValidatorsOperatorsByPrivilege(
	_ sdk.Context,
	privilege string,
) []sdk.ValAddress {
	return k.privileges[privilege]
}
