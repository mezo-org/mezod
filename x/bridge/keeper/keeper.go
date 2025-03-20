package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/mezo-org/mezod/x/bridge/types"
)

type Keeper struct {
	cdc          codec.Codec
	storeKey     storetypes.StoreKey
	bankKeeper   types.BankKeeper
	evmKeeper    types.EvmKeeper
	blockedAddrs map[string]bool
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	evmKeeper types.EvmKeeper,
	blockAddrs map[string]bool,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		bankKeeper:   bankKeeper,
		evmKeeper:    evmKeeper,
		blockedAddrs: blockAddrs,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
