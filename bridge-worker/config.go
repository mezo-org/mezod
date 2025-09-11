package bridgeworker

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/bridge-worker/bitcoin/electrum"
)

const (
	DefaultEthereumBatchSize         = 1000
	DefaultEthereumRequestsPerMinute = 600

	DefaultBTCWithdrawalQueueCheckFrequency = 1 * time.Minute
)

type ConfigProperties struct {
	LogLevel      string
	LogFormatJSON bool

	EthereumProviderURL            string
	EthereumNetwork                string
	EthereumBatchSize              uint64
	EthereumRequestsPerMinute      uint64
	EthereumAccountKeyFile         string
	EthereumAccountKeyFilePassword string

	BitcoinNetwork     string
	BitcoinElectrumURL string

	MezoAssetsUnlockEndpoint string

	JobBTCWithdrawalQueueCheckFrequency time.Duration
}

type Config struct {
	Ethereum EthereumConfig
	Bitcoin  BitcoinConfig
	Mezo     MezoConfig
	Job      JobConfig
}

type BitcoinConfig struct {
	Network  bitcoin.Network // "mainnet" | "testnet" | "regtest"
	Electrum electrum.Config
}

type EthereumConfig struct {
	ProviderURL       string // http(s)/wss Ethereum endpoint
	Network           string // e.g. "sepolia" | "mainnet"
	BatchSize         uint64 // e.g. 1000
	RequestsPerMinute uint64 // e.g. 600 (10 per second)
	Account           EthereumAccount
}

type EthereumAccount struct {
	KeyFile         string // path to geth keystore JSON
	KeyFilePassword string // read from env
}

type MezoConfig struct {
	AssetsUnlockEndpoint string // e.g. "127.0.0.1:9090"
}

type JobConfig struct {
	BTCWithdrawal BTCWithdrawalConfig
}

type BTCWithdrawalConfig struct {
	QueueCheckFrequency time.Duration
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
		return fmt.Errorf("ethereum provider URL is required")
	}
	if c.Ethereum.Network == "" {
		return fmt.Errorf("ethereum network is required")
	}
	if c.Ethereum.BatchSize == 0 {
		return fmt.Errorf("ethereum batch size is required")
	}
	if c.Ethereum.RequestsPerMinute == 0 {
		return fmt.Errorf("ethereum requests per minute is required")
	}
	if c.Ethereum.Account.KeyFile == "" {
		return fmt.Errorf("ethereum account key file is required")
	}
	if c.Ethereum.Account.KeyFilePassword == "" {
		return fmt.Errorf("ethereum account key file password is required")
	}

	// Bitcoin
	if c.Bitcoin.Network == bitcoin.Unknown {
		return fmt.Errorf("bitcoin network is required")
	}
	if c.Bitcoin.Electrum.URL == "" {
		return fmt.Errorf("bitcoin electrum URL is required")
	}

	// Mezo
	if c.Mezo.AssetsUnlockEndpoint == "" {
		return fmt.Errorf("mezo assets unlock endpoint is required")
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

func FromProperties(properties ConfigProperties) (*Config, error) {
	var cfg Config

	cfg.Ethereum = EthereumConfig{
		ProviderURL:       properties.EthereumProviderURL,
		Network:           properties.EthereumNetwork,
		BatchSize:         properties.EthereumBatchSize,
		RequestsPerMinute: properties.EthereumRequestsPerMinute,
		Account: EthereumAccount{
			KeyFile:         properties.EthereumAccountKeyFile,
			KeyFilePassword: properties.EthereumAccountKeyFilePassword,
		},
	}

	var bitcoinNetwork bitcoin.Network
	switch properties.BitcoinNetwork {
	case bitcoin.Mainnet.String():
		bitcoinNetwork = bitcoin.Mainnet
	case bitcoin.Testnet.String():
		bitcoinNetwork = bitcoin.Testnet
	case bitcoin.Regtest.String():
		bitcoinNetwork = bitcoin.Regtest
	default:
		bitcoinNetwork = bitcoin.Unknown
	}

	cfg.Bitcoin = BitcoinConfig{
		Network: bitcoinNetwork,
		Electrum: electrum.Config{
			URL: properties.BitcoinElectrumURL,
		},
	}

	cfg.Mezo = MezoConfig{
		AssetsUnlockEndpoint: properties.MezoAssetsUnlockEndpoint,
	}

	cfg.Job = JobConfig{
		BTCWithdrawal: BTCWithdrawalConfig{
			QueueCheckFrequency: properties.JobBTCWithdrawalQueueCheckFrequency,
		},
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
