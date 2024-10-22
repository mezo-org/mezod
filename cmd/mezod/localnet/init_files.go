package localnet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	tmconfig "github.com/cometbft/cometbft/config"
	cometos "github.com/cometbft/cometbft/libs/os"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ethereum/go-ethereum/common"
	cmdcfg "github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/crypto/hd"
	mezokr "github.com/mezo-org/mezod/crypto/keyring"
	"github.com/mezo-org/mezod/server/config"
	mezotypes "github.com/mezo-org/mezod/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	poatypes "github.com/mezo-org/mezod/x/poa/types"
	"github.com/spf13/cobra"
)

var (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "v"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagStartingIPAddress = "starting-ip-address"
)

type initArgs struct {
	algo              string
	chainID           string
	keyringBackend    string
	minGasPrices      string
	nodeDaemonHome    string
	nodeDirPrefix     string
	numValidators     int
	outputDir         string
	startingIPAddress string
}

const InitFilesCmdLong = `Initialize config directories and files for a multi-validator localnet. ` + "\n" +
	`This command sets up the given number of directories and populate each with ` +
	`necessary files (validator key, node key, genesis, configuration).` + "\n" +
	`Resulting directories can be used to start a localnet based either on the native binary or Docker.`

const InitFilesCmdExample = "init-files --v 4 --output-dir ./.localnet --starting-ip-address 192.168.10.2"

func NewInitFilesCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init-files",
		Short:   "Initialize config directories and files for a multi-validator localnet",
		Long:    InitFilesCmdLong,
		Example: InitFilesCmdExample,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := sdkserver.GetServerContextFromCmd(cmd)

			args := initArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.keyringBackend, _ = cmd.Flags().GetString(flags.FlagKeyringBackend)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(sdkserver.FlagMinGasPrices)
			args.nodeDirPrefix, _ = cmd.Flags().GetString(flagNodeDirPrefix)
			args.nodeDaemonHome, _ = cmd.Flags().GetString(flagNodeDaemonHome)
			args.startingIPAddress, _ = cmd.Flags().GetString(flagStartingIPAddress)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyType)

			return initLocalnetFiles(clientCtx, cmd, serverCtx.Config, mbm, args)
		},
	}

	cmd.Flags().Int(flagNumValidators, 4, "Number of validators to initialize the localnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./.localnet", "Directory to store initialization data for the localnet")
	cmd.Flags().String(sdkserver.FlagMinGasPrices, fmt.Sprintf("0.000006%s", cmdcfg.BaseDenom), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01abtc)")
	cmd.Flags().String(flags.FlagKeyType, string(hd.EthSecp256k1Type), "Key signing algorithm to generate keys for")
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "mezod", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, etc.). Setting this flag to localhost results in configuration generation for binary-based localnet")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")

	return cmd
}

const (
	nodeDirPerm = 0o755
	localhost   = "localhost"
)

// initLocalnetFiles initializes localnet files for a localnet to be run in a separate process
func initLocalnetFiles(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	args initArgs,
) error {
	if args.chainID == "" {
		args.chainID = fmt.Sprintf("mezo_%d-1", tmrand.Int63n(9999999999999)+1)
	}

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
	)

	genFiles := make([]string, args.numValidators)
	nodeIDs := make([]string, args.numValidators)
	valPubKeys := make([]cryptotypes.PubKey, args.numValidators)
	validators := make([]poatypes.Validator, args.numValidators)
	memos := make([]string, args.numValidators)
	nodeDirNames := make([]string, args.numValidators)
	nodeDirs := make([]string, args.numValidators)

	inBuf := bufio.NewReader(cmd.InOrStdin())

	for i := 0; i < args.numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", args.nodeDirPrefix, i)
		nodeDirNames[i] = nodeDirName

		nodeDir := filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
		nodeDirs[i] = nodeDir

		nodeConfig.SetRoot(nodeDir)

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		var err error
		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(nodeConfig)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		memo, err := getMemo(nodeIDs[i], i, args.startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		memos[i] = memo

		genFiles[i] = nodeConfig.GenesisFile()

		kr, err := keyring.New(
			sdk.KeyringServiceName(),
			args.keyringBackend,
			nodeDir,
			inBuf,
			clientCtx.Codec,
			mezokr.Option(),
		)
		if err != nil {
			return err
		}

		keyringAlgos, _ := kr.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(args.algo, keyringAlgos)
		if err != nil {
			return err
		}

		address, secret, err := testutil.GenerateSaveCoinKey(
			kr,
			nodeDirName,
			"",
			true,
			algo,
		)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		info := map[string]string{"secret": secret}
		infoBytes, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err := writeFile(
			fmt.Sprintf("%v.json", "key_seed"),
			nodeDir,
			infoBytes,
		); err != nil {
			return err
		}

		balance, _ := sdkmath.NewIntFromString("100000000000000000000000000")
		coins := sdk.NewCoins(sdk.NewCoin(cmdcfg.BaseDenom, balance))

		genBalances = append(
			genBalances,
			banktypes.Balance{Address: address.String(), Coins: coins.Sort()},
		)
		genAccounts = append(
			genAccounts, &mezotypes.EthAccount{
				BaseAccount: authtypes.NewBaseAccount(address, nil, 0, 0),
				CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).Hex(),
			},
		)

		validator, err := poatypes.NewValidator(
			sdk.ValAddress(address),
			valPubKeys[i],
			poatypes.Description{
				Moniker: nodeDirName,
			},
		)
		if err != nil {
			return err
		}

		validators[i] = validator
	}

	// Build persistent peers list based on validators' memos.
	sort.Strings(memos)
	persistentPeers := strings.Join(memos, ",")

	// Initialize config files for each validator node.
	for i := range validators {
		nodeDir := nodeDirs[i]

		nodeConfig.SetRoot(nodeDir)
		nodeConfig.Moniker = nodeDirNames[i]
		nodeConfig.RPC.ListenAddress = getRPCAddress(i, args.startingIPAddress)
		nodeConfig.P2P.ListenAddress = getP2PAddress(i, args.startingIPAddress)
		nodeConfig.RPC.PprofListenAddress = getPprofAddress(i, args.startingIPAddress)
		nodeConfig.P2P.PersistentPeers = persistentPeers
		nodeConfig.P2P.AddrBookStrict = getAddrBookStrict(args.startingIPAddress)
		nodeConfig.P2P.AllowDuplicateIP = getAllowDuplicateIP(args.startingIPAddress)
		tmconfig.WriteConfigFile(filepath.Join(nodeDir, "config/config.toml"), nodeConfig)

		appConfig := config.DefaultConfig()
		appConfig.MinGasPrices = args.minGasPrices
		appConfig.API.Enable = true
		appConfig.API.Address = getAPIAddress(i, args.startingIPAddress)
		appConfig.GRPC.Address = getGRPCAddress(i, args.startingIPAddress)
		appConfig.Telemetry.Enabled = true
		appConfig.Telemetry.PrometheusRetentionTime = 60
		appConfig.Telemetry.EnableHostnameLabel = false
		appConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", args.chainID}}
		appConfig.JSONRPC.Address = getJSONRPCAddress(i, args.startingIPAddress)
		appConfig.JSONRPC.WsAddress = getJSONRPCWsAddress(i, args.startingIPAddress)
		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), appConfig)
	}

	if err := initGenesisFiles(
		clientCtx,
		mbm,
		args.chainID,
		cmdcfg.BaseDenom,
		genAccounts,
		genBalances,
		genFiles,
		validators,
	); err != nil {
		return err
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", args.numValidators)

	return nil
}

func initGenesisFiles(
	clientCtx client.Context,
	mbm module.BasicManager,
	chainID,
	coinDenom string,
	genAccounts []authtypes.GenesisAccount,
	genBalances []banktypes.Balance,
	genFiles []string,
	validators []poatypes.Validator,
) error {
	appGenState := mbm.DefaultGenesis(clientCtx.Codec)
	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}
	authGenState.Accounts = accounts
	appGenState[authtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)
	bankGenState.Balances = genBalances
	appGenState[banktypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&bankGenState)

	var crisisGenState crisistypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[crisistypes.ModuleName], &crisisGenState)
	crisisGenState.ConstantFee.Denom = coinDenom
	appGenState[crisistypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&crisisGenState)

	var evmGenState evmtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[evmtypes.ModuleName], &evmGenState)
	evmGenState.Params.EvmDenom = coinDenom
	appGenState[evmtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&evmGenState)

	var poaGenState poatypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[poatypes.ModuleName], &poaGenState)
	// Set the first validator as the initial owner.
	poaGenState.Owner = sdk.AccAddress(validators[0].GetOperator()).String()
	poaGenState.Validators = validators
	appGenState[poatypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&poaGenState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	// generate genesis files for each validator and save
	genTime := tmtime.Now()
	for i := range validators {
		genFile := genFiles[i]

		// Create the genesis.json file.
		if err := genutil.ExportGenesisFileWithTime(
			genFile,
			chainID,
			nil,
			appGenStateJSON,
			genTime,
		); err != nil {
			return err
		}

		// Load the genesis.json file and update to update the consensus params.
		appGenesis, err := types.AppGenesisFromFile(genFile)
		if err != nil {
			return err
		}
		// Set the block gas limit to 10M.
		appGenesis.Consensus.Params.Block.MaxGas = 10_000_000
		// Enable vote extensions from block 1.
		appGenesis.Consensus.Params.ABCI.VoteExtensionsEnableHeight = 1
		// Export the updated genesis file.
		if err := genutil.ExportGenesisFile(appGenesis, genFile); err != nil {
			return err
		}
	}
	return nil
}

func getMemo(nodeID string, i int, startingIPAddr string) (string, error) {
	if startingIPAddr == localhost {
		return fmt.Sprintf("%s@localhost:%d", nodeID, 26656+2*i), nil
	}

	ip, err := getIP(i, startingIPAddr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s@%s:26656", nodeID, ip), nil
}

func getRPCAddress(i int, startingIPAddr string) string {
	port := 26657

	if startingIPAddr == localhost {
		port += 2 * i
	}

	return fmt.Sprintf("tcp://0.0.0.0:%d", port)
}

func getP2PAddress(i int, startingIPAddr string) string {
	port := 26656

	if startingIPAddr == localhost {
		port += 2 * i
	}

	return fmt.Sprintf("tcp://0.0.0.0:%d", port)
}

func getPprofAddress(i int, startingIPAddr string) string {
	port := 6060

	if startingIPAddr == localhost {
		port += i
	}

	return fmt.Sprintf("localhost:%d", port)
}

func getAPIAddress(i int, startingIPAddr string) string {
	port := 1317

	if startingIPAddr == localhost {
		port += i
	}

	return fmt.Sprintf("tcp://0.0.0.0:%d", port)
}

func getGRPCAddress(i int, startingIPAddr string) string {
	port := 9090

	if startingIPAddr == localhost {
		port += 2 * i
	}

	return fmt.Sprintf("0.0.0.0:%d", port)
}

func getJSONRPCAddress(i int, startingIPAddr string) string {
	port := 8545

	if startingIPAddr == localhost {
		port += 2 * i
	}

	return fmt.Sprintf("0.0.0.0:%d", port)
}

func getJSONRPCWsAddress(i int, startingIPAddr string) string {
	port := 8546

	if startingIPAddr == localhost {
		port += 2 * i
	}

	return fmt.Sprintf("0.0.0.0:%d", port)
}

func getAddrBookStrict(startingIPAddr string) bool {
	return startingIPAddr != localhost
}

func getAllowDuplicateIP(startingIPAddr string) bool {
	return startingIPAddr == localhost
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = sdkserver.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	err := cometos.EnsureDir(dir, 0o755)
	if err != nil {
		return err
	}

	return cometos.WriteFile(file, contents, 0o644)
}
