package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"
)

var (
	// available for all commands
	rpcAddress     string
	mnemonic       string
	localKey       string
	privKey        string
	cnt            uint
	waitForReceipt bool
	accountsStore  string
	rateLimit      time.Duration
	tendermintRPC  string

	// aggregation only
	fromBlock uint
	toBlock   uint
	output    string

	// erc20 run
	tokenAddress string

	// perf runs only
	duration time.Duration

	generateCmd           = flag.NewFlagSet("generate", flag.ExitOnError)
	deployTokenCmd        = flag.NewFlagSet("deploy_token", flag.ExitOnError)
	topupCmd              = flag.NewFlagSet("topup", flag.ExitOnError)
	topupERC20Cmd         = flag.NewFlagSet("topup_erc20", flag.ExitOnError)
	runNativeCmd          = flag.NewFlagSet("run_native", flag.ExitOnError)
	runErc20PrecompileCmd = flag.NewFlagSet("run_erc20_precompile", flag.ExitOnError)
	runErc20Cmd           = flag.NewFlagSet("run_erc20", flag.ExitOnError)
	aggregateCmd          = flag.NewFlagSet("aggregate", flag.ExitOnError)
	helpCmd               = flag.NewFlagSet("help", flag.ExitOnError)
)

var subcommands = map[string]*flag.FlagSet{
	deployTokenCmd.Name():        deployTokenCmd,
	generateCmd.Name():           generateCmd,
	topupCmd.Name():              topupCmd,
	topupERC20Cmd.Name():         topupERC20Cmd,
	runNativeCmd.Name():          runNativeCmd,
	runErc20PrecompileCmd.Name(): runErc20PrecompileCmd,
	runErc20Cmd.Name():           runErc20Cmd,
	runNativeCmd.Name():          runNativeCmd,
	aggregateCmd.Name():          aggregateCmd,
	helpCmd.Name():               helpCmd,
	"-h":                         helpCmd,
	"-help":                      helpCmd,
	"--help":                     helpCmd,
}

var (
	allSubcommands = map[string]*flag.FlagSet{
		generateCmd.Name():           generateCmd,
		topupCmd.Name():              topupCmd,
		topupERC20Cmd.Name():         topupERC20Cmd,
		runNativeCmd.Name():          runNativeCmd,
		runErc20PrecompileCmd.Name(): runErc20PrecompileCmd,
		runErc20Cmd.Name():           runErc20Cmd,
		aggregateCmd.Name():          aggregateCmd,
		deployTokenCmd.Name():        deployTokenCmd,
		helpCmd.Name():               helpCmd,
	}

	transactSubCommands = map[string]*flag.FlagSet{
		deployTokenCmd.Name():        deployTokenCmd,
		generateCmd.Name():           generateCmd,
		topupCmd.Name():              topupCmd,
		topupERC20Cmd.Name():         topupERC20Cmd,
		runNativeCmd.Name():          runNativeCmd,
		runErc20PrecompileCmd.Name(): runErc20PrecompileCmd,
		runErc20Cmd.Name():           runErc20Cmd,
	}

	walletSubCommands = map[string]*flag.FlagSet{
		generateCmd.Name():           generateCmd,
		topupCmd.Name():              topupCmd,
		topupERC20Cmd.Name():         topupERC20Cmd,
		runNativeCmd.Name():          runNativeCmd,
		runErc20PrecompileCmd.Name(): runErc20PrecompileCmd,
		runErc20Cmd.Name():           runErc20Cmd,
	}

	runSubCommands = map[string]*flag.FlagSet{
		runNativeCmd.Name():          runNativeCmd,
		runErc20PrecompileCmd.Name(): runErc20PrecompileCmd,
		runErc20Cmd.Name():           runErc20Cmd,
	}
)

func commonFlags() {
	// flags for all commands
	for _, fs := range allSubcommands {
		fs.StringVar(&rpcAddress, "rpc_address", "http://0.0.0.0:8545", "ethereum RPC address")
		fs.StringVar(&accountsStore, "accounts", "accounts.json", "file to store generated accounts")
	}

	// flags for all commands which transact with the network
	for _, fs := range transactSubCommands {
		fs.StringVar(&localKey, "localkey", "", "path to a local key to use (testnet)")
		fs.StringVar(&mnemonic, "mnemonic", "", "mnemonic of the wallet")
		fs.StringVar(&privKey, "privkey", "", "the private key to use")
		fs.BoolVar(&waitForReceipt, "wait_for_receipt", true, "wait for the mezo receipt")
		fs.DurationVar(&rateLimit, "rate", 100*time.Millisecond, "rate at which to send transactions")
	}

	// flags for all commands with wallets
	for _, fs := range walletSubCommands {
		fs.UintVar(&cnt, "count", 10, "number of private keys to use")
		fs.StringVar(&tendermintRPC, "tendermint_rpc_address", "http://0.0.0.0:26657", "tendermint RPC address")
		fs.DurationVar(&duration, "duration", time.Minute*1, "how long the run should last")
	}

	// individual flags
	runErc20Cmd.StringVar(&tokenAddress, "address", "", "the erc20 token address")
	runErc20PrecompileCmd.StringVar(&tokenAddress, "address", "", "the erc20 token address")
	topupERC20Cmd.StringVar(&tokenAddress, "address", "", "the erc20 token address")
	aggregateCmd.StringVar(&output, "output", "output.csv", "output for the aggregate")
	aggregateCmd.UintVar(&fromBlock, "from", 0, "from block")
	aggregateCmd.UintVar(&toBlock, "to", 0, "to block")
}

func printHelp() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "\nAvailable commands:\n")
	commands := maps.Keys(allSubcommands)
	slices.Sort(commands)

	for _, v := range commands {
		fmt.Fprintf(flag.CommandLine.Output(), "  %v\n", subcommands[v].Name())
	}
}

func runCLI() {
	commonFlags()

	if len(os.Args) <= 1 {
		printHelp()
		return
	}

	cmd := subcommands[os.Args[1]]
	if cmd == nil {
		log.Fatalf("error: unknown subcommand '%s', see help for more details.", os.Args[1])
	}

	cmd.Parse(os.Args[2:])

	if cmd.Name() != "help" {
		setupMainAccount()
	}

	switch cmd.Name() {
	case deployTokenCmd.Name():
		deployToken()
	case generateCmd.Name():
		generateAndTransferERC20Precompile(cnt)
	case runNativeCmd.Name():
		_, _, mainAddress := getAccount(mainAccount)
		runNative(int(cnt), duration, mainAddress)
	case runErc20Cmd.Name():
		_, _, mainAddress := getAccount(mainAccount)
		runERC20(int(cnt), duration, mainAddress, common.HexToAddress(tokenAddress))
	case runErc20PrecompileCmd.Name():
		_, _, mainAddress := getAccount(mainAccount)
		runERC20(int(cnt), duration, mainAddress, btcTokenAddress)
	case topupCmd.Name():
		topupERC20Precompile(int(cnt))
	case topupERC20Cmd.Name():
		topupERC20(int(cnt), common.HexToAddress(tokenAddress))
	case aggregateCmd.Name():
		aggregateBlockData()
	case helpCmd.Name():
		printHelp()
	}
}

func setupMainAccount() {
	fmt.Printf("loading main account\n")
	if len(localKey) > 0 {
		fmt.Printf("loading main account from local key %v\n", localKey)
		mainAccount = importKey(localKey)
	} else if len(mnemonic) > 0 {
		fmt.Printf("loading main account from mnemonic %v\n", mnemonic)
		mainAccount = importMnemonic(mnemonic)
	} else if len(privKey) > 0 {
		fmt.Printf("loading main account from private key %v\n", privKey)
		mainAccount = privKey
		_, _, address := getAccount(privKey)
		fmt.Printf("loaded private key: %v", address)
	}
}
