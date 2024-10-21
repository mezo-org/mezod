// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mezo-org/mezod/chain"

	"github.com/mezo-org/mezod/types"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const flagIgnorePredefined = "ignore-predefined"

type printInfo struct {
	Moniker     string `json:"moniker" yaml:"moniker"`
	ChainID     string `json:"chain_id" yaml:"chain_id"`
	NodeID      string `json:"node_id" yaml:"node_id"`
	GenesisTime string `json:"genesis_time" yaml:"genesis_time"`
}

func newPrintInfo(moniker, chainID, nodeID, genesisTime string) printInfo {
	return printInfo{
		Moniker:     moniker,
		ChainID:     chainID,
		NodeID:      nodeID,
		GenesisTime: genesisTime,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(os.Stderr, "%s\n", string(sdk.MustSortJSON(out))); err != nil {
		return err
	}

	return nil
}

const InitCmdLong =
	`Initialize the node's home directory with required files. ` +
	`Specifically, this command initializes the following: ` + "\n" +
	`- The genesis file for the chain (genesis.json). ` + "\n" +
	`  By default, this file is taken from a predefined chain config, if such a config exists for the given chain id. ` + "\n" +
	`  Otherwise, a default genesis file will be generated. ` + "\n" +
	`  To ignore an existing predefined chain config and always generate a default genesis file, use the --ignore-predefined flag. ` + "\n" +
	`  If the genesis file already exists in the home directory, this command will fail. ` + "\n" +
	`  To overwrite an existing genesis file, use the --overwrite flag. ` + "\n" +
	`- Configuration files (app.toml, client.toml, config.toml) ` + "\n" +
	`- Validator key (priv_validator_key.json). ` + "\n" +
	`  To recover an existing key from a seed phrase, use the --recover flag. ` + "\n" +
	`- Node peer-to-peer key (node_key.json)`

// NewInitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func NewInitCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init MONIKER",
		Short: "Initialize the node's home directory with required files",
		Long:  InitCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if !types.IsValidChainID(chainID) {
				return fmt.Errorf("invalid chain-id format: %s", chainID)
			}

			var predefinedChainConfig chain.Config
			if ignorePredefined, _ := cmd.Flags().GetBool(flagIgnorePredefined); !ignorePredefined {
				var err error
				predefinedChainConfig, err = chain.LoadConfig(chainID)
				if err != nil {
					return fmt.Errorf("failed to load predefined config for chain: %w", err)
				}
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			// Set peers in and out to an 8:1 ratio to prevent choking
			config.P2P.MaxNumInboundPeers = 240
			config.P2P.MaxNumOutboundPeers = 30

			// Set default seeds
			var seeds []string
			if predefinedChainConfig.Exists() {
				seeds = predefinedChainConfig.Seeds
			}
			config.P2P.Seeds = strings.Join(seeds, ",")

			config.Mempool.Size = 10000
			config.StateSync.TrustPeriod = 112 * time.Hour

			config.SetRoot(clientCtx.HomeDir)

			// Get bip39 mnemonic
			var mnemonic string

			recoverKey, _ := cmd.Flags().GetBool(genutilcli.FlagRecover)
			if recoverKey {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				mnemonic = value
				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			genFile := config.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(genutilcli.FlagOverwrite)

			if !overwrite && tmos.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			var appGenesis *genutiltypes.AppGenesis
			if predefinedChainConfig.Exists() {
				appGenesis = predefinedChainConfig.Genesis
			} else {
				appState, err := json.MarshalIndent(
					mbm.DefaultGenesis(cdc),
					"",
					" ",
				)
				if err != nil {
					return errors.Wrap(
						err,
						"Failed to marshall default genesis state",
					)
				}

				appGenesis = &genutiltypes.AppGenesis{}
				if _, err := os.Stat(genFile); err != nil {
					if !os.IsNotExist(err) {
						return err
					}
				} else {
					appGenesis, err = genutiltypes.AppGenesisFromFile(genFile)
					if err != nil {
						return errors.Wrap(
							err,
							"Failed to read genesis doc from file",
						)
					}
				}

				appGenesis.ChainID = chainID

				appGenesis.Consensus = &genutiltypes.ConsensusGenesis{
					Validators: nil,
					Params:     tmtypes.DefaultConsensusParams(),
				}
				// Set the block gas limit to 10M.
				appGenesis.Consensus.Params.Block.MaxGas = 10_000_000
				// Enable vote extensions from block 1.
				appGenesis.Consensus.Params.ABCI.VoteExtensionsEnableHeight = 1

				appGenesis.AppState = appState
			}

			if err := genutil.ExportGenesisFile(
				appGenesis,
				genFile,
			); err != nil {
				return errors.Wrap(err, "Failed to export gensis file")
			}

			toPrint := newPrintInfo(
				config.Moniker,
				chainID,
				nodeID,
				appGenesis.GenesisTime.String(),
			)

			cfg.WriteConfigFile(
				filepath.Join(
					config.RootDir,
					"config",
					"config.toml",
				), config,
			)
			return displayInfo(toPrint)
		},
	}

	cmd.Flags().Bool(
		flagIgnorePredefined,
		false,
		"Ignore predefined config for the chain and initialize a fresh genesis file",
	)
	cmd.Flags().BoolP(
		genutilcli.FlagOverwrite,
		"o",
		false,
		"Overwrite the genesis file already existing in the home directory",
	)
	cmd.Flags().Bool(
		genutilcli.FlagRecover,
		false,
		"Provide seed to recover validator key instead of creating",
	)

	return cmd
}
