package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/evmos/evmos/v12/testutil/keeper"
	"github.com/evmos/evmos/v12/x/bridge/keeper"
	"github.com/evmos/evmos/v12/x/bridge/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.BridgeKeeper(t)
	return keeper.NewMsgServer(*k), sdk.WrapSDKContext(ctx)
}
