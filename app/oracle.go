package app

import (
	"context"
	"encoding/json"
	"fmt"
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

var (
	MezoMarketMap     marketmaptypes.MarketMap
	MezoMarketMapJSON = `
	{
		"markets": {
		  "BTC/USD": {
			"ticker": {
			  "currency_pair": {
				"Base": "BTC",
				"Quote": "USD"
			  },
			  "decimals": 5,
			  "min_provider_count": 3,
			  "enabled": true
			},
			"provider_configs": [
			  {
				"name": "binance_ws",
				"off_chain_ticker": "BTCUSDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "bybit_ws",
				"off_chain_ticker": "BTCUSDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "coinbase_ws",
				"off_chain_ticker": "BTC-USD"
			  },
			  {
				"name": "huobi_ws",
				"off_chain_ticker": "btcusdt",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "kraken_api",
				"off_chain_ticker": "XXBTZUSD"
			  },
			  {
				"name": "kucoin_ws",
				"off_chain_ticker": "BTC-USDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "mexc_ws",
				"off_chain_ticker": "BTCUSDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "okx_ws",
				"off_chain_ticker": "BTC-USDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "crypto_dot_com_ws",
				"off_chain_ticker": "BTC_USD"
			  }
			]
		  },
		  "ETH/USD": {
			"ticker": {
			  "currency_pair": {
				"Base": "ETH",
				"Quote": "USD"
			  },
			  "decimals": 6,
			  "min_provider_count": 3,
			  "enabled": true
			},
			"provider_configs": [
			  {
				"name": "binance_ws",
				"off_chain_ticker": "ETHUSDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "bybit_ws",
				"off_chain_ticker": "ETHUSDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "coinbase_ws",
				"off_chain_ticker": "ETH-USD"
			  },
			  {
				"name": "huobi_ws",
				"off_chain_ticker": "ethusdt",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "kraken_api",
				"off_chain_ticker": "XETHZUSD"
			  },
			  {
				"name": "kucoin_ws",
				"off_chain_ticker": "ETH-USDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "mexc_ws",
				"off_chain_ticker": "ETHUSDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "okx_ws",
				"off_chain_ticker": "ETH-USDT",
				"normalize_by_pair": {
				  "Base": "USDT",
				  "Quote": "USD"
				}
			  },
			  {
				"name": "crypto_dot_com_ws",
				"off_chain_ticker": "ETH_USD"
			  }
			]
		  },
		  "USDT/USD": {
			"ticker": {
			  "currency_pair": {
				"Base": "USDT",
				"Quote": "USD"
			  },
			  "decimals": 9,
			  "min_provider_count": 1,
			  "enabled": true
			},
			"provider_configs": [
			  {
				"name": "binance_ws",
				"off_chain_ticker": "USDCUSDT",
				"invert": true
			  },
			  {
				"name": "bybit_ws",
				"off_chain_ticker": "USDCUSDT",
				"invert": true
			  },
			  {
				"name": "coinbase_ws",
				"off_chain_ticker": "USDT-USD"
			  },
			  {
				"name": "huobi_ws",
				"off_chain_ticker": "ethusdt",
				"normalize_by_pair": {
				  "Base": "ETH",
				  "Quote": "USD"
				},
				"invert": true
			  },
			  {
				"name": "kraken_api",
				"off_chain_ticker": "USDTZUSD"
			  },
			  {
				"name": "kucoin_ws",
				"off_chain_ticker": "BTC-USDT",
				"normalize_by_pair": {
				  "Base": "BTC",
				  "Quote": "USD"
				},
				"invert": true
			  },
			  {
				"name": "okx_ws",
				"off_chain_ticker": "USDC-USDT",
				"invert": true
			  },
			  {
				"name": "crypto_dot_com_ws",
				"off_chain_ticker": "USDT_USD"
			  }
			]
		  }
		}
	}
`
)

// unmarshalValidate unmarshalls data into mm and then calls ValidateBasic.
func unmarshalValidate(name, data string, mm *marketmaptypes.MarketMap) error {
	if err := json.Unmarshal([]byte(data), mm); err != nil {
		return fmt.Errorf("failed to unmarshal %sMarketMap: %w", name, err)
	}
	if err := mm.ValidateBasic(); err != nil {
		return fmt.Errorf("%sMarketMap failed validation: %w", name, err)
	}
	return nil
}

func init() {
	if err := unmarshalValidate("MezoMarketMap", MezoMarketMapJSON, &MezoMarketMap); err != nil {
		panic(err)
	}
}

// connectABCIHandlers returns the Connect ABCI handlers.
func (app *Mezo) connectABCIHandlers() (
	*connectve.VoteExtensionHandler, *connectproposals.ProposalHandler, *connectpreblocker.PreBlockHandler,
) {
	veCodec := NewConnectVEExtractionCodec(compression.NewCompressionVoteExtensionCodec(
		compression.NewDefaultVoteExtensionCodec(),
		compression.NewZLibCompressor(),
	))
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
		NewConnectVEExtractionCodec(
			compression.NewCompressionVoteExtensionCodec(
				compression.NewDefaultVoteExtensionCodec(),
				compression.NewZLibCompressor(),
			),
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
		NewConnectVEExtractionCodec(
			compression.NewCompressionVoteExtensionCodec(
				compression.NewDefaultVoteExtensionCodec(),
				compression.NewZLibCompressor(),
			),
		),
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
	marketmapGenState.MarketMap = MezoMarketMap

	// update oracle genesis state
	id := uint64(1)
	for _, market := range MezoMarketMap.Markets {
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
func (c *ConnectVEExtractionCodec) Encode(ve vetypes.OracleVoteExtension) ([]byte, error) {
	return c.codec.Encode(ve)
}

// Decode takes a set of vote extension data and returns the OracleVoteExtension
func (c *ConnectVEExtractionCodec) Decode(veBytes []byte) (vetypes.OracleVoteExtension, error) {
	connectVEBytes, err := abci.VoteExtensionDecomposer(abci.VoteExtensionPartConnect)(veBytes)
	if err != nil {
		return vetypes.OracleVoteExtension{}, err
	}
	return c.codec.Decode(connectVEBytes)

}
