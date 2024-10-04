package cli

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

const (
	FlagServerAddress             = "ethereum-sidecar.server.address"
	FlagServerEthereumNodeAddress = "ethereum-sidecar.server.ethereum-node-address"
	FlagServerNetwork             = "ethereum-sidecar.server.network"
)

func NewFlagSetEthereumSidecar(
	defaultServerAddress,
	defaultServerEthereumNodeAddress,
	defualtServerNetwork string,
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
		defualtServerNetwork,
		fmt.Sprintf(
			"The sidecar server Ethereum network (e.g. %s)",
			defualtServerNetwork,
		),
	)

	return fs
}
