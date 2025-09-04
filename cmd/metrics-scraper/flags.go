package main

import (
	"time"

	flag "github.com/spf13/pflag"
)

// Flags list
const (
	flagNodesConfig    = "nodes-config"
	flagPrometheusPort = "prometheus-port"
	flagChainID        = "chain-id"
	flagNodePollRate   = "node-poll-rate"
	flagBridgePollRate = "bridge-poll-rate"
	flagMezoRPCURL     = "mezo-rpc-url"
	flagEthereumRPCURL = "ethereum-rpc-url"
)

// Flags default values
const (
	flagNodesConfigDefault    = "nodes-config.json"
	flagPrometheusPortDefault = 2112
	flagNodePollRateDefault   = 30 * time.Second
	flagBridgePollRateDefault = 5 * time.Minute
)

func newFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(
		flagNodesConfig,
		flagNodesConfigDefault,
		"Path to the JSON config file containing the mezod nodes to monitor",
	)

	fs.Uint(
		flagPrometheusPort,
		flagPrometheusPortDefault,
		"Port to expose Prometheus metrics on",
	)

	fs.String(
		flagChainID,
		"",
		"Mezo chain ID (mezo_31612-1 for mainnet or mezo_31611-1 for testnet)",
	)

	fs.Duration(
		flagNodePollRate,
		flagNodePollRateDefault,
		"Frequency of polling data from the mezod nodes",
	)

	fs.Duration(
		flagBridgePollRate,
		flagBridgePollRateDefault,
		"Frequency of polling data from the bridge components",
	)

	fs.String(
		flagMezoRPCURL,
		"",
		"Mezo RPC URL",
	)

	fs.String(
		flagEthereumRPCURL,
		"",
		"Ethereum RPC URL",
	)

	return fs
}
