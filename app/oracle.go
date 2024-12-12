package app

import (
	"context"
	"slices"

	"github.com/mezo-org/mezod/types"

	"golang.org/x/exp/maps"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/mezo-org/mezod/app/abci"
	connectpreblocker "github.com/skip-mev/connect/v2/abci/preblock/oracle"
	connectproposals "github.com/skip-mev/connect/v2/abci/proposals"
	"github.com/skip-mev/connect/v2/abci/strategies/aggregator"
	compression "github.com/skip-mev/connect/v2/abci/strategies/codec"
	"github.com/skip-mev/connect/v2/abci/strategies/currencypair"
	connectve "github.com/skip-mev/connect/v2/abci/ve"
	vetypes "github.com/skip-mev/connect/v2/abci/ve/types"
	oracleconfig "github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/pkg/math/voteweighted"
	oracleclient "github.com/skip-mev/connect/v2/service/clients/oracle"
	servicemetrics "github.com/skip-mev/connect/v2/service/metrics"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
)

// connectABCIHandlers returns the Connect ABCI handlers.
func (app *Mezo) connectABCIHandlers() (
	*connectve.VoteExtensionHandler, *connectproposals.ProposalHandler, *connectpreblocker.PreBlockHandler,
) {
	// The basic VE codec is meant to be used by the Connect vote extension
	// sub-handler. In our composite model, each vote extension sub-handler
	// produces its specific vote extension part directly and the same
	// part is later given for verification. That means the Connect vote
	// extension sub-handler can encode/decode the vote extensions directly.
	veCodec := compression.NewCompressionVoteExtensionCodec(
		compression.NewDefaultVoteExtensionCodec(),
		compression.NewZLibCompressor(),
	)
	// The composite VE codec is meant to be used by:
	// - The oracle price applier used within the Connect vote extension sub-handler,
	// - The Connect proposal sub-handler,
	// - The Connect pre-block handler.
	// This is because those components decode Connect-specific vote extensions
	// from the extended commit info. In our composite model, the extended
	// commit info contains composite vote extensions so Connect-specific
	// parts are not directly available. The composite vote extensions must be
	// decomposed first. This is exactly what ConnectVEExtractionCodec does.
	// Note that ConnectVEExtractionCodec does not support encoding. This
	// action is not needed here. Connect-specific vote extensions
	// are encoded by the Connect vote extension sub-handler
	// (using the basic VE codec, see veCodec) and then, by the composite
	// app-level handler.
	compositeVeCodec := NewConnectVEExtractionCodec(veCodec)

	extCommitCodec := compression.NewCompressionExtendedCommitCodec(
		compression.NewDefaultExtendedCommitCodec(),
		compression.NewZStdCompressor(),
	)

	priceApplier := aggregator.NewOraclePriceApplier(
		aggregator.NewDefaultVoteAggregator(
			app.Logger(),
			voteweighted.MedianFromContext(
				app.Logger(),
				app.PoaKeeper,
				voteweighted.DefaultPowerThreshold,
			),
			currencypair.NewDeltaCurrencyPairStrategy(&app.OracleKeeper),
		),
		&app.OracleKeeper,
		compositeVeCodec,
		extCommitCodec,
		app.Logger(),
	)

	voteExtensionsHandler := connectve.NewVoteExtensionHandler(
		app.Logger(),
		app.oracleClient,
		DefaultOracleTimeout,
		currencypair.NewDeltaCurrencyPairStrategy(&app.OracleKeeper),
		veCodec,
		priceApplier,
		app.oracleMetrics,
	)

	proposalHandler := connectproposals.NewProposalHandler(
		app.Logger(),
		// Inject no-ops here since we're not wrapping other handlers, we're including ours as a sub-handler
		baseapp.NoOpPrepareProposal(),
		baseapp.NoOpProcessProposal(),
		connectve.NewDefaultValidateVoteExtensionsFn(app.PoaKeeper),
		compositeVeCodec,
		compression.NewCompressionExtendedCommitCodec(
			compression.NewDefaultExtendedCommitCodec(),
			compression.NewZStdCompressor(),
		),
		currencypair.NewDeltaCurrencyPairStrategy(&app.OracleKeeper),
		app.oracleMetrics,
	)

	aggregatorFn := voteweighted.MedianFromContext(
		app.Logger(),
		app.PoaKeeper,
		voteweighted.DefaultPowerThreshold,
	)
	preBlocker := connectpreblocker.NewOraclePreBlockHandler(
		app.Logger(),
		aggregatorFn,
		&app.OracleKeeper,
		app.oracleMetrics,
		currencypair.NewDeltaCurrencyPairStrategy(&app.OracleKeeper),
		compositeVeCodec,
		compression.NewCompressionExtendedCommitCodec(
			compression.NewDefaultExtendedCommitCodec(),
			compression.NewZStdCompressor(),
		),
	)

	return voteExtensionsHandler, proposalHandler, preBlocker
}

func customMarketGenesis() (*oracletypes.GenesisState, *marketmaptypes.GenesisState) {
	// Get defaults
	oracleGenState := oracletypes.DefaultGenesisState()
	marketmapGenState := marketmaptypes.DefaultGenesisState()

	// Update Markets
	marketmapGenState.MarketMap = types.MezoMarketMap

	// Ensure deterministic order of markets in genesis. This is a must
	// so all nodes get the same ID for the same currency pair.
	marketsKeys := maps.Keys(types.MezoMarketMap.Markets)
	slices.Sort(marketsKeys)

	// update oracle genesis state
	id := uint64(1)
	for _, marketKey := range marketsKeys {
		market := types.MezoMarketMap.Markets[marketKey]

		cp := oracletypes.CurrencyPairGenesis{
			Id:                id,
			Nonce:             0,
			CurrencyPairPrice: nil,
			CurrencyPair:      market.Ticker.CurrencyPair,
		}
		id++
		oracleGenState.CurrencyPairGenesis = append(oracleGenState.CurrencyPairGenesis, cp)
	}
	oracleGenState.NextId = id

	return oracleGenState, marketmapGenState
}

// initializeOracle initializes the oracle client and metrics.
func (app *Mezo) initializeOracle(appOpts servertypes.AppOptions) (oracleclient.OracleClient, servicemetrics.Metrics, error) {
	// Read general config from app-opts, and construct oracle service.
	cfg, err := oracleconfig.ReadConfigFromAppOpts(appOpts)
	if err != nil {
		return nil, nil, err
	}

	// If app level instrumentation is enabled, then wrap the oracle service with a metrics client
	// to get metrics on the oracle service (for ABCI++). This will allow the instrumentation to track
	// latency in VerifyVoteExtension requests and more.
	oracleMetrics, err := servicemetrics.NewMetricsFromConfig(cfg, app.ChainID())
	if err != nil {
		return nil, nil, err
	}

	// Create the oracle service.
	oracleClient, err := oracleclient.NewPriceDaemonClientFromConfig(
		cfg,
		app.Logger().With("client", "oracle"),
		oracleMetrics,
	)
	if err != nil {
		return nil, nil, err
	}

	// Connect to the oracle service (default timeout of 5 seconds).
	go func() {
		app.Logger().Info("attempting to start oracle client...", "address", cfg.OracleAddress)
		if err := oracleClient.Start(context.Background()); err != nil {
			app.Logger().Error("failed to start oracle client", "err", err)
			panic(err)
		}
	}()

	return oracleClient, oracleMetrics, nil
}

type ConnectVEExtractionCodec struct {
	codec compression.VoteExtensionCodec
}

func NewConnectVEExtractionCodec(codec compression.VoteExtensionCodec) *ConnectVEExtractionCodec {
	return &ConnectVEExtractionCodec{codec: codec}
}

// Encode just passes through to the wrapped VoteExtensionCodec
func (c *ConnectVEExtractionCodec) Encode(vetypes.OracleVoteExtension) ([]byte, error) {
	// The sole purpose of ConnectVEExtractionCodec is to decode Connect-specific
	// vote extensions from composite vote extensions. It should never be used
	// for encoding. Encoding happens elsewhere.
	panic("this codec should not be used for encoding")
}

// Decode takes a set of vote extension data and returns the OracleVoteExtension
func (c *ConnectVEExtractionCodec) Decode(veBytes []byte) (vetypes.OracleVoteExtension, error) {
	connectVEBytes, err := abci.VoteExtensionDecomposer(abci.VoteExtensionPartConnect)(veBytes)
	if err != nil {
		return vetypes.OracleVoteExtension{}, err
	}
	return c.codec.Decode(connectVEBytes)
}
