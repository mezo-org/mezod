package main

import (
	bridgeworker "github.com/mezo-org/mezod/bridge-worker"
	flag "github.com/spf13/pflag"
)

// Flags list
const (
	flagLogLevel      = "log.level"
	flagLogFormatJSON = "log.format-json"

	flagEthereumProviderURL       = "ethereum.provider-url"
	flagEthereumNetwork           = "ethereum.network"
	flagEthereumBatchSize         = "ethereum.batch-size"
	flagEthereumRequestsPerMinute = "ethereum.requests-per-minute"
	flagEthereumAccountKeyFile    = "ethereum.account.key-file"

	flagBitcoinNetwork     = "bitcoin.network"
	flagBitcoinElectrumURL = "bitcoin.electrum.url"

	flagMezoAssetsUnlockEndpoint = "mezo.assets-unlock-endpoint"

	flagJobBTCWithdrawalQueueCheckFrequency = "job.btc-withdrawal.queue-check-frequency"

	flagPrometheusPort = "prometheus-port"
)

// Flags default values
const (
	flagLogLevelDefault      = "info"
	flagLogFormatJSONDefault = false

	flagEthereumBatchSizeDefault         = bridgeworker.DefaultEthereumBatchSize
	flagEthereumRequestsPerMinuteDefault = bridgeworker.DefaultEthereumRequestsPerMinute

	flagJobBTCWithdrawalQueueCheckFrequencyDefault = bridgeworker.DefaultBTCWithdrawalQueueCheckFrequency

	flagPrometheusPortDefault = 2112
)

func newFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(
		flagLogLevel,
		flagLogLevelDefault,
		"Log level",
	)

	fs.Bool(
		flagLogFormatJSON,
		flagLogFormatJSONDefault,
		"Format logs as JSON",
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

	fs.Duration(
		flagJobBTCWithdrawalQueueCheckFrequency,
		flagJobBTCWithdrawalQueueCheckFrequencyDefault,
		"Frequency of the queue check made by the BTC withdrawal job",
	)

	fs.Uint(
		flagPrometheusPort,
		flagPrometheusPortDefault,
		"Port to expose Prometheus metrics on",
	)

	return fs
}
