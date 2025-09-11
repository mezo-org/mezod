package bridgeworker

import (
	"fmt"
	"os"

	"cosmossdk.io/log"
	"github.com/rs/zerolog"

	"github.com/ethereum/go-ethereum/crypto"
)

func Start(properties ConfigProperties) error {
	logLevel, err := zerolog.ParseLevel(properties.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: [%w]", err)
	}

	logger := log.NewLogger(
		os.Stdout,
		log.LevelOption(logLevel),
	).With(log.ModuleKey, "bridge-worker")

	cfg, err := FromProperties(properties)
	if err != nil {
		return fmt.Errorf("config error: [%w]", err)
	}

	privateKey, err := DecryptKeyFile(
		cfg.Ethereum.Account.KeyFile,
		cfg.Ethereum.Account.KeyFilePassword,
	)
	if err != nil {
		return fmt.Errorf("keyfile load error: [%w]", err)
	}

	accountAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	logger.Info(
		"loaded Ethereum private key",
		"account_address", accountAddress,
	)

	return RunBridgeWorker(
		logger,
		*cfg,
		privateKey,
	)
}
