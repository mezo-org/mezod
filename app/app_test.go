package app

import (
	"encoding/json"
	kitlog "github.com/go-kit/log"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/cometbft/cometbft/crypto/ed25519"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"cosmossdk.io/log"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/mezo-org/mezod/encoding"
	"github.com/mezo-org/mezod/utils"
)

func TestMezoExport(t *testing.T) {
	// create public key
	privVal := ed25519.GenPrivKey()
	pubKey := privVal.PubKey()

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(100000000000000))),
	}

	db := dbm.NewMemDB()
	chainID := utils.MainnetChainID + "-1"
	app := NewMezo(
		log.NewLogger(kitlog.NewSyncWriter(os.Stdout)),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		encoding.MakeConfig(ModuleBasics),
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(chainID),
	)

	genesisState := NewDefaultGenesisState()
	genesisState = GenesisStateWithValSet(
		app,
		genesisState,
		acc.GetAddress(),
		valSet,
		[]authtypes.GenesisAccount{acc},
		balance,
	)
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			ChainId:       utils.MainnetChainID + "-1",
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewMezo(
		log.NewLogger(kitlog.NewSyncWriter(os.Stdout)),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		encoding.MakeConfig(ModuleBasics),
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(chainID),
	)
	_, err = app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
