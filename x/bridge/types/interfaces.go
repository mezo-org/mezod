package types

import (
	"context"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidatorPrivilege represents the privilege of bridge validators that
// have authority to attest bridging. The non-bridge validators validate
// what bridge validators attested.
const ValidatorPrivilege = "bridge"

// ValidatorStore is an interface to the validator store.
type ValidatorStore interface {
	baseapp.ValidatorStore

	// GetValidatorsConsAddrsByPrivilege returns the consensus addresses of
	// all validators that are currently present in the store and have the
	// given privilege. There is no guarantee that the returned validators
	// are currently part of the CometBFT validator set.
	GetValidatorsConsAddrsByPrivilege(
		ctx sdk.Context,
		privilege string,
	) []sdk.ConsAddress
}

// BankKeeper is an interface to the x/bank module keeper.
type BankKeeper interface {
	// MintCoins creates new coins and adds them to the module account.
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	// SendCoinsFromModuleToAccount sends coins from the module account to the
	// recipient account.
	SendCoinsFromModuleToAccount(
		ctx context.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error
}
