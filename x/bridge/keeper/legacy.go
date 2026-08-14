package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// This file holds the pause surface replaced by the poa emergency team and the
// two lockdown flags. It exists for the v13.0.0 migration and for the assets
// bridge precompile versions 4 and 5, which still register the retired methods.

// GetLegacyPauser returns the current pauser address.
func (k Keeper) GetLegacyPauser(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)

	pauser := store.Get(types.LegacyPauserKey)

	if len(pauser) == 0 {
		pauser = evmtypes.HexAddressToBytes(evmtypes.ZeroHexAddress())
	}

	return pauser
}

// SetLegacyPauser sets the pauser address.
func (k Keeper) SetLegacyPauser(ctx sdk.Context, pauser sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	if len(pauser) == 0 {
		pauser = evmtypes.HexAddressToBytes(evmtypes.ZeroHexAddress())
	}

	store.Set(types.LegacyPauserKey, pauser)
}

// DeleteLegacyPauser removes the pauser address.
func (k Keeper) DeleteLegacyPauser(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete(types.LegacyPauserKey)
}

// LegacyPauseBridgeOut sets the outflow limit to 0 for all supported tokens.
func (k Keeper) LegacyPauseBridgeOut(ctx sdk.Context, caller sdk.AccAddress) error {
	pauser := k.GetLegacyPauser(ctx)
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

// IsLegacyTripartyPaused checks if triparty bridging is paused.
func (k Keeper) IsLegacyTripartyPaused(ctx sdk.Context) bool {
	return ctx.KVStore(k.storeKey).Has(types.LegacyTripartyPausedKey)
}

// SetLegacyTripartyPaused sets or removes the triparty paused flag.
func (k Keeper) SetLegacyTripartyPaused(ctx sdk.Context, isPaused bool) {
	store := ctx.KVStore(k.storeKey)

	if isPaused {
		store.Set(types.LegacyTripartyPausedKey, []byte{0x01})
	} else {
		store.Delete(types.LegacyTripartyPausedKey)
	}
}

// DeleteLegacyTripartyPaused removes the triparty paused flag.
func (k Keeper) DeleteLegacyTripartyPaused(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete(types.LegacyTripartyPausedKey)
}
