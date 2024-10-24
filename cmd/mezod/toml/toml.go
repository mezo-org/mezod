package toml

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toml",
		Short: "Utilities for setting/getting parts of configuration",
	}

	cmd.AddCommand(
		NewSetTOML(),
	)

	return cmd
}
