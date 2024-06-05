package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/evmos/evmos/v12/x/dualstaking/types"
    "github.com/evmos/evmos/v12/x/dualstaking/keeper"
    keepertest "github.com/evmos/evmos/v12/testutil/keeper"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.DualstakingKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
