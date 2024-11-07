package toml

import (
	"os"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toml",
		Short: "Utilities for setting/getting values in TOML configuration files",
	}

	cmd.AddCommand(
		NewSetCmd(),
		NewGetCmd(),
	)

	return cmd
}

type keyValueTOML struct {
	Key   []string
	Value string
}

func isFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}
