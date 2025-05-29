package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetERC20TokensMappings returns all ERC20 token mappings supported by the bridge.
func (k Keeper) GetERC20TokensMappings(ctx sdk.Context) []*types.ERC20TokenMapping {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.ERC20TokenMappingKeyPrefix)
	defer func() {
		_ = iterator.Close()
	}()

	var mappings []*types.ERC20TokenMapping

	for ; iterator.Valid(); iterator.Next() {
		mapping := types.MustUnmarshalERC20TokenMapping(k.cdc, iterator.Value())
		mappings = append(mappings, &mapping)
	}

	return mappings
}

// GetERC20TokenMapping returns an ERC20 token mapping by the corresponding
// source token address. The boolean return value indicates if the mapping was found.
func (k Keeper) GetERC20TokenMapping(
	ctx sdk.Context,
	sourceToken []byte,
) (*types.ERC20TokenMapping, bool) {
	store := ctx.KVStore(k.storeKey)

	mappingBytes := store.Get(types.GetERC20TokenMappingKey(sourceToken))
	if mappingBytes == nil {
		return nil, false
	}

	mapping := types.MustUnmarshalERC20TokenMapping(k.cdc, mappingBytes)

	return &mapping, true
}

// CreateERC20TokenMapping creates a new ERC20 token mapping.
// Requirements:
// - The source token address must be a valid EVM address and not zero,
// - The Mezo token address must be a valid EVM address and not zero,
// - The source token address must not be already mapped,
// - The maximum number of mappings must not be reached.
func (k Keeper) CreateERC20TokenMapping(
	ctx sdk.Context,
	sourceToken, mezoToken []byte,
) error {
	mapping := types.NewERC20TokenMapping(sourceToken, mezoToken)

	// In the current implementation using types.NewERC20TokenMapping, there is
	// no possibility for the sourceToken to be an invalid hex-encoded EVM address.
	// However, we keep this check for to make this code future-proof. Same for the mezoToken.
	if !evmtypes.IsHexAddress(mapping.SourceToken) {
		return sdkerrors.Wrap(types.ErrInvalidEVMAddress, "invalid source token")
	}

	if evmtypes.IsZeroHexAddress(mapping.SourceToken) {
		return sdkerrors.Wrap(types.ErrZeroEVMAddress, "zero source token")
	}

	if !evmtypes.IsHexAddress(mapping.MezoToken) {
		return sdkerrors.Wrap(types.ErrInvalidEVMAddress, "invalid mezo token")
	}

	if evmtypes.IsZeroHexAddress(mapping.MezoToken) {
		return sdkerrors.Wrap(types.ErrZeroEVMAddress, "zero mezo token")
	}

	if _, exists := k.GetERC20TokenMapping(ctx, mapping.SourceTokenBytes()); exists {
		return types.ErrAlreadyMapping
	}

	existingMappingsCount := uint32(len(k.GetERC20TokensMappings(ctx))) //nolint:gosec
	maxMappingsCount := k.GetParams(ctx).MaxErc20TokensMappings

	if existingMappingsCount >= maxMappingsCount {
		return types.ErrMaxMappingsReached
	}

	if !k.evmKeeper.IsContract(ctx, mapping.MezoTokenBytes()) {
		return sdkerrors.Wrap(types.ErrTokenNotContract, "mezo token")
	}

	k.setERC20TokenMapping(ctx, mapping)

	return nil
}

// DeleteERC20TokenMapping deletes an ERC20 token mapping.
// Requirements:
// - The mapping must exist.
func (k Keeper) DeleteERC20TokenMapping(
	ctx sdk.Context,
	sourceToken []byte,
) error {
	if _, exists := k.GetERC20TokenMapping(ctx, sourceToken); !exists {
		return types.ErrNotMapping
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetERC20TokenMappingKey(sourceToken))

	return nil
}

// setERC20TokensMappings bulk sets the provided ERC20 token mappings.
func (k Keeper) setERC20TokensMappings(
	ctx sdk.Context,
	mappings []*types.ERC20TokenMapping,
) {
	for _, mapping := range mappings {
		k.setERC20TokenMapping(ctx, mapping)
	}
}

// setERC20TokenMapping sets an ERC20 token mapping.
func (k Keeper) setERC20TokenMapping(
	ctx sdk.Context,
	mapping *types.ERC20TokenMapping,
) {
	store := ctx.KVStore(k.storeKey)
	mappingBytes := types.MustMarshalERC20TokenMapping(k.cdc, *mapping)
	store.Set(types.GetERC20TokenMappingKey(mapping.SourceTokenBytes()), mappingBytes)
}
