package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/evmos/evmos/v12/testutil/keeper"
	"github.com/evmos/evmos/v12/x/dualstaking/keeper"
	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

//nolint:unused //remove if not used in the future
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.DualstakingKeeper(t)
	return keeper.NewMsgServer(*k), sdk.WrapSDKContext(ctx)
}
