package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagServerAddress             = "ethereum-sidecar.server.address"
	FlagServerEthereumNodeAddress = "ethereum-sidecar.server.ethereum-node-address"
)

func NewFlagSetEthereumSidecar(
	defaultServerAddress,
	defaultServerEthereumNodeAddress string,
) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(
		FlagServerAddress,
		defaultServerAddress,
		"The ethereum sidecar gRPC address (e.g. 127.0.0.1:5051)",
	)
	fs.String(
		FlagServerEthereumNodeAddress,
		defaultServerEthereumNodeAddress,
		"The ethereum node address (e.g. ws://127.0.0.1:8546)",
	)

	return fs
}
