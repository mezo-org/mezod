package types

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidatorPrivilege is the privilege allowing a validator to participate
// in the bridge.
const ValidatorPrivilege = "bridge"

// ValidatorStore is an interface to the validator store.
type ValidatorStore interface {
	baseapp.ValidatorStore

	// GetValidatorsConsAddrsByPrivilege returns the consensus addresses of
	// all validators that are currently present in the store and have the
	// given privilege. There is no guarantee that the returned validators
	// are currently part of the CometBFT validator set.
	GetValidatorsConsAddrsByPrivilege(privilege string) []sdk.ConsAddress
}
