package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
)

// ValidatorState is the state of a validator.
type ValidatorState uint8

const (
	// ValidatorStateUnknown is the default state of a validator.
	ValidatorStateUnknown ValidatorState = iota
	// ValidatorStateJoining means that the validator is not yet present in the
	// Tendermint consensus validator set and will join it at the end of the block.
	ValidatorStateJoining
	// ValidatorStateJoined means that the validator is present in the
	// Tendermint consensus validator set.
	ValidatorStateJoined
	// ValidatorStateLeaving means that the validator will leave the Tendermint
	// consensus validator set at the end of the block.
	ValidatorStateLeaving
)

func NewValidator(operator sdk.ValAddress, pubKey cryptotypes.PubKey, description Description) Validator {
	var pkStr string
	if pubKey != nil {
		//nolint:staticcheck
		pkStr = legacybech32.MustMarshalPubKey(legacybech32.ConsPK, pubKey)
	}

	return Validator{
		OperatorAddress: operator,
		ConsensusPubkey: pkStr,
		Description:     description,
	}
}

// Accessors
func (v Validator) GetOperator() sdk.ValAddress {
	return v.OperatorAddress
}

func (v Validator) GetConsPubKeyString() string {
	return v.ConsensusPubkey
}

func (v Validator) GetConsPubKey() cryptotypes.PubKey {
	//nolint:staticcheck
	pubKey, err := legacybech32.UnmarshalPubKey(
		legacybech32.ConsPK,
		v.ConsensusPubkey,
	)
	if err != nil {
		panic(err)
	}

	return pubKey
}

func (v Validator) GetConsAddr() sdk.ConsAddress {
	return sdk.ConsAddress(v.GetConsPubKey().Address())
}

func (v Validator) CheckValid() error {
	if v.GetOperator().Empty() {
		//nolint:staticcheck
		return sdkerrors.Wrap(ErrInvalidValidator, "missing validator address")
	}
	if v.GetConsPubKeyString() == "" {
		//nolint:staticcheck
		return sdkerrors.Wrap(ErrInvalidValidator, "missing consensus pubkey")
	}
	if v.GetDescription() == (Description{}) {
		//nolint:staticcheck
		return sdkerrors.Wrap(ErrInvalidValidator, "empty description")
	}
	return nil
}

// Get a ABCI validator update object from the validator
func (v Validator) ABCIValidatorUpdateAppend() abci.ValidatorUpdate {
	pubKey, err := cryptocdc.ToTmProtoPublicKey(v.GetConsPubKey())
	if err != nil {
		panic(err)
	}

	return abci.ValidatorUpdate{
		PubKey: pubKey,
		Power:  1, // Same weight for all the validators
	}
}

// Get a ABCI validator update with no voting power from the validator
func (v Validator) ABCIValidatorUpdateRemove() abci.ValidatorUpdate {
	pubKey, err := cryptocdc.ToTmProtoPublicKey(v.GetConsPubKey())
	if err != nil {
		panic(err)
	}

	return abci.ValidatorUpdate{
		PubKey: pubKey,
		Power:  0,
	}
}

// Validator encoding functions
func MustMarshalValidator(cdc codec.BinaryCodec, validator Validator) []byte {
	return cdc.MustMarshal(&validator)
}

func MustUnmarshalValidator(cdc codec.BinaryCodec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}

func UnmarshalValidator(cdc codec.BinaryCodec, value []byte) (v Validator, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}

// Create a new Description
func NewDescription(moniker, identity, website, securityContact, details string) Description {
	return Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}
