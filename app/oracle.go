package app

import (
	"context"
	"slices"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	connectpreblocker "github.com/skip-mev/connect/v2/abci/preblock/oracle"
	connectproposals "github.com/skip-mev/connect/v2/abci/proposals"
	"github.com/skip-mev/connect/v2/abci/strategies/aggregator"
	compression "github.com/skip-mev/connect/v2/abci/strategies/codec"
	"github.com/skip-mev/connect/v2/abci/strategies/currencypair"
	connectve "github.com/skip-mev/connect/v2/abci/ve"
	"github.com/skip-mev/connect/v2/cmd/constants/marketmaps"
	oracleconfig "github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/pkg/math/voteweighted"
	oracleclient "github.com/skip-mev/connect/v2/service/clients/oracle"
	servicemetrics "github.com/skip-mev/connect/v2/service/metrics"
)

// connectABCIHandlers returns the Connect ABCI handlers.
func (app *Mezo) connectABCIHandlers() (
	*connectve.VoteExtensionHandler, *connectproposals.ProposalHandler, *connectpreblocker.PreBlockHandler) {

	veCodec := compression.NewCompressionVoteExtensionCodec(
		compression.NewDefaultVoteExtensionCodec(),
		compression.NewZLibCompressor(),
	)
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
		veCodec,
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
		compression.NewCompressionVoteExtensionCodec(
			compression.NewDefaultVoteExtensionCodec(),
			compression.NewZLibCompressor(),
		),
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
		compression.NewCompressionVoteExtensionCodec(
			compression.NewDefaultVoteExtensionCodec(),
			compression.NewZLibCompressor(),
		),
		compression.NewCompressionExtendedCommitCodec(
			compression.NewDefaultExtendedCommitCodec(),
			compression.NewZStdCompressor(),
		),
	)

	return voteExtensionsHandler, proposalHandler, preBlocker
}

func (app *Mezo) setupMarkets(ctx sdk.Context) error {
	// add core markets
	coreMarkets := marketmaps.CoreMarketMap
	markets := coreMarkets.Markets

	// sort keys so we can deterministically iterate over map items.
	keys := make([]string, 0, len(markets))
	for name := range markets {
		keys = append(keys, name)
	}
	slices.Sort(keys)

	for _, marketName := range keys {
		// create market
		market := markets[marketName]
		err := app.MarketMapKeeper.CreateMarket(ctx, market)
		if err != nil {
			return err
		}

		// invoke hooks. this syncs the market to x/oracle.
		err = app.MarketMapKeeper.Hooks().AfterMarketCreated(ctx, market)
		if err != nil {
			return err
		}
	}

	return nil
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
