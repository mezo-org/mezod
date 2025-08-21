package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetPauser returns the current pauser address.
func (k Keeper) GetPauser(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)

	pauser := store.Get(types.PauserKey)

	if len(pauser) == 0 {
		pauser = evmtypes.HexAddressToBytes(evmtypes.ZeroHexAddress())
	}

	return pauser
}

// SetPauser sets the pauser address.
func (k Keeper) SetPauser(ctx sdk.Context, pauser sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	if len(pauser) == 0 {
		pauser = evmtypes.HexAddressToBytes(evmtypes.ZeroHexAddress())
	}

	store.Set(types.PauserKey, pauser)
}

// PauseBridgeOut sets the outflow limit to 0 for all supported tokens.
func (k Keeper) PauseBridgeOut(ctx sdk.Context, caller sdk.AccAddress) error {
	pauser := k.GetPauser(ctx)
	if evmtypes.IsZeroHexAddress(evmtypes.BytesToHexAddress(pauser)) {
		return fmt.Errorf("no pauser is set")
	}

	if !pauser.Equals(caller) {
		return fmt.Errorf("caller is not the pauser")
	}

	// Set outflow limit to 0 for BTC token
	btcToken := evmtypes.HexAddressToBytes(evmtypes.BTCTokenPrecompileAddress)
	k.SetOutflowLimit(ctx, btcToken, math.ZeroInt())

	// Set outflow limit to 0 for all ERC20 tokens (using mezo token addresses)
	mappings := k.GetERC20TokensMappings(ctx)
	for _, mapping := range mappings {
		k.SetOutflowLimit(ctx, evmtypes.HexAddressToBytes(mapping.MezoToken), math.ZeroInt())
	}

	return nil
}
