package cli

import (
	"fmt"
	"time"

	flag "github.com/spf13/pflag"
)

const (
	FlagServerAddress             = "ethereum-sidecar.server.address"
	FlagServerEthereumNodeAddress = "ethereum-sidecar.server.ethereum-node-address"
	FlagServerNetwork             = "ethereum-sidecar.server.network"
	FlagServerBatchSize           = "ethereum-sidecar.server.batch-size"
	FlagServerRequestsPerMinute   = "ethereum-sidecar.server.requests-per-minute"
	FlagBridgeOutServerAddress    = "ethereum-sidecar.bridge-out.server-address"
	FlagBridgeOutRequestTimeout   = "ethereum-sidecar.bridge-out.request-timeout"
)

func NewFlagSetEthereumSidecar(
	defaultServerAddress,
	defaultServerEthereumNodeAddress,
	defaultServerNetwork string,
	defaultServerBatchSize uint64,
	defaultServerRequestsPerMinute uint64,
	defaultBridgeOutServerAddress string,
	defaultBridgeOutRequestTimeout time.Duration,
) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(
		FlagServerAddress,
		defaultServerAddress,
		fmt.Sprintf(
			"The sidecar server gRPC listen address (e.g. %s)",
			defaultServerAddress,
		),
	)
	fs.String(
		FlagServerEthereumNodeAddress,
		defaultServerEthereumNodeAddress,
		fmt.Sprintf(
			"The sidecar server Ethereum node address (e.g. %s)",
			defaultServerEthereumNodeAddress,
		),
	)
	fs.String(
		FlagServerNetwork,
		defaultServerNetwork,
		fmt.Sprintf(
			"The sidecar server Ethereum network. "+
				"Possible values: mainnet | sepolia | developer "+
				"If not set, sepolia is used by default",
		),
	)

	fs.Uint64(
		FlagServerBatchSize,
		defaultServerBatchSize,
		"Size of the block batch for fallback AssetsLocked events lookup",
	)

	fs.Uint64(
		FlagServerRequestsPerMinute,
		defaultServerRequestsPerMinute,
		"Requests per minute for an Ethereum RPC provider",
	)

	fs.String(
		FlagBridgeOutServerAddress,
		defaultBridgeOutServerAddress,
		"Address of the gRPC bridge-out server. The bridge-out server is "+
			"run by validators as part of mezod and responds with "+
			"AssetsUnlock entries",
	)

	fs.Duration(
		FlagBridgeOutRequestTimeout,
		defaultBridgeOutRequestTimeout,
		"Timeout for AssetsUnlock requests sent to bridge-out server",
	)

	return fs
}
