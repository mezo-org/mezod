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
	FlagServerRequestsPerMinute   = "ethereum-sidecar.server.requests-per-minute"
	FlagKeyringBackend            = "keyring-backend"
	FlagKeyringDir                = "keyring-dir"
	FlagKeyType                   = "keyring-type"
	FlagKeyName                   = "key-name"
)

func NewFlagSetEthereumSidecar(
	defaultServerAddress,
	defaultServerEthereumNodeAddress,
	defaultServerNetwork string,
	defaultServerBatchSize uint64,
	defaultServerRequestsPerMinute uint64,
	defaultKeyringBackend,
	defaultKeyringDir,
	defaultKeyType,
	defaultKeyName string,
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
		FlagKeyringBackend,
		defaultKeyringBackend,
		"Select keyring's backend (os|file|test)",
	)

	fs.String(
		FlagKeyringDir,
		defaultKeyringDir,
		"The client Keyring directory; if omitted, the default 'home' directory will be used",
	)

	fs.String(
		FlagKeyType,
		defaultKeyType,
		"Key signing algorithm to generate keys for",
	)

	fs.String(
		FlagKeyName,
		defaultKeyName,
		"Name of the key to extract from keyring (optional)",
	)

	return fs
}
