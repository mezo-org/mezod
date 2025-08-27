package bridgeworker

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

const PasswordEnvVariable = "BRIDGE_WORKER_KEY_PASSWORD"

type Config struct {
	ProviderURL     string  `json:"provider_url"`     // http(s)/wss Ethereum endpoint
	EthereumNetwork string  `json:"ethereum_network"` // e.g. "sepolia" | "mainnet"
	Account         Account `json:"account"`
}

type Account struct {
	KeyFile         string `json:"key_file"`          // path to geth keystore JSON
	KeyFilePassword string `json:"key_file_password"` // read from env
}

func DecryptKeyFile(keyFile, password string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read key file: %v", err)
	}
	key, err := keystore.DecryptKey(data, password)
	if err != nil {
		return nil, fmt.Errorf("decrypt key file: %v", err)
	}
	return key.PrivateKey, nil
}

func ReadConfig(path string) (*Config, error) {
	cfg := &Config{}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %v", err)
	}

	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %v", err)
	}

	// Load password from env.
	cfg.Account.KeyFilePassword = os.Getenv(PasswordEnvVariable)

	if cfg.Account.KeyFilePassword == "" {
		return nil, fmt.Errorf(
			"%s not set; please export the keystore password",
			PasswordEnvVariable,
		)
	}

	if cfg.ProviderURL == "" {
		return nil, fmt.Errorf("provider_url is required")
	}

	if cfg.EthereumNetwork == "" {
		return nil, fmt.Errorf("ethereum_network is required")
	}

	if cfg.Account.KeyFile == "" {
		return nil, fmt.Errorf("account.key_file is required")
	}

	return cfg, nil
}
