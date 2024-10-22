package main_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/mezo-org/mezod/app"
	mezod "github.com/mezo-org/mezod/cmd/mezod"
	"github.com/mezo-org/mezod/utils"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := mezod.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",      // Test the init cmd
		"mezo-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
		fmt.Sprintf("--%s=%s", flags.FlagChainID, utils.TestnetChainID+"-1"),
	})

	err := svrcmd.Execute(rootCmd, mezod.EnvPrefix, app.DefaultNodeHome)
	require.NoError(t, err)
}

func TestAddKeyLedgerCmd(t *testing.T) {
	rootCmd, _ := mezod.NewRootCmd()
	rootCmd.SetArgs([]string{
		"keys",
		"add",
		"dev0",
		fmt.Sprintf("--%s", flags.FlagUseLedger),
	})

	require.Panics(t, func() {
		_ = svrcmd.Execute(rootCmd, mezod.EnvPrefix, app.DefaultNodeHome)
	})
}
