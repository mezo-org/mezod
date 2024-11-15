// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/runtime"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/btctoken"
	"github.com/mezo-org/mezod/precompile/validatorpool"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	tmos "github.com/cometbft/cometbft/libs/os"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"

	"cosmossdk.io/simapp"
	simappparams "cosmossdk.io/simapp/params"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparams "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamstypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	// Connect imports
	connectpreblocker "github.com/skip-mev/connect/v2/abci/preblock/oracle"
	oracleclient "github.com/skip-mev/connect/v2/service/clients/oracle"
	servicemetrics "github.com/skip-mev/connect/v2/service/metrics"
	"github.com/skip-mev/connect/v2/x/marketmap"
	marketmapkeeper "github.com/skip-mev/connect/v2/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/oracle"
	oraclekeeper "github.com/skip-mev/connect/v2/x/oracle/keeper"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	appabci "github.com/mezo-org/mezod/app/abci"
	ethante "github.com/mezo-org/mezod/app/ante/evm"
	"github.com/mezo-org/mezod/encoding"
	"github.com/mezo-org/mezod/ethereum/eip712"
	srvflags "github.com/mezo-org/mezod/server/flags"
	mezotypes "github.com/mezo-org/mezod/types"

	"github.com/mezo-org/mezod/x/evm"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/mezo-org/mezod/x/feemarket"
	feemarketkeeper "github.com/mezo-org/mezod/x/feemarket/keeper"
	feemarkettypes "github.com/mezo-org/mezod/x/feemarket/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/mezo-org/mezod/client/docs/statik"

	"github.com/mezo-org/mezod/app/ante"
	"github.com/mezo-org/mezod/x/bridge"
	bridgeabci "github.com/mezo-org/mezod/x/bridge/abci"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/mezo-org/mezod/x/poa"
	poakeeper "github.com/mezo-org/mezod/x/poa/keeper"
	poatypes "github.com/mezo-org/mezod/x/poa/types"

	// Force-load the tracer engines to trigger registration due to Go-Ethereum v1.10.15 changes

	//nolint:revive
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"

	//nolint:revive
	_ "github.com/ethereum/go-ethereum/eth/tracers/native"
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".mezod")

	// manually update the power reduction by replacing micro (u) -> atto (a) btc
	sdk.DefaultPowerReduction = mezotypes.PowerReduction
	// modify fee market parameter defaults through global
	feemarkettypes.DefaultMinGasPrice = MainnetMinGasPrices
	feemarkettypes.DefaultMinGasMultiplier = MainnetMinGasMultiplier
}

// Name defines the application binary name
const Name = "mezod"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	DefaultOracleTimeout = time.Second

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		consensusparams.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		poa.AppModuleBasic{},
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evm.AppModuleBasic{},
		feemarket.AppModuleBasic{},
		bridge.AppModuleBasic{},
		marketmap.AppModuleBasic{},
		oracle.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		poatypes.ModuleName:        nil,
		evmtypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
		bridgetypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{}
)

var _ servertypes.Application = (*Mezo)(nil)

// Mezo implements an extended ABCI application. It is an application
// that may process transactions through Ethereum's EVM running atop of
// Tendermint consensus.
type Mezo struct {
	*baseapp.BaseApp

	// encoding
	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*storetypes.KVStoreKey
	tkeys map[string]*storetypes.TransientStoreKey

	// keepers
	ConsensusParamsKeeper consensusparamskeeper.Keeper
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	PoaKeeper             poakeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvmKeeper             *evmkeeper.Keeper
	FeeMarketKeeper       feemarketkeeper.Keeper
	BridgeKeeper          bridgekeeper.Keeper
	OracleKeeper          oraclekeeper.Keeper
	MarketMapKeeper       marketmapkeeper.Keeper

	// the module manager
	mm *module.Manager

	// the configurator
	configurator module.Configurator

	tpsCounter *tpsCounter

	// Connect client
	oracleClient      oracleclient.OracleClient
	oracleMetrics     servicemetrics.Metrics
	connectPreBlocker *connectpreblocker.PreBlockHandler

	preBlockHandler *appabci.PreBlockHandler
}

// NewMezo returns a reference to a new initialized Ethermint application.
func NewMezo(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig simappparams.EncodingConfig,
	ethereumSidecarClient bridgeabci.EthereumSidecarClient,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *Mezo {
	appCodec := encodingConfig.Codec
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	eip712.SetEncodingConfig(encodingConfig)

	// NOTE we use custom transaction decoder that supports the sdk.Tx interface instead of sdk.StdTx
	bApp := baseapp.NewBaseApp(
		Name,
		logger,
		db,
		encodingConfig.TxConfig.TxDecoder(),
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := storetypes.NewKVStoreKeys(
		consensusparamstypes.StoreKey,
		authtypes.StoreKey,
		banktypes.StoreKey,
		poatypes.StoreKey,
		crisistypes.StoreKey,
		paramstypes.StoreKey,
		authzkeeper.StoreKey,
		upgradetypes.StoreKey,
		evmtypes.StoreKey,
		feemarkettypes.StoreKey,
		bridgetypes.StoreKey,
		marketmaptypes.StoreKey,
		oracletypes.StoreKey,
	)

	tkeys := storetypes.NewTransientStoreKeys(
		paramstypes.TStoreKey,
		evmtypes.TransientKey,
		feemarkettypes.TransientKey,
	)

	app := &Mezo{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
	}

	if err := app.RegisterStreamingServices(appOpts, app.keys); err != nil {
		panic(fmt.Sprintf("failed to register streaming services: %s", err))
	}

	// Most of the modules require setting a Cosmos-level authority account
	// which has privileges to perform governance actions (e.g. parameters change).
	// Upon a governance action, the modules' keepers perform a check that
	// the operation is actually executed by the authority account. This is
	// necessary to ensure proper authorization of governance actions done
	// through native Cosmos transactions. For Mezo, the actual authority will
	// be in hands of a multi-sig account deployed on EVM. Moreover, the governance
	// actions will be exposed through dedicated precompiled EVM contracts.
	// Those precompiles will validate the authority of the caller on EVM-level
	// and will execute state updates on specific modules keepers. However,
	// given the Cosmos-level authority check in keepers, the precompiles
	// will have to impersonate the Cosmos-level authority account.
	// As the precompiles live in the context of the `x/evm` module, using
	// the account of this module as authority seems to be a natural choice.
	authority := authtypes.NewModuleAddress(evmtypes.ModuleName)

	// init params keeper and subspaces
	app.ParamsKeeper = initParamsKeeper(appCodec, cdc, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
	// init consensus params keeper
	app.ConsensusParamsKeeper = consensusparamskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamstypes.StoreKey]),
		authority.String(),
		runtime.EventService{},
	)
	// set the BaseApp's parameter store
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	addressCodec := authcodec.NewBech32Codec(bech32Prefix)

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		mezotypes.ProtoAccount,
		maccPerms,
		addressCodec,
		bech32Prefix,
		authority.String(),
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.BlockedAddrs(),
		authority.String(),
		logger,
	)
	app.PoaKeeper = poakeeper.NewKeeper(
		keys[poatypes.StoreKey],
		appCodec,
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authority.String(),
		app.AccountKeeper.AddressCodec(),
	)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authority.String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	tracer := cast.ToString(appOpts.Get(srvflags.EVMTracer))

	app.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		appCodec, authority,
		keys[feemarkettypes.StoreKey],
		tkeys[feemarkettypes.TransientKey],
		app.GetSubspace(feemarkettypes.ModuleName),
	)

	app.EvmKeeper = evmkeeper.NewKeeper(
		appCodec,
		keys[evmtypes.StoreKey],
		tkeys[evmtypes.TransientKey],
		authority,
		app.AccountKeeper,
		app.BankKeeper,
		app.PoaKeeper,
		app.FeeMarketKeeper,
		&app.ConsensusParamsKeeper,
		tracer,
		app.GetSubspace(evmtypes.ModuleName),
	)

	precompiles, err := customEvmPrecompiles(app.BankKeeper, app.AuthzKeeper, app.PoaKeeper, *app.EvmKeeper, bApp.ChainID())
	if err != nil {
		panic(fmt.Sprintf("failed to build custom EVM precompiles: [%s]", err))
	}
	app.EvmKeeper.RegisterCustomPrecompiles(precompiles...)

	app.BridgeKeeper = bridgekeeper.NewKeeper(
		appCodec,
		keys[bridgetypes.StoreKey],
		app.BankKeeper,
	)

	app.MarketMapKeeper = *marketmapkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[marketmaptypes.StoreKey]),
		appCodec,
		authority,
	)
	app.OracleKeeper = oraclekeeper.NewKeeper(
		runtime.NewKVStoreService(keys[oracletypes.StoreKey]),
		appCodec,
		&app.MarketMapKeeper,
		authority,
	)

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		consensusparams.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		poa.NewAppModule(app.PoaKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper, addressCodec),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		evm.NewAppModule(app.EvmKeeper, app.AccountKeeper, app.GetSubspace(evmtypes.ModuleName)),
		feemarket.NewAppModule(app.FeeMarketKeeper, app.GetSubspace(feemarkettypes.ModuleName)),
		bridge.NewAppModule(app.BridgeKeeper),
		marketmap.NewAppModule(appCodec, &app.MarketMapKeeper),
		oracle.NewAppModule(appCodec, app.OracleKeeper),
	)

	// NOTE: upgrade module must go first to handle software upgrades.
	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	app.mm.SetOrderBeginBlockers(
		feemarkettypes.ModuleName,
		evmtypes.ModuleName,
		poatypes.ModuleName,
		oracletypes.ModuleName,
		// no-op modules
		authtypes.ModuleName,
		banktypes.ModuleName,
		crisistypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		bridgetypes.ModuleName,
		consensusparamstypes.ModuleName,
		marketmaptypes.ModuleName,
	)

	// NOTE: fee market module must go last in order to retrieve the block gas used.
	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		poatypes.ModuleName,
		evmtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		bridgetypes.ModuleName,
		consensusparamstypes.ModuleName,
		marketmaptypes.ModuleName,
		oracletypes.ModuleName,
		feemarkettypes.ModuleName,
	)

	// NOTE: crisis module must go at the end to check for invariants on each module
	app.mm.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		poatypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		bridgetypes.ModuleName,
		marketmaptypes.ModuleName,
		oracletypes.ModuleName,
		crisistypes.ModuleName,
		consensusparamstypes.ModuleName,
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	// initialize the BaseApp with markets in state.
	app.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		// set vote extension height. (must be greater than 1).
		req.ConsensusParams.Abci.VoteExtensionsEnableHeight = 2

		// initialize module state
		app.OracleKeeper.InitGenesis(ctx, *oracletypes.DefaultGenesisState())
		app.MarketMapKeeper.InitGenesis(ctx, *marketmaptypes.DefaultGenesisState())

		// initialize markets
		err := app.setupMarkets(ctx)
		if err != nil {
			return nil, err
		}

		return app.InitChainer(ctx, req)
	})
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)

	maxGasWanted := cast.ToUint64(appOpts.Get(srvflags.EVMMaxTxGasWanted))

	app.setAnteHandler(encodingConfig.TxConfig, maxGasWanted)
	app.setPostHandler()
	app.SetEndBlocker(app.EndBlocker)

	// Set the x/marketmap keeper hooks
	app.MarketMapKeeper.SetHooks(app.OracleKeeper.Hooks())
	// oracle initialization
	app.oracleClient, app.oracleMetrics, err = app.initializeOracle(appOpts)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize oracle client and metrics: %s", err))
	}
	// Connect ABCI initialization requires the oracle client/metrics to be setup first.
	app.setABCIExtensions(ethereumSidecarClient)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	// Finally start the tpsCounter.
	app.tpsCounter = newTPSCounter(logger)
	go func() {
		// Unfortunately golangci-lint is so pedantic,
		// so we have to ignore this error explicitly.
		_ = app.tpsCounter.start(context.Background())
	}()

	return app
}

// Name returns the name of the App
func (app *Mezo) Name() string { return app.BaseApp.Name() }

func (app *Mezo) setAnteHandler(txConfig client.TxConfig, maxGasWanted uint64) {
	options := ante.HandlerOptions{
		Cdc:                    app.appCodec,
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		ExtensionOptionChecker: mezotypes.HasDynamicFeeExtensionOption,
		EvmKeeper:              app.EvmKeeper,
		FeeMarketKeeper:        app.FeeMarketKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		SigGasConsumer:         ante.SigVerificationGasConsumer,
		MaxTxGasWanted:         maxGasWanted,
		TxFeeChecker:           ethante.NewDynamicFeeChecker(app.EvmKeeper),
	}

	if err := options.Validate(); err != nil {
		panic(err)
	}

	app.SetAnteHandler(ante.NewAnteHandler(options))
}

func (app *Mezo) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

func (app *Mezo) PreBlocker(
	ctx sdk.Context,
	req *abci.RequestFinalizeBlock,
) (*sdk.ResponsePreBlock, error) {
	// TODO eric
	// return app.connectPreBlocker.WrappedPreBlocker(app.mm)(ctx, req)
	return app.preBlockHandler.PreBlocker(app.mm)(ctx, req)
}

func (app *Mezo) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

func (app *Mezo) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// FinalizeBlock method is intentionally decomposed to calculate the
// transactions per second.
func (app *Mezo) FinalizeBlock(req *abci.RequestFinalizeBlock) (
	res *abci.ResponseFinalizeBlock,
	err error,
) {
	defer func() {
		// Check required to not panic during res.TxResults in case the
		// upstream FinalizeBlock errors out and returns a nil response.
		if res == nil {
			return
		}

		for _, txResult := range res.TxResults {
			if txResult.IsErr() {
				app.tpsCounter.incrementFailure()
			} else {
				app.tpsCounter.incrementSuccess()
			}
		}
	}()

	return app.BaseApp.FinalizeBlock(req)
}

// InitChainer updates at chain initialization
func (app *Mezo) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState simapp.GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	if err != nil {
		panic(err)
	}

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// setABCIExtensions sets the ABCI++ extensions on the application.
// This function assumes the BridgeKeeper and PoaKeeper are already set in the app.
func (app *Mezo) setABCIExtensions(
	ethereumSidecarClient bridgeabci.EthereumSidecarClient,
) {
	// Create the bridge ABCI handlers.
	bridgeVoteExtensionHandler, bridgeProposalHandler, bridgePreBlockHandler := app.bridgeABCIHandlers(ethereumSidecarClient)

	// Create the Connect ABCI handlers.
	connectVEHandler, connectProposalHandler, connectPreBlocker := app.connectABCIHandlers()

	// Create and attach the app-level composite vote extension handler for
	// ExtendVote and VerifyVoteExtension ABCI requests.
	voteExtensionHandler := appabci.NewVoteExtensionHandler(
		app.Logger(),
		bridgeVoteExtensionHandler,
		connectVEHandler,
	)
	voteExtensionHandler.SetHandlers(app.BaseApp)

	// Create and attach the app-level composite proposal handler for
	// PrepareProposal and ProcessProposal ABCI requests.
	proposalHandler := appabci.NewProposalHandler(
		app.Logger(),
		bridgeProposalHandler,
		connectProposalHandler,
	)
	proposalHandler.SetHandlers(app.BaseApp)

	app.connectPreBlocker = connectPreBlocker
	// TODO eric
	app.preBlockHandler = appabci.NewPreBlockHandler(
		app.Logger(),
		bridgePreBlockHandler,
	)
}

// bridgeABCIHandlers returns the bridge ABCI handlers.
// This function assumes the BridgeKeeper and PoaKeeper are already set in the app.
func (app *Mezo) bridgeABCIHandlers(
	ethereumSidecarClient bridgeabci.EthereumSidecarClient,
) (
	*bridgeabci.VoteExtensionHandler,
	*bridgeabci.ProposalHandler,
	*bridgeabci.PreBlockHandler,
) {
	voteExtensionHandler := bridgeabci.NewVoteExtensionHandler(
		app.Logger(),
		ethereumSidecarClient,
		app.BridgeKeeper,
	)

	proposalHandler := bridgeabci.NewProposalHandler(
		app.Logger(),
		app.PoaKeeper,
		app.BridgeKeeper,
		appabci.VoteExtensionDecomposer(appabci.VoteExtensionPartBridge),
		baseapp.ValidateVoteExtensions,
	)

	preBlockHandler := bridgeabci.NewPreBlockHandler(
		app.Logger(),
		app.BridgeKeeper,
	)

	return voteExtensionHandler, proposalHandler, preBlockHandler
}

// LoadHeight loads state at a particular height
func (app *Mezo) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *Mezo) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)

	accs := make([]string, 0, len(maccPerms))
	for k := range maccPerms {
		accs = append(accs, k)
	}
	sort.Strings(accs)

	for _, acc := range accs {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *Mezo) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)

	accs := make([]string, 0, len(maccPerms))
	for k := range maccPerms {
		accs = append(accs, k)
	}
	sort.Strings(accs)

	for _, acc := range accs {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blockedAddrs
}

// LegacyAmino returns Mezo's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *Mezo) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns Mezo's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *Mezo) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Mezo's InterfaceRegistry
func (app *Mezo) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *Mezo) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *Mezo) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *Mezo) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *Mezo) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register node gRPC service for grpc-gateway.
	node.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

func (app *Mezo) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *Mezo) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterNodeService registers the node gRPC service on the provided
// application gRPC query router.
func (app *Mezo) RegisterNodeService(
	clientCtx client.Context,
	cfg config.Config,
) {
	node.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// GetBaseApp implements the TestingApp interface.
func (app *Mezo) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetTxConfig implements the TestingApp interface.
func (app *Mezo) GetTxConfig() client.TxConfig {
	cfg := encoding.MakeConfig(ModuleBasics)
	return cfg.TxConfig
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(
	appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey,
) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(evmtypes.ModuleName).WithKeyTable(evmtypes.ParamKeyTable()) //nolint: staticcheck
	paramsKeeper.Subspace(feemarkettypes.ModuleName).WithKeyTable(feemarkettypes.ParamKeyTable())

	return paramsKeeper
}

// customEvmPrecompiles builds custom precompiles of the EVM module.
func customEvmPrecompiles(
	bankKeeper bankkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	poaKeeper poakeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	chainID string,
) ([]*precompile.Contract, error) {
	btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper, authzKeeper, evmKeeper, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create BTC token precompile: [%w]", err)
	}

	validatorPoolPrecompile, err := validatorpool.NewPrecompile(poaKeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to create validatorpool precompile: [%w]", err)
	}

	return []*precompile.Contract{
		btcTokenPrecompile,
		validatorPoolPrecompile,
	}, nil
}
