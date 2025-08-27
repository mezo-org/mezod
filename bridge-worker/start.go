package bridgeworker

import (
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	bwconfig "github.com/mezo-org/mezod/bridge-worker/config"
)

func Start(configPath string) {
	cfg, err := bwconfig.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	privateKey, err := bwconfig.DecryptKeyFile(
		cfg.Account.KeyFile,
		cfg.Account.KeyFilePassword,
	)
	if err != nil {
		log.Fatalf("keyfile load error: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Printf("loaded Ethereum private key for address: %s", address.Hex())

	RunBridgeWorker(
		cfg.ProviderURL,
		cfg.EthereumNetwork,
		privateKey,
	)
}
