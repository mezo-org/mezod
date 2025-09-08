package bridgeworker

import (
	"fmt"
	"os"

	"cosmossdk.io/log"

	"github.com/ethereum/go-ethereum/crypto"
	bwconfig "github.com/mezo-org/mezod/bridge-worker/config"
)

func Start(configPath string) {
	logger := log.NewLogger(os.Stdout).With(log.ModuleKey, "bridge-worker")

	cfg, err := bwconfig.ReadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("config error: %v", err))
	}

	privateKey, err := bwconfig.DecryptKeyFile(
		cfg.EthereumAccount.KeyFile,
		cfg.EthereumAccount.KeyFilePassword,
	)
	if err != nil {
		panic(fmt.Sprintf("keyfile load error: %v", err))
	}

	accountAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	logger.Info(
		"loaded Ethereum private key",
		"account_address", accountAddress,
	)

	RunBridgeWorker(
		logger,
		cfg.ProviderURL,
		cfg.EthereumNetwork,
		cfg.BatchSize,
		cfg.RequestsPerMinute,
		privateKey,
	)
}
