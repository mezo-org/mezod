package genesis

import (
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/mezo-org/mezod/app"
	poacli "github.com/mezo-org/mezod/x/poa/client/cli"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis",
		Short: "Utilities for chain bootstrapping",
	}

	cmd.AddCommand(
		NewAddAccountCmd(),
		NewMigrateCmd(),
		poacli.NewGenValCmd(),
		poacli.NewCollectGenValsCmd(),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
	)

	return cmd
}
