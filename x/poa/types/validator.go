package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
)

func NewValidator(operator sdk.ValAddress, pubKey cryptotypes.PubKey, description Description) Validator {
	var pkStr string
	if pubKey != nil {
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
		return sdkerrors.Wrap(ErrInvalidValidator, "missing validator address")
	}
	if v.GetConsPubKeyString() == "" {
		return sdkerrors.Wrap(ErrInvalidValidator, "missing consensus pubkey")
	}
	if v.GetDescription() == (Description{}) {
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
func MustMarshalValidator(cdc codec.Codec, validator Validator) []byte {
	return cdc.MustMarshal(&validator)
}
func MustUnmarshalValidator(cdc codec.Codec, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, value)
	if err != nil {
		panic(err)
	}
	return validator
}
func UnmarshalValidator(cdc codec.Codec, value []byte) (v Validator, err error) {
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

// Validator states
const (
	ValidatorStateJoining uint16 = iota // The validator is joining the validator set, it is not yet present in Tendermint validator set
	ValidatorStateJoined  uint16 = iota // The validator is already present in Tendermind validator set
	ValidatorStateLeaving uint16 = iota // The validator is leaving the validator set, it will leave Tendermint validator set at the end of the block
)
