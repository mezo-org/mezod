package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// ValidatorStateActive means that the validator is present in the
	// Tendermint consensus validator set.
	ValidatorStateActive
	// ValidatorStateLeaving means that the validator will leave the Tendermint
	// consensus validator set at the end of the block.
	ValidatorStateLeaving
)

// NewValidator creates a new validator. Validates inputs and returns an error
// if any of them is invalid.
func NewValidator(
	operator sdk.ValAddress,
	consPubKey cryptotypes.PubKey,
	description Description,
) (Validator, error) {
	if operator.Empty() {
		return Validator{}, errorsmod.Wrap(
			ErrInvalidValidator,
			"missing operator address",
		)
	}

	if consPubKey == nil {
		return Validator{}, errorsmod.Wrap(
			ErrInvalidValidator,
			"missing consensus public key",
		)
	}

	consPubKeyBech32, err := legacybech32.MarshalPubKey(
		legacybech32.ConsPK,
		consPubKey,
	)
	if err != nil {
		return Validator{}, errorsmod.Wrapf(
			ErrInvalidValidator,
			"cannot marshal consensus public key: %v",
			err,
		)
	}

	return Validator{
		// This converts the ValAddress to a bech32 string.
		OperatorBech32:   operator.String(),
		ConsPubKeyBech32: consPubKeyBech32,
		Description:      description,
	}, nil
}

// GetOperator gets the operator address of the validator.
func (v Validator) GetOperator() sdk.ValAddress {
	operator, err := sdk.ValAddressFromBech32(v.OperatorBech32)
	if err != nil {
		// Should never happen. The address is validated when the validator is created.
		panic(err)
	}
	return operator
}

// GetConsPubKey gets the consensus public key of the validator.
func (v Validator) GetConsPubKey() cryptotypes.PubKey {
	//nolint:staticcheck
	pubKey, err := legacybech32.UnmarshalPubKey(
		legacybech32.ConsPK,
		v.ConsPubKeyBech32,
	)
	if err != nil {
		// Should never happen. The public key is validated when the validator is created.
		panic(err)
	}

	return pubKey
}

// GetConsAddress gets the consensus address of the validator.
func (v Validator) GetConsAddress() sdk.ConsAddress {
	return sdk.ConsAddress(v.GetConsPubKey().Address())
}

// ABCIValidatorUpdateAppend gets an ABCI validator update object from the validator.
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

// ABCIValidatorUpdateRemove gets a ABCI validator update with no voting power
// from the validator.
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

// MustMarshalValidator marshals a validator to bytes. It panics on error.
func MustMarshalValidator(cdc codec.BinaryCodec, validator Validator) []byte {
	return cdc.MustMarshal(&validator)
}

// MustUnmarshalValidator unmarshals a validator from bytes. It panics on error.
func MustUnmarshalValidator(cdc codec.BinaryCodec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// UnmarshalValidator unmarshals a validator from bytes.
func UnmarshalValidator(cdc codec.BinaryCodec, value []byte) (v Validator, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}

// NewDescription creates a new description.
func NewDescription(moniker, identity, website, securityContact, details string) Description {
	return Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}
