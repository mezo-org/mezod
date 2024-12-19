package cli

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

const (
	FlagServerAddress             = "ethereum-sidecar.server.address"
	FlagServerEthereumNodeAddress = "ethereum-sidecar.server.ethereum-node-address"
	FlagServerNetwork             = "ethereum-sidecar.server.network"
	FlagServerBatchSize           = "ethereum-sidecar.server.batch-size"
)

func NewFlagSetEthereumSidecar(
	defaultServerAddress,
	defaultServerEthereumNodeAddress,
	defaultServerNetwork string,
	defaultServerBatchSize uint64,
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

	return fs
}
