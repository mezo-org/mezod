package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// IsBridgeOutChainEnabled checks if the given target chain accepts bridge-outs.
// An unknown chain is disabled, so a network that never seeded the set blocks
// every bridge-out.
func (k Keeper) IsBridgeOutChainEnabled(ctx sdk.Context, chain uint8) bool {
	return ctx.KVStore(k.storeKey).Has(types.GetBridgeOutChainKey(chain))
}

// EnableBridgeOutChain lets the given target chain accept bridge-outs. Enabling
// a chain that is already enabled changes nothing.
func (k Keeper) EnableBridgeOutChain(ctx sdk.Context, chain uint8) {
	ctx.KVStore(k.storeKey).Set(types.GetBridgeOutChainKey(chain), []byte{0x01})
}

// DisableBridgeOutChain stops the given target chain from accepting
// bridge-outs. Disabling a chain that is already disabled changes nothing.
func (k Keeper) DisableBridgeOutChain(ctx sdk.Context, chain uint8) {
	ctx.KVStore(k.storeKey).Delete(types.GetBridgeOutChainKey(chain))
}

// GetBridgeOutChains returns all target chains that accept bridge-outs, in
// ascending order.
func (k Keeper) GetBridgeOutChains(ctx sdk.Context) []uint8 {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(
		store, types.BridgeOutChainKeyPrefix,
	)
	defer func() {
		_ = iterator.Close()
	}()

	var out []uint8

	for ; iterator.Valid(); iterator.Next() {
		chain := iterator.Key()[len(types.BridgeOutChainKeyPrefix):]
		out = append(out, chain[0])
	}

	return out
}
