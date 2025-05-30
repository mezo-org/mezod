// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package config

import (
	"errors"
	"fmt"
	"path"
	"time"

	oracleconfig "github.com/skip-mev/connect/v2/oracle/config"

	"github.com/spf13/viper"

	"github.com/cometbft/cometbft/libs/strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/server/config"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// DefaultGRPCAddress is the default address the gRPC server binds to.
	DefaultGRPCAddress = "0.0.0.0:9900"

	// DefaultJSONRPCAddress is the default address the JSON-RPC server binds to.
	DefaultJSONRPCAddress = "127.0.0.1:8545"

	// DefaultJSONRPCWsAddress is the default address the JSON-RPC WebSocket server binds to.
	DefaultJSONRPCWsAddress = "127.0.0.1:8546"

	// DefaultJsonRPCMetricsAddress is the default address the JSON-RPC Metrics server binds to.
	DefaultJSONRPCMetricsAddress = "127.0.0.1:6065"

	// DefaultEVMTracer is the default vm.Tracer type
	DefaultEVMTracer = ""

	// DefaultFixRevertGasRefundHeight is the default height at which to overwrite gas refund
	DefaultFixRevertGasRefundHeight = 0

	// DefaultMaxTxGasWanted is the default gas wanted for each eth tx returned in ante handler in check tx mode
	DefaultMaxTxGasWanted = 0

	// DefaultGasCap is the default cap on gas that can be used in eth_call/estimateGas
	DefaultGasCap uint64 = 25000000

	// DefaultFilterCap is the default cap for total number of filters that can be created
	DefaultFilterCap int32 = 200

	// DefaultFeeHistoryCap is the default cap for total number of blocks that can be fetched
	DefaultFeeHistoryCap int32 = 100

	// DefaultLogsCap is the default cap of results returned from single 'eth_getLogs' query
	DefaultLogsCap int32 = 10000

	// DefaultBlockRangeCap is the default cap of block range allowed for 'eth_getLogs' query
	DefaultBlockRangeCap int32 = 10000

	// DefaultEVMTimeout is the default timeout for eth_call
	DefaultEVMTimeout = 5 * time.Second

	// DefaultTxFeeCap is the default tx-fee cap for sending a transaction
	DefaultTxFeeCap float64 = 1.0

	// DefaultHTTPTimeout is the default read/write timeout of the http json-rpc server
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultHTTPIdleTimeout is the default idle timeout of the http json-rpc server
	DefaultHTTPIdleTimeout = 120 * time.Second

	// DefaultAllowUnprotectedTxs value is false
	DefaultAllowUnprotectedTxs = false

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0

	// DefaultEthereumSidecarServerAddress is the default address of the Ethereum sidecar server.
	DefaultEthereumSidecarServerAddress = "127.0.0.1:7500"

	// DefaultEthereumSidecarRequestTimeout is the default timeout for requests to the Ethereum sidecar server.
	DefaultEthereumSidecarRequestTimeout = 5 * time.Second

	// DefaultConnectOracleEnabled is the default value indicating whether the oracle is enabled.
	DefaultConnectOracleEnabled = true

	// DefaultConnectOracleAddress is the default address of the oracle sidecar.
	DefaultConnectOracleAddress = "localhost:8080"

	// DefaultConnectOracleClientTimeout is the default time that the client is
	// willing to wait for responses from the oracle sidecar before timing out.
	DefaultConnectOracleClientTimeout = 2 * time.Second

	// DefaultConnectOracleMetricsEnabled is the default value indicating
	// whether the oracle metrics are enabled.
	DefaultConnectOracleMetricsEnabled = true

	// DefaultConnectOracleInterval is the time between each price update request.
	// The recommended interval is the block time of the chain.
	DefaultConnectOracleInterval = 1500 * time.Millisecond

	// DefaultConnectOraclePriceTTL is the maximum age of the latest price
	// response before it is considered stale.
	DefaultConnectOraclePriceTTL = 10 * time.Second
)

var evmTracers = []string{"json", "markdown", "struct", "access_list"}

// Config defines the server's top level configuration. It includes the default app config
// from the SDK as well as the EVM configuration to enable the JSON-RPC APIs.
type Config struct {
	config.Config

	EVM             EVMConfig              `mapstructure:"evm"`
	JSONRPC         JSONRPCConfig          `mapstructure:"json-rpc"`
	TLS             TLSConfig              `mapstructure:"tls"`
	EthereumSidecar EthereumSidecarConfig  `mapstructure:"ethereum-sidecar.client"`
	Oracle          oracleconfig.AppConfig `mapstructure:"oracle"`
}

// EVMConfig defines the application configuration values for the EVM.
type EVMConfig struct {
	// Tracer defines vm.Tracer type that the EVM will use if the node is run in
	// trace mode. Default: 'json'.
	Tracer string `mapstructure:"tracer"`
	// MaxTxGasWanted defines the gas wanted for each eth tx returned in ante handler in check tx mode.
	MaxTxGasWanted uint64 `mapstructure:"max-tx-gas-wanted"`
}

// JSONRPCConfig defines configuration for the EVM RPC server.
type JSONRPCConfig struct {
	// API defines a list of JSON-RPC namespaces that should be enabled
	API []string `mapstructure:"api"`
	// Address defines the HTTP server to listen on
	Address string `mapstructure:"address"`
	// WsAddress defines the WebSocket server to listen on
	WsAddress string `mapstructure:"ws-address"`
	// GasCap is the global gas cap for eth-call variants.
	GasCap uint64 `mapstructure:"gas-cap"`
	// EVMTimeout is the global timeout for eth-call.
	EVMTimeout time.Duration `mapstructure:"evm-timeout"`
	// TxFeeCap is the global tx-fee cap for send transaction
	TxFeeCap float64 `mapstructure:"txfee-cap"`
	// FilterCap is the global cap for total number of filters that can be created.
	FilterCap int32 `mapstructure:"filter-cap"`
	// FeeHistoryCap is the global cap for total number of blocks that can be fetched
	FeeHistoryCap int32 `mapstructure:"feehistory-cap"`
	// Enable defines if the EVM RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// LogsCap defines the max number of results can be returned from single `eth_getLogs` query.
	LogsCap int32 `mapstructure:"logs-cap"`
	// BlockRangeCap defines the max block range allowed for `eth_getLogs` query.
	BlockRangeCap int32 `mapstructure:"block-range-cap"`
	// HTTPTimeout is the read/write timeout of http json-rpc server.
	HTTPTimeout time.Duration `mapstructure:"http-timeout"`
	// HTTPIdleTimeout is the idle timeout of http json-rpc server.
	HTTPIdleTimeout time.Duration `mapstructure:"http-idle-timeout"`
	// AllowUnprotectedTxs restricts unprotected (non EIP155 signed) transactions to be submitted via
	// the node's RPC when global parameter is disabled.
	AllowUnprotectedTxs bool `mapstructure:"allow-unprotected-txs"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int `mapstructure:"max-open-connections"`
	// EnableIndexer defines if enable the custom indexer service.
	EnableIndexer bool `mapstructure:"enable-indexer"`
	// MetricsAddress defines the metrics server to listen on
	MetricsAddress string `mapstructure:"metrics-address"`
	// FixRevertGasRefundHeight defines the upgrade height for fix of revert gas refund logic when transaction reverted
	FixRevertGasRefundHeight int64 `mapstructure:"fix-revert-gas-refund-height"`
}

// TLSConfig defines the certificate and matching private key for the server.
type TLSConfig struct {
	// CertificatePath the file path for the certificate .pem file
	CertificatePath string `mapstructure:"certificate-path"`
	// KeyPath the file path for the key .pem file
	KeyPath string `mapstructure:"key-path"`
}

// Defines the configuration for the Ethereum sidecar configuration.
type EthereumSidecarConfig struct {
	// ServerAddress is the address of the Ethereum sidecar server to connect to.
	ServerAddress string `mapstructure:"server-address"`
	// RequestTimeout is the timeout for requests to the Ethereum sidecar server.
	RequestTimeout time.Duration `mapstructure:"request-timeout"`
}

// AppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func AppConfig(denom string) (string, interface{}) {
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := config.DefaultConfig()

	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In mezo, we set the min gas prices to 0.
	if denom != "" {
		srvCfg.MinGasPrices = "0" + denom
	}

	customAppConfig := Config{
		Config:          *srvCfg,
		EVM:             *DefaultEVMConfig(),
		JSONRPC:         *DefaultJSONRPCConfig(),
		TLS:             *DefaultTLSConfig(),
		EthereumSidecar: *DefaultEthereumSidecarConfig(),
		Oracle:          *DefaultOracleConfig(),
	}

	customAppTemplate := config.DefaultConfigTemplate + DefaultConfigTemplate + oracleconfig.DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		Config:          *config.DefaultConfig(),
		EVM:             *DefaultEVMConfig(),
		JSONRPC:         *DefaultJSONRPCConfig(),
		TLS:             *DefaultTLSConfig(),
		EthereumSidecar: *DefaultEthereumSidecarConfig(),
		Oracle:          *DefaultOracleConfig(),
	}
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultEVMConfig() *EVMConfig {
	return &EVMConfig{
		Tracer:         DefaultEVMTracer,
		MaxTxGasWanted: DefaultMaxTxGasWanted,
	}
}

func DefaultOracleConfig() *oracleconfig.AppConfig {
	return &oracleconfig.AppConfig{
		Enabled:        DefaultConnectOracleEnabled,
		OracleAddress:  DefaultConnectOracleAddress,
		ClientTimeout:  DefaultConnectOracleClientTimeout,
		MetricsEnabled: DefaultConnectOracleMetricsEnabled,
		PriceTTL:       DefaultConnectOraclePriceTTL,
		Interval:       DefaultConnectOracleInterval,
	}
}

// Validate returns an error if the tracer type is invalid.
func (c EVMConfig) Validate() error {
	if c.Tracer != "" && !strings.StringInSlice(c.Tracer, evmTracers) {
		return fmt.Errorf("invalid tracer type %s, available types: %v", c.Tracer, evmTracers)
	}

	return nil
}

// GetDefaultAPINamespaces returns the default list of JSON-RPC namespaces that should be enabled
func GetDefaultAPINamespaces() []string {
	return []string{"eth", "net", "web3", "mezo"}
}

// GetAPINamespaces returns the all the available JSON-RPC API namespaces.
func GetAPINamespaces() []string {
	return []string{"web3", "eth", "personal", "net", "txpool", "debug", "miner", "mezo"}
}

// DefaultJSONRPCConfig returns an EVM config with the JSON-RPC API enabled by default
func DefaultJSONRPCConfig() *JSONRPCConfig {
	return &JSONRPCConfig{
		Enable:                   true,
		API:                      GetDefaultAPINamespaces(),
		Address:                  DefaultJSONRPCAddress,
		WsAddress:                DefaultJSONRPCWsAddress,
		GasCap:                   DefaultGasCap,
		EVMTimeout:               DefaultEVMTimeout,
		TxFeeCap:                 DefaultTxFeeCap,
		FilterCap:                DefaultFilterCap,
		FeeHistoryCap:            DefaultFeeHistoryCap,
		BlockRangeCap:            DefaultBlockRangeCap,
		LogsCap:                  DefaultLogsCap,
		HTTPTimeout:              DefaultHTTPTimeout,
		HTTPIdleTimeout:          DefaultHTTPIdleTimeout,
		AllowUnprotectedTxs:      DefaultAllowUnprotectedTxs,
		MaxOpenConnections:       DefaultMaxOpenConnections,
		EnableIndexer:            false,
		MetricsAddress:           DefaultJSONRPCMetricsAddress,
		FixRevertGasRefundHeight: DefaultFixRevertGasRefundHeight,
	}
}

// Validate returns an error if the JSON-RPC configuration fields are invalid.
func (c JSONRPCConfig) Validate() error {
	if c.Enable && len(c.API) == 0 {
		return errors.New("cannot enable JSON-RPC without defining any API namespace")
	}

	if c.FilterCap < 0 {
		return errors.New("JSON-RPC filter-cap cannot be negative")
	}

	if c.FeeHistoryCap <= 0 {
		return errors.New("JSON-RPC feehistory-cap cannot be negative or 0")
	}

	if c.TxFeeCap < 0 {
		return errors.New("JSON-RPC tx fee cap cannot be negative")
	}

	if c.EVMTimeout < 0 {
		return errors.New("JSON-RPC EVM timeout duration cannot be negative")
	}

	if c.LogsCap < 0 {
		return errors.New("JSON-RPC logs cap cannot be negative")
	}

	if c.BlockRangeCap < 0 {
		return errors.New("JSON-RPC block range cap cannot be negative")
	}

	if c.HTTPTimeout < 0 {
		return errors.New("JSON-RPC HTTP timeout duration cannot be negative")
	}

	if c.HTTPIdleTimeout < 0 {
		return errors.New("JSON-RPC HTTP idle timeout duration cannot be negative")
	}

	// check for duplicates
	seenAPIs := make(map[string]bool)
	for _, api := range c.API {
		if seenAPIs[api] {
			return fmt.Errorf("repeated API namespace '%s'", api)
		}

		seenAPIs[api] = true
	}

	return nil
}

// DefaultTLSConfig returns the default TLS configuration
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		CertificatePath: "",
		KeyPath:         "",
	}
}

// Validate returns an error if the TLS certificate and key file extensions are invalid.
func (c TLSConfig) Validate() error {
	certExt := path.Ext(c.CertificatePath)

	if c.CertificatePath != "" && certExt != ".pem" {
		return fmt.Errorf("invalid extension %s for certificate path %s, expected '.pem'", certExt, c.CertificatePath)
	}

	keyExt := path.Ext(c.KeyPath)

	if c.KeyPath != "" && keyExt != ".pem" {
		return fmt.Errorf("invalid extension %s for key path %s, expected '.pem'", keyExt, c.KeyPath)
	}

	return nil
}

// DefaultEthereumSidecarConfig returns the default Ethereum sidecar configuration.
func DefaultEthereumSidecarConfig() *EthereumSidecarConfig {
	return &EthereumSidecarConfig{
		ServerAddress:  DefaultEthereumSidecarServerAddress,
		RequestTimeout: DefaultEthereumSidecarRequestTimeout,
	}
}

func (c EthereumSidecarConfig) Validate() error {
	if c.ServerAddress == "" {
		return fmt.Errorf("ethereum sidecar server address cannot be empty")
	}

	if c.RequestTimeout < 0 {
		return fmt.Errorf("ethereum sidecar request timeout cannot be negative")
	}

	return nil
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) (Config, error) {
	cfg, err := config.GetConfig(v)
	if err != nil {
		return Config{}, err
	}

	// Because the configuration used to generate wrong
	// default, these two could prevent the node to start
	// after an upgrade and panic. To avoid this we check for
	// the go zero value and assign the default value instead
	var (
		oraclePriceTTL = v.GetDuration("oracle.price_ttl")
		oracleInterval = v.GetDuration("oracle.interval")
	)
	if oraclePriceTTL == 0 {
		oraclePriceTTL = DefaultConnectOraclePriceTTL
	}
	if oracleInterval == 0 {
		oracleInterval = DefaultConnectOracleInterval
	}

	return Config{
		Config: cfg,
		EVM: EVMConfig{
			Tracer:         v.GetString("evm.tracer"),
			MaxTxGasWanted: v.GetUint64("evm.max-tx-gas-wanted"),
		},
		JSONRPC: JSONRPCConfig{
			Enable:                   v.GetBool("json-rpc.enable"),
			API:                      v.GetStringSlice("json-rpc.api"),
			Address:                  v.GetString("json-rpc.address"),
			WsAddress:                v.GetString("json-rpc.ws-address"),
			GasCap:                   v.GetUint64("json-rpc.gas-cap"),
			FilterCap:                v.GetInt32("json-rpc.filter-cap"),
			FeeHistoryCap:            v.GetInt32("json-rpc.feehistory-cap"),
			TxFeeCap:                 v.GetFloat64("json-rpc.txfee-cap"),
			EVMTimeout:               v.GetDuration("json-rpc.evm-timeout"),
			LogsCap:                  v.GetInt32("json-rpc.logs-cap"),
			BlockRangeCap:            v.GetInt32("json-rpc.block-range-cap"),
			HTTPTimeout:              v.GetDuration("json-rpc.http-timeout"),
			HTTPIdleTimeout:          v.GetDuration("json-rpc.http-idle-timeout"),
			MaxOpenConnections:       v.GetInt("json-rpc.max-open-connections"),
			EnableIndexer:            v.GetBool("json-rpc.enable-indexer"),
			MetricsAddress:           v.GetString("json-rpc.metrics-address"),
			FixRevertGasRefundHeight: v.GetInt64("json-rpc.fix-revert-gas-refund-height"),
			AllowUnprotectedTxs:      v.GetBool("json-rpc.allow-unprotected-txs"),
		},
		TLS: TLSConfig{
			CertificatePath: v.GetString("tls.certificate-path"),
			KeyPath:         v.GetString("tls.key-path"),
		},
		EthereumSidecar: EthereumSidecarConfig{
			ServerAddress:  v.GetString("ethereum-sidecar.client.server-address"),
			RequestTimeout: v.GetDuration("ethereum-sidecar.client.request-timeout"),
		},
		Oracle: oracleconfig.AppConfig{
			Enabled:        v.GetBool("oracle.enabled"),
			OracleAddress:  v.GetString("oracle.oracle_address"),
			ClientTimeout:  v.GetDuration("oracle.client_timeout"),
			MetricsEnabled: v.GetBool("oracle.metrics_enabled"),
			PriceTTL:       oraclePriceTTL,
			Interval:       oracleInterval,
		},
	}, nil
}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig(v *viper.Viper) (*Config, error) {
	conf := DefaultConfig()
	err := v.Unmarshal(conf)

	return conf, err
}

// ValidateBasic returns an error any of the application configuration fields are invalid
func (c Config) ValidateBasic() error {
	if err := c.EVM.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid evm config value: %s", err.Error())
	}

	if err := c.JSONRPC.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid json-rpc config value: %s", err.Error())
	}

	if err := c.TLS.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid tls config value: %s", err.Error())
	}

	if err := c.EthereumSidecar.Validate(); err != nil {
		return errorsmod.Wrapf(errortypes.ErrAppConfig, "invalid ethereum sidecar config value: %s", err.Error())
	}

	return c.Config.ValidateBasic()
}
