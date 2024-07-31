package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"

	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewGenValCmd creates the Cobra command to generate a new validator.
func NewGenValCmd(defaultHome string) *cobra.Command {
	defaultIP, _ := server.ExternalIP()

	cmd := &cobra.Command{
		Use:   "genval [key_name]",
		Short: "Generate data for a new validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(serverCtx.Config)
			if err != nil {
				return errors.Wrap(err, "failed to initialize node validator files")
			}

			name := args[0]
			key, err := clientCtx.Keyring.Key(name)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch '%s' from the keyring", name)
			}
			address, err := key.GetAddress()
			if err != nil {
				return err
			}

			moniker := config.Moniker
			if monikerFlag, _ := cmd.Flags().GetString(FlagMoniker); monikerFlag != "" {
				moniker = monikerFlag
			}
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			securityContact, _ := cmd.Flags().GetString(FlagSecurityContact)
			details, _ := cmd.Flags().GetString(FlagDetails)

			ip, _ := cmd.Flags().GetString(FlagIP)
			p2pPort, _ := cmd.Flags().GetString(FlagP2PPort)

			validator, err := poatypes.NewValidator(
				sdk.ValAddress(address),
				valPubKey,
				poatypes.Description{
					Moniker:         moniker,
					Identity:        identity,
					Website:         website,
					SecurityContact: securityContact,
					Details:         details,
				},
			)
			if err != nil {
				return errors.Wrap(err, "failed to create validator")
			}

			outDocContent := map[string]interface{}{
				"validator": validator,
				"memo":      fmt.Sprintf("%s@%s:%s", nodeID, ip, p2pPort),
			}

			outputDocument, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
			if outputDocument == "" {
				outputDocument, err = makeOutputFilepath(config.RootDir, nodeID)
				if err != nil {
					return errors.Wrap(err, "failed to create output file path")
				}
			}

			err = writeOutputDocument(outputDocument, outDocContent)
			if err != nil {
				return errors.Wrap(err, "failed to write output file")
			}

			cmd.Printf("Validator data written to %q\n", outputDocument)

			return nil
		},
	}

	cmd.Flags().AddFlagSet(NewFlagSetGenVal(defaultHome, defaultIP, "26656"))

	return cmd
}

// NewCollectGenValsCmd creates the Cobra command to collect generated validators.
func NewCollectGenValsCmd(defaultHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect-genvals",
		Short: "Collect generated validators and output a genesis.json file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			clientCtx := client.GetClientContextFromCmd(cmd)

			config.SetRoot(clientCtx.HomeDir)

			genValDir, _ := cmd.Flags().GetString(FlagGenValDir)
			if genValDir == "" {
				genValDir = defaultGenValDir(config.RootDir)
			}

			genvals, err := os.ReadDir(genValDir)
			if err != nil {
				return errors.Wrap(err, "cannot read genval dir")
			}

			validators := make([]poatypes.Validator, 0)
			persistentPeers := make([]string, 0)

			for _, genval := range genvals {
				if genval.IsDir() {
					continue
				}
				if !strings.HasSuffix(genval.Name(), ".json") {
					continue
				}

				contentBytes, err := os.ReadFile(
					filepath.Join(
						genValDir,
						genval.Name(),
					),
				)
				if err != nil {
					return errors.Wrap(err, "cannot read genval file")
				}

				content := struct {
					Validator poatypes.Validator
					Memo      string
				}{}
				err = json.Unmarshal(contentBytes, &content)
				if err != nil {
					return errors.Wrap(err, "cannot parse genval file")
				}

				validators = append(validators, content.Validator)
				persistentPeers = append(persistentPeers, content.Memo)
			}

			if len(validators) == 0 {
				cmd.Println("No validators found in genval dir")
				return nil
			}

			// Read the global genesis file.
			appGenesis, err := types.AppGenesisFromFile(config.GenesisFile())
			if err != nil {
				return errors.Wrap(err, "failed to read app genesis from file")
			}
			// Get the app state (all modules) from global genesis.
			appState, err := types.GenesisStateFromAppGenesis(appGenesis)
			if err != nil {
				return errors.Wrap(err, "failed to create genesis state from gen doc")
			}
			// Get state of the x/poa module.
			poaState := getModuleStateFromAppState(clientCtx.Codec, appState)
			// Set initial validators for the x/poa module.
			poaState.Validators = validators
			// Propagate changes in the x/poa module state back to the app state.
			appState = setModuleStateInAppState(clientCtx.Codec, appState, poaState)
			// Marshal the app state.
			appStateBytes, err := json.MarshalIndent(appState, "", "  ")
			if err != nil {
				return errors.Wrap(err, "failed to marshal app state")
			}
			// Set the updated app state in the global genesis file.
			appGenesis.AppState = appStateBytes
			// Export the updated global genesis file.
			err = genutil.ExportGenesisFile(appGenesis, config.GenesisFile())
			if err != nil {
				return errors.Wrap(err, "failed to export genesis file")
			}
			// Set initial validators as persistent files in the node configuration.
			sort.Strings(persistentPeers)
			config.P2P.PersistentPeers = strings.Join(persistentPeers, ",")
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			return nil
		},
	}

	cmd.Flags().AddFlagSet(NewFlagSetCollectGenVals(defaultHome))

	return cmd
}

func getModuleStateFromAppState(
	cdc codec.JSONCodec,
	appState map[string]json.RawMessage,
) *poatypes.GenesisState {
	var moduleState poatypes.GenesisState
	if content := appState[poatypes.ModuleName]; content != nil {
		cdc.MustUnmarshalJSON(content, &moduleState)
	}
	return &moduleState
}

func setModuleStateInAppState(
	cdc codec.JSONCodec,
	appState map[string]json.RawMessage,
	moduleState *poatypes.GenesisState,
) map[string]json.RawMessage {
	content := cdc.MustMarshalJSON(moduleState)
	appState[poatypes.ModuleName] = content
	return appState
}

func defaultGenValDir(rootDir string) string {
	return filepath.Join(rootDir, "config", "genval")
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := defaultGenValDir(rootDir)
	if err := tmos.EnsureDir(writePath, 0o700); err != nil {
		return "", err
	}

	return filepath.Join(writePath, fmt.Sprintf("genval-%v.json", nodeID)), nil
}

func writeOutputDocument(outputDocument string, content interface{}) error {
	outputFile, err := os.OpenFile(
		outputDocument,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = outputFile.Close()
	}()

	contentBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(outputFile, "%s\n", contentBytes)

	return err
}
