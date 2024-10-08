package keeper

import (
	"context"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/mock"
)

func mockContext() (sdk.Context, Keeper) {
	logger := log.NewNopLogger()

	keys := storetypes.NewKVStoreKeys(types.StoreKey)

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create the keeper
	keeper := NewKeeper(cdc, keys[types.StoreKey], newMockBankKeeper())

	// Create multiStore in memory
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	// Mount stores
	cms.MountStoreWithDB(keys[types.StoreKey], storetypes.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	// Create context
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, logger)

	return ctx, keeper
}

type mockBankKeeper struct {
	mock.Mock
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{}
}

func (mbk *mockBankKeeper) MintCoins(
	ctx context.Context,
	moduleName string,
	amt sdk.Coins,
) error {
	args := mbk.Called(ctx, moduleName, amt)
	return args.Error(0)
}

func (mbk *mockBankKeeper) SendCoinsFromModuleToAccount(
	ctx context.Context,
	senderModule string,
	recipientAddr sdk.AccAddress,
	amt sdk.Coins,
) error {
	args := mbk.Called(ctx, senderModule, recipientAddr, amt)
	return args.Error(0)
}
