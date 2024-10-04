package cli

import (
	"context"
	"fmt"

	"google.golang.org/grpc/encoding"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/mezo-org/mezod/ethereum/sidecar"
	"github.com/spf13/cobra"

	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
)

func NewEthereumSidecarCmd() *cobra.Command {
	defaultServerAddress := "0.0.0.0:7500"
	defaultServerEthereumNodeAddress := "ws://127.0.0.1:8546"
	defaultServerEthereumNetwork := ethconfig.Sepolia

	cmd := &cobra.Command{
		Use:   "ethereum-sidecar",
		Short: "Starts Ethereum sidecar",
		Long: "Starts the Ethereum sidecar which observes the assets locked " +
			"in the Mezo bridge on the Ethereum chain",
		RunE: runEthereumSidecar,
	}

	cmd.Flags().AddFlagSet(
		NewFlagSetEthereumSidecar(
			defaultServerAddress,
			defaultServerEthereumNodeAddress,
			defaultServerEthereumNetwork.String(),
		))

	return cmd
}

func runEthereumSidecar(cmd *cobra.Command, _ []string) error {
	logger := server.GetServerContextFromCmd(cmd).Logger.With(
		"module",
		"ethereum-sidecar",
	)

	grpcAddress, _ := cmd.Flags().GetString(FlagServerAddress)
	ethNodeAddress, _ := cmd.Flags().GetString(FlagServerEthereumNodeAddress)
	network, _ := cmd.Flags().GetString(FlagServerNetwork)

	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// The messages handled by the server contain custom types. Add codecs so
	// that the messages can be marshaled/unmarshalled.
	encoding.RegisterCodec(
		codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec(),
	)

	ctx := context.Background()
	sidecar.RunServer(ctx, grpcAddress, ethNodeAddress, network, logger)
	<-ctx.Done()

	return fmt.Errorf("unexpected context cancellation")
}
