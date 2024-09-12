package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/mezo-org/mezod/ethereum/sidecar"
	"github.com/spf13/cobra"
)

func NewEthereumSidecarCmd() *cobra.Command {
	defaultGRPCIP, _ := server.ExternalIP()
	defaultGRPCAddress := fmt.Sprintf("%s:50051", defaultGRPCIP)

	defaultEthNodeAddress := "ws://127.0.0.1:8546"

	cmd := &cobra.Command{
		Use:   "ethereum-sidecar",
		Short: "Starts Ethereum sidecar",
		Long: "Starts the Ethereum sidecar which observes the assets locked " +
			"in the Mezo bridge on the Ethereum chain",
		RunE: runEthereumSidecar,
	}

	cmd.Flags().AddFlagSet(
		NewFlagSetEthereumSidecar(
			defaultGRPCAddress,
			defaultEthNodeAddress,
		))

	return cmd
}

func runEthereumSidecar(cmd *cobra.Command, _ []string) error {
	grpcAddress, _ := cmd.Flags().GetString(FlagGRPCAddress)
	ethNodeAddress, _ := cmd.Flags().GetString(FlagEthNodeAddress)

	ctx := context.Background()
	sidecar.RunServer(ctx, grpcAddress, ethNodeAddress)
	<-ctx.Done()

	return fmt.Errorf("unexpected context cancellation")
}
