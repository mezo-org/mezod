package sidecar

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func NewEthereumSidecarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ethereum-sidecar",
		Short: "Starts Ethereum sidecar",
		Long: "Starts the Ethereum sidecar which observes the assets locked " +
			"in the Mezo bridge on the Ethereum chain",
		RunE: ethereumSidecar,
	}

	return cmd
}

func ethereumSidecar(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	RunServer(ctx)

	<-ctx.Done()

	return fmt.Errorf("unexpected context cancellation")
}
