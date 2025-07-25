package types

import (
	"context"

	"github.com/mezo-org/mezod/x/evm/types"

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

	// BurnCoins burns coin from the module account.
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	// GetSupply retrieves the Supply from store
	GetSupply(ctx context.Context, denom string) sdk.Coin

	// SendCoinsFromModuleToAccount sends coins from the module account to the
	// recipient account.
	SendCoinsFromModuleToAccount(
		ctx context.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error

	// SendCoinsFromAccountToModule sends coins from the sender account to the
	// module account.
	SendCoinsFromAccountToModule(
		ctx context.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error
}

// EvmKeeper is an interface to the x/evm module keeper.
type EvmKeeper interface {
	// ExecuteContractCall executes an EVM contract call.
	ExecuteContractCall(
		ctx sdk.Context,
		call types.ContractCall,
	) (*types.MsgEthereumTxResponse, error)

	// IsContract returns if the account contains contract code.
	IsContract(ctx sdk.Context, address []byte) bool
}

// AccountKeeper is an interface to the x/auth module keeper.
type AccountKeeper interface {
	// GetModuleAccount gets the module account from the auth account store, if the account does not
	// exist in the AccountKeeper, then it is created.
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}
