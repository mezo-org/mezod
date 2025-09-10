package main

import (
	"fmt"
	"log"
	"net/url"

	metricsscraper "github.com/mezo-org/mezod/metrics-scraper"
	"github.com/mezo-org/mezod/utils"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics-scraper",
		Short: "Metrics scraper",
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := parseFlags(cmd)
			if err != nil {
				return fmt.Errorf("failed to parse flags: [%w]", err)
			}

			return metricsscraper.Start(config)
		},
	}

	cmd.Flags().AddFlagSet(newFlagSet())

	return cmd
}

func parseFlags(cmd *cobra.Command) (metricsscraper.Config, error) {
	nodesConfigPath, err := cmd.Flags().GetString(flagNodesConfig)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("failed to get nodes path: [%w]", err)
	}
	if len(nodesConfigPath) == 0 {
		return metricsscraper.Config{}, fmt.Errorf("nodes config path is required")
	}

	prometheusPort, err := cmd.Flags().GetUint(flagPrometheusPort)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("failed to get prometheus port: [%w]", err)
	}
	if prometheusPort == 0 {
		return metricsscraper.Config{}, fmt.Errorf("prometheus port must be greater than 0")
	}

	chainID, err := cmd.Flags().GetString(flagChainID)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("failed to get chain ID: [%w]", err)
	}
	if !utils.IsMainnet(chainID) && !utils.IsTestnet(chainID) {
		return metricsscraper.Config{}, fmt.Errorf("invalid chain ID")
	}

	nodePollRate, err := cmd.Flags().GetDuration(flagNodePollRate)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("failed to get node poll rate: [%w]", err)
	}
	if nodePollRate <= 0 {
		return metricsscraper.Config{}, fmt.Errorf("node poll rate must be greater than 0")
	}

	bridgePollRate, err := cmd.Flags().GetDuration(flagBridgePollRate)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("failed to get bridge poll rate: [%w]", err)
	}
	if bridgePollRate <= 0 {
		return metricsscraper.Config{}, fmt.Errorf("bridge poll rate must be greater than 0")
	}

	mezoRPCURL, err := cmd.Flags().GetString(flagMezoRPCURL)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("failed to get Mezo RPC URL: [%w]", err)
	}
	_, err = url.ParseRequestURI(mezoRPCURL)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("mezo RPC URL is not valid: [%w]", err)
	}

	ethereumRPCURL, err := cmd.Flags().GetString(flagEthereumRPCURL)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("failed to get Ethereum RPC URL: [%w]", err)
	}
	_, err = url.ParseRequestURI(ethereumRPCURL)
	if err != nil {
		return metricsscraper.Config{}, fmt.Errorf("ethereum RPC URL is not valid: [%w]", err)
	}

	return metricsscraper.Config{
		NodesConfigPath: nodesConfigPath,
		PrometheusPort:  prometheusPort,
		ChainID:         chainID,
		NodePollRate:    nodePollRate,
		BridgePollRate:  bridgePollRate,
		MezoRPCURL:      mezoRPCURL,
		EthereumRPCURL:  ethereumRPCURL,
	}, nil
}

func main() {
	rootCmd := newRootCmd()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
