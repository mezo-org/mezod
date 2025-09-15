package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	bridgeworker "github.com/mezo-org/mezod/bridge-worker"
	"github.com/spf13/cobra"
)

const EthereumAccountKeyFilePasswordEnv = "ETHEREUM_ACCOUNT_KEY_FILE_PASSWORD"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge-worker",
		Short: "Bridge worker",
		RunE: func(cmd *cobra.Command, _ []string) error {
			properties, err := parseFlags(cmd)
			if err != nil {
				return fmt.Errorf("failed to parse flags: [%w]", err)
			}

			return bridgeworker.Start(properties)
		},
	}

	cmd.Flags().AddFlagSet(newFlagSet())

	return cmd
}

func parseFlags(cmd *cobra.Command) (bridgeworker.ConfigProperties, error) {
	logLevel, err := cmd.Flags().GetString(flagLogLevel)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get log level: [%w]", err)
	}
	if len(logLevel) == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("log level is required")
	}

	logFormatJSON, err := cmd.Flags().GetBool(flagLogFormatJSON)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get log JSON: [%w]", err)
	}

	ethereumProviderURL, err := cmd.Flags().GetString(flagEthereumProviderURL)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get ethereum provider URL: [%w]", err)
	}
	_, err = url.ParseRequestURI(ethereumProviderURL)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("ethereum provider URL is not valid: [%w]", err)
	}

	ethereumNetwork, err := cmd.Flags().GetString(flagEthereumNetwork)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get ethereum network: [%w]", err)
	}
	if len(ethereumNetwork) == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("ethereum network is required")
	}

	ethereumBatchSize, err := cmd.Flags().GetUint64(flagEthereumBatchSize)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get ethereum batch size: [%w]", err)
	}
	if ethereumBatchSize == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("ethereum batch size must be greater than 0")
	}

	ethereumRequestsPerMinute, err := cmd.Flags().GetUint64(flagEthereumRequestsPerMinute)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get ethereum requests per minute: [%w]", err)
	}
	if ethereumRequestsPerMinute == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("ethereum requests per minute must be greater than 0")
	}

	ethereumAccountKeyFile, err := cmd.Flags().GetString(flagEthereumAccountKeyFile)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get ethereum account key file: [%w]", err)
	}
	if len(ethereumAccountKeyFile) == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("ethereum account key file is required")
	}

	ethereumAccountKeyFilePassword := os.Getenv(EthereumAccountKeyFilePasswordEnv)
	if len(ethereumAccountKeyFilePassword) == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf(
			"ethereum account key file password is required; "+
				"make sure the %s environment variable is set",
			EthereumAccountKeyFilePasswordEnv,
		)
	}

	bitcoinNetwork, err := cmd.Flags().GetString(flagBitcoinNetwork)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get bitcoin network: [%w]", err)
	}
	if len(bitcoinNetwork) == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("bitcoin network is required")
	}

	bitcoinElectrumURL, err := cmd.Flags().GetString(flagBitcoinElectrumURL)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get bitcoin electrum URL: [%w]", err)
	}
	_, err = url.ParseRequestURI(bitcoinElectrumURL)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("bitcoin electrum URL is not valid: [%w]", err)
	}

	mezoAssetsUnlockEndpoint, err := cmd.Flags().GetString(flagMezoAssetsUnlockEndpoint)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get mezo assets unlock endpoint: [%w]", err)
	}
	_, err = url.ParseRequestURI(mezoAssetsUnlockEndpoint)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("mezo assets unlock endpoint is not valid: [%w]", err)
	}

	jobBTCWithdrawalQueueCheckFrequency, err := cmd.Flags().GetDuration(flagJobBTCWithdrawalQueueCheckFrequency)
	if err != nil {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("failed to get BTC withdrawal job queue check frequency: [%w]", err)
	}
	if jobBTCWithdrawalQueueCheckFrequency == 0 {
		return bridgeworker.ConfigProperties{}, fmt.Errorf("BTC withdrawal job queue check frequency must be greater than 0")
	}

	return bridgeworker.ConfigProperties{
		LogLevel:                            logLevel,
		LogFormatJSON:                       logFormatJSON,
		EthereumProviderURL:                 ethereumProviderURL,
		EthereumNetwork:                     ethereumNetwork,
		EthereumBatchSize:                   ethereumBatchSize,
		EthereumRequestsPerMinute:           ethereumRequestsPerMinute,
		EthereumAccountKeyFile:              ethereumAccountKeyFile,
		EthereumAccountKeyFilePassword:      ethereumAccountKeyFilePassword,
		BitcoinNetwork:                      bitcoinNetwork,
		BitcoinElectrumURL:                  bitcoinElectrumURL,
		MezoAssetsUnlockEndpoint:            mezoAssetsUnlockEndpoint,
		JobBTCWithdrawalQueueCheckFrequency: jobBTCWithdrawalQueueCheckFrequency,
	}, nil
}

func main() {
	rootCmd := newRootCmd()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
