package config

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

const PasswordEnvVariable = "BRIDGE_WORKER_KEY_PASSWORD"

type Config struct {
	ProviderURL       string          `json:"provider_url"`        // http(s)/wss Ethereum endpoint
	EthereumNetwork   string          `json:"ethereum_network"`    // e.g. "sepolia" | "mainnet"
	BatchSize         uint64          `json:"batch_size"`          // e.g. 1000
	RequestsPerMinute uint64          `json:"requests_per_minute"` // e.g. 600 (10 per second)
	EthereumAccount   EthereumAccount `json:"ethereum_account"`
	BitcoinAccount    BitcoinAccount  `json:"bitcoin_account"`
}

type EthereumAccount struct {
	KeyFile         string `json:"key_file"`          // path to geth keystore JSON
	KeyFilePassword string `json:"key_file_password"` // read from env
}

type BitcoinAccount struct {
	URL      string `json:"url"`      // e.g. "127.0.0.1:8332"
	Password string `json:"password"` // e.g. "password"
	Username string `json:"user"`     // e.g. "user"
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
	cfg.EthereumAccount.KeyFilePassword = os.Getenv(PasswordEnvVariable)

	if cfg.EthereumAccount.KeyFilePassword == "" {
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

	if cfg.BatchSize == 0 {
		return nil, fmt.Errorf("batch_size is required")
	}

	if cfg.RequestsPerMinute == 0 {
		return nil, fmt.Errorf("requests_per_minute is required")
	}

	if cfg.EthereumAccount.KeyFile == "" {
		return nil, fmt.Errorf("ethereum_account.key_file is required")
	}

	if cfg.BitcoinAccount.URL == "" {
		return nil, fmt.Errorf("bitcoin_account.url is required")
	}

	if cfg.BitcoinAccount.Password == "" {
		return nil, fmt.Errorf("bitcoin_account.password is required")
	}

	if cfg.BitcoinAccount.Username == "" {
		return nil, fmt.Errorf("bitcoin_account.username is required")
	}

	return cfg, nil
}
