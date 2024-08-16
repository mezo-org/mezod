package keeper

import (
	"cosmossdk.io/store/metrics"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
)

func mockContext() (sdk.Context, Keeper) {
	logger := log.NewNopLogger()

	keys := storetypes.NewKVStoreKeys(types.StoreKey)

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create a poa keeper
	poaKeeper := NewKeeper(keys[types.StoreKey], cdc)

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

	return ctx, poaKeeper
}

func mockValidator() (types.Validator, string) {
	// Junk description
	validatorDescription := types.Description{
		Moniker:         "Moniker",
		Identity:        "Identity",
		Website:         "Website",
		SecurityContact: "SecurityContact",
		Details:         "Details",
	}

	// Generate operator address
	tmpk := ed25519.GenPrivKey().PubKey()
	addr := tmpk.Address()
	operatorAddress := sdk.ValAddress(addr)

	// Generate a consPubKey
	tmpk = ed25519.GenPrivKey().PubKey()
	pk, err := cryptocdc.FromCmtPubKeyInterface(tmpk)
	if err != nil {
		panic(err)
	}
	//nolint:staticcheck
	consPubKey := legacybech32.MustMarshalPubKey(legacybech32.ConsPK, pk)

	validator := types.Validator{
		OperatorBech32:   operatorAddress.String(),
		ConsPubKeyBech32: consPubKey,
		Description:      validatorDescription,
	}

	return validator, consPubKey
}
