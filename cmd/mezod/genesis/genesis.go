package genesis

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis",
		Short: "Utilities for chain bootstrapping",
	}

	cmd.AddCommand(
		NewAddAccountCmd(),
		NewMigrateCmd(),
	)

	return cmd
}
