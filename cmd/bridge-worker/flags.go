package main

import (
	bridgeworker "github.com/mezo-org/mezod/bridge-worker"
	flag "github.com/spf13/pflag"
)

// Flags list
const (
	flagLogLevel = "log-level"

	flagEthereumProviderURL       = "ethereum.provider-url"
	flagEthereumNetwork           = "ethereum.network"
	flagEthereumBatchSize         = "ethereum.batch-size"
	flagEthereumRequestsPerMinute = "ethereum.requests-per-minute"
	flagEthereumAccountKeyFile    = "ethereum.account.key-file"

	flagBitcoinNetwork     = "bitcoin.network"
	flagBitcoinElectrumURL = "bitcoin.electrum.url"

	flagMezoAssetsUnlockEndpoint = "mezo.assets-unlock-endpoint"
)

// Flags default values
const (
	flagLogLevelDefault = "info"

	flagEthereumBatchSizeDefault         = bridgeworker.DefaultEthereumBatchSize
	flagEthereumRequestsPerMinuteDefault = bridgeworker.DefaultEthereumRequestsPerMinute
)

func newFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(
		flagLogLevel,
		flagLogLevelDefault,
		"Log level",
	)

	fs.String(
		flagEthereumProviderURL,
		"",
		"Ethereum provider URL (must be websocket)",
	)

	fs.String(
		flagEthereumNetwork,
		"",
		"Ethereum network (must be sepolia, mainnet, or developer)",
	)

	fs.Uint64(
		flagEthereumBatchSize,
		flagEthereumBatchSizeDefault,
		"Ethereum batch size used for partial events lookup",
	)

	fs.Uint64(
		flagEthereumRequestsPerMinute,
		flagEthereumRequestsPerMinuteDefault,
		"Ethereum requests per minute",
	)

	fs.String(
		flagEthereumAccountKeyFile,
		"",
		"Ethereum account key file",
	)

	fs.String(
		flagBitcoinNetwork,
		"",
		"Bitcoin network (must be mainnet, testnet, or regtest)",
	)

	fs.String(
		flagBitcoinElectrumURL,
		"",
		"Bitcoin Electrum URL",
	)

	fs.String(
		flagMezoAssetsUnlockEndpoint,
		"",
		"Mezo assets unlock endpoint (must be the gRPC endpoint of a mezod node; e.g. localhost:9090)",
	)

	return fs
}
