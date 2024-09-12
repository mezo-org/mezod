package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagGRPCAddress    = "ethereum-sidecar.server.grpc-address"
	FlagEthNodeAddress = "ethereum-sidecar.server.ethereum-node-address"
)

func NewFlagSetEthereumSidecar(
	defaultAddress,
	defaultEthNodeAddress string,
) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(
		FlagGRPCAddress,
		defaultAddress,
		"The ethereum sidecar gRPC address (e.g. 127.0.0.1:5051)",
	)
	fs.String(
		FlagEthNodeAddress,
		defaultEthNodeAddress,
		"The ethereum node address (e.g. ws://127.0.0.1:8546)",
	)

	return fs
}
