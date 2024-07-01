package keeper

import (
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func mockContext() (sdk.Context, Keeper) {
	keys := sdk.NewKVStoreKeys(types.StoreKey)

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// Create a poa keeper
	poaKeeper := NewKeeper(keys[types.StoreKey], cdc, sdk.AccAddress{})

	// Create multiStore in memory
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)

	// Mount stores
	cms.MountStoreWithDB(keys[types.StoreKey], storetypes.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	// Create context
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

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
	pk, err := cryptocdc.FromTmPubKeyInterface(tmpk)
	if err != nil {
		panic(err)
	}
	consPubKey := legacybech32.MustMarshalPubKey(legacybech32.ConsPK, pk)

	validator := types.Validator{
		OperatorAddress: operatorAddress,
		ConsensusPubkey: consPubKey,
		Description:     validatorDescription,
	}

	return validator, consPubKey
}
