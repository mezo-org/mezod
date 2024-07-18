package keeper

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// Keeper of the poa store
type Keeper struct {
	storeKey          storetypes.StoreKey
	cdc               codec.BinaryCodec
	authority         sdk.AccAddress
	historicalEntries uint32
}

// NewKeeper creates a poa keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority sdk.AccAddress,
) Keeper {
	keeper := Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		authority:         authority,
		historicalEntries: types.DefaultHistoricalEntries,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Authority returns the authority address.
func (k Keeper) Authority() sdk.AccAddress {
	return k.authority
}
