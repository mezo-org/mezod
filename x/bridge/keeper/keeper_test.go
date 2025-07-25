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
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/mock"
)

//nolint:gosec
const testSourceBTCToken = "0x517f2982701695D4E52f1ECFBEf3ba31Df470161"

const testBlockedAddress = "mezo10d07y265gmmuvt4z0w9aw880jnsr700jdl5p9z"

func mockContext() (sdk.Context, Keeper) {
	logger := log.NewNopLogger()

	keys := storetypes.NewKVStoreKeys(types.StoreKey)

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create the keeper
	keeper := NewKeeper(cdc, keys[types.StoreKey], newMockBankKeeper(), newMockEvmKeeper(), map[string]bool{testBlockedAddress: true})

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

	keeper.SetSourceBTCToken(ctx, evmtypes.HexAddressToBytes(testSourceBTCToken))

	err = keeper.SetParams(ctx, types.DefaultParams())
	if err != nil {
		panic(err)
	}

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

func (mbk *mockBankKeeper) BurnCoins(
	ctx context.Context,
	moduleName string,
	amt sdk.Coins,
) error {
	args := mbk.Called(ctx, moduleName, amt)
	return args.Error(0)
}

func (mbk *mockBankKeeper) GetSupply(
	ctx context.Context,
	denom string,
) sdk.Coin {
	args := mbk.Called(ctx, denom)

	var ret sdk.Coin
	if rf, ok := args.Get(0).(func(context.Context, string) sdk.Coin); ok {
		ret = rf(ctx, denom)
	} else {
		ret = args.Get(0).(sdk.Coin)
	}

	return ret
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

func (mbk *mockBankKeeper) SendCoinsFromAccountToModule(
	ctx context.Context,
	senderAddr sdk.AccAddress,
	recipientModule string,
	amt sdk.Coins,
) error {
	args := mbk.Called(ctx, senderAddr, recipientModule, amt)
	return args.Error(0)
}

type mockEvmKeeper struct {
	mock.Mock
}

func newMockEvmKeeper() *mockEvmKeeper {
	return &mockEvmKeeper{}
}

func (mek *mockEvmKeeper) ExecuteContractCall(
	ctx sdk.Context,
	call evmtypes.ContractCall,
) (*evmtypes.MsgEthereumTxResponse, error) {
	args := mek.Called(ctx, call)

	if res := args.Get(0); res != nil {
		return res.(*evmtypes.MsgEthereumTxResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func (mek *mockEvmKeeper) IsContract(ctx sdk.Context, address []byte) bool {
	args := mek.Called(ctx, address)
	return args.Bool(0)
}
