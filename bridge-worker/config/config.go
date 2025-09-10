package config

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/bridge-worker/bitcoin/electrum"
)

const (
	PasswordEnvVariable              = "BRIDGE_WORKER_KEY_PASSWORD"
	DefaultEthereumBatchSize         = 1000
	DefaultEthereumRequestsPerMinute = 600
)

type Config struct {
	Ethereum EthereumConfig `json:"ethereum"`
	Bitcoin  BitcoinConfig  `json:"bitcoin"`
	Mezo     MezoConfig     `json:"mezo"`
}

type BitcoinConfig struct {
	Network  bitcoin.Network `json:"network"` // "mainnet" | "testnet" | "regtest"
	Electrum electrum.Config `json:"electrum"`
}

type EthereumConfig struct {
	ProviderURL       string          `json:"provider_url"`        // http(s)/wss Ethereum endpoint
	Network           string          `json:"network"`             // e.g. "sepolia" | "mainnet"
	BatchSize         uint64          `json:"batch_size"`          // e.g. 1000
	RequestsPerMinute uint64          `json:"requests_per_minute"` // e.g. 600 (10 per second)
	Account           EthereumAccount `json:"account"`
}

type EthereumAccount struct {
	KeyFile         string `json:"key_file"`          // path to geth keystore JSON
	KeyFilePassword string `json:"key_file_password"` // read from env
}

type MezoConfig struct {
	AssetsUnlockEndpoint string `json:"assets_unlock_endpoint"` // e.g. "127.0.0.1:9090"
}

func (c *Config) applyDefaults() {
	// Ethereum
	if c.Ethereum.BatchSize == 0 {
		c.Ethereum.BatchSize = DefaultEthereumBatchSize
	}
	if c.Ethereum.RequestsPerMinute == 0 {
		c.Ethereum.RequestsPerMinute = DefaultEthereumRequestsPerMinute
	}

	// Bitcoin
	if c.Bitcoin.Electrum.ConnectTimeout == 0 {
		c.Bitcoin.Electrum.ConnectTimeout = electrum.DefaultConnectTimeout
	}
	if c.Bitcoin.Electrum.ConnectRetryTimeout == 0 {
		c.Bitcoin.Electrum.ConnectRetryTimeout = electrum.DefaultConnectRetryTimeout
	}
	if c.Bitcoin.Electrum.RequestTimeout == 0 {
		c.Bitcoin.Electrum.RequestTimeout = electrum.DefaultRequestTimeout
	}
	if c.Bitcoin.Electrum.RequestRetryTimeout == 0 {
		c.Bitcoin.Electrum.RequestRetryTimeout = electrum.DefaultRequestRetryTimeout
	}
	if c.Bitcoin.Electrum.KeepAliveInterval == 0 {
		c.Bitcoin.Electrum.KeepAliveInterval = electrum.DefaultKeepAliveInterval
	}
}

func (c *Config) validate() error {
	// Ethereum
	if c.Ethereum.ProviderURL == "" {
		return fmt.Errorf("ethereum.provider_url is required")
	}
	if c.Ethereum.Network == "" {
		return fmt.Errorf("ethereum.network is required")
	}
	if c.Ethereum.BatchSize == 0 {
		return fmt.Errorf("ethereum.batch_size is required")
	}
	if c.Ethereum.RequestsPerMinute == 0 {
		return fmt.Errorf("ethereum.requests_per_minute is required")
	}
	if c.Ethereum.Account.KeyFile == "" {
		return fmt.Errorf("ethereum.account.key_file is required")
	}
	if c.Ethereum.Account.KeyFilePassword == "" {
		return fmt.Errorf("%s not set; please export the keystore password", PasswordEnvVariable)
	}

	// Bitcoin
	if c.Bitcoin.Network == bitcoin.Unknown {
		return fmt.Errorf("bitcoin.network is required")
	}
	if c.Bitcoin.Electrum.URL == "" {
		return fmt.Errorf("bitcoin.electrum.url is required")
	}

	// Mezo
	if c.Mezo.AssetsUnlockEndpoint == "" {
		return fmt.Errorf("mezo.assets_unlock_endpoint is required")
	}

	return nil
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
	var cfg Config

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}

	cfg.Ethereum.Account.KeyFilePassword = os.Getenv(PasswordEnvVariable)

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
