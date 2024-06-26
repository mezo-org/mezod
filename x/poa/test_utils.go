package poa

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/evmos/evmos/v12/x/poa/keeper"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// This package contains various mocks for testing purpose

// Context and keeper used for mocking purpose
func MockContext() (sdk.Context, keeper.Keeper) {
	// Store keys
	keys := sdk.NewKVStoreKeys(types.StoreKey, params.StoreKey)
	tKeys := sdk.NewTransientStoreKeys(params.TStoreKey)

	cdc := codec.New()

	// Create the params keeper
	paramsKeeper := params.NewKeeper(cdc, keys[params.StoreKey], tKeys[params.TStoreKey])

	// Create a poa keeper
	poaKeeper := keeper.NewKeeper(cdc, keys[types.StoreKey], paramsKeeper.Subspace(types.ModuleName))

	// Create multiStore in memory
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)

	// Mount stores
	cms.MountStoreWithDB(keys[types.StoreKey], sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(keys[params.StoreKey], sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tKeys[params.TStoreKey], sdk.StoreTypeTransient, db)
	cms.LoadLatestVersion()

	// Create context
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())

	return ctx, poaKeeper
}

// Create a validator for test
func MockValidator() (types.Validator, string) {
	// Junk description
	validatorDescription := types.Description{
		Moniker:         "Moniker",
		Identity:        "Identity",
		Website:         "Website",
		SecurityContact: "SecurityContact",
		Details:         "Details",
	}

	// Generate operator address
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	operatorAddress := sdk.ValAddress(addr)

	// Generate a consPubKey
	pk = ed25519.GenPrivKey().PubKey()
	consPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, pk)
	if err != nil {
		panic(fmt.Sprintf("Cannot create a consPubKey: %v", err))
	}
	validator := types.Validator{
		OperatorAddress: operatorAddress,
		ConsensusPubkey: consPubKey,
		Description:     validatorDescription,
	}

	return validator, consPubKey
}

// Create an validator address
func MockValAddress() sdk.ValAddress {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.ValAddress(addr)
}
