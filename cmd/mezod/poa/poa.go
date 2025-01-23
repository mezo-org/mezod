package poa

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poa",
		Short: "Tool for PoA network",
	}
	cmd.AddCommand(
		NewSubmitApplicationCmd(),
	)
	return cmd
}
