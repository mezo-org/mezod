package localnet

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func NewCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "localnet",
		Short:                      "Utilities for localnet",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewInitFilesCmd(mbm))

	return cmd
}
