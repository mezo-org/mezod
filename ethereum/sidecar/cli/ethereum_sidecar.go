package cli

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"google.golang.org/grpc/encoding"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	clientkeys "github.com/mezo-org/mezod/client/keys"
	"github.com/mezo-org/mezod/ethereum/sidecar"
	"github.com/spf13/cobra"

	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
)

func NewEthereumSidecarCmd() *cobra.Command {
	defaultServerAddress := "0.0.0.0:7500"
	defaultServerEthereumNodeAddress := "ws://127.0.0.1:8546"
	defaultServerEthereumNetwork := ethconfig.Sepolia
	defaultServerBatchSize := uint64(1000)
	// Default requests per minute. A 'minute' unit was chosen so that a
	// validator can chose from the wider range of values.
	// This default value might not work with all the Ethereum providers and should
	// be adjusted based on the Ethereum provider capabilities. It happens that a
	// free provider account won't even accept 1 request per second, meaning
	// that this number should be adjusted by setting a flag to e.g. 1 request
	// per 2 seconds, which is 30 requests per minute. Flag should be set to 30 for
	// this example.
	defaultServerRequestsPerMinute := uint64(600) // 10 requests per second
	defaultAssetsUnlockedEndpoint := "127.0.0.1:9090"
	defaultKeyringBackend := flags.DefaultKeyringBackend
	defaultKeyringDir := ""
	defaultKeyName := ""

	cmd := &cobra.Command{
		Use:   "ethereum-sidecar",
		Short: "Starts Ethereum sidecar",
		Long: "Starts the Ethereum sidecar which observes the assets locked " +
			"in the Mezo bridge on the Ethereum chain",
		RunE: runEthereumSidecar,
	}

	cmd.Flags().AddFlagSet(
		NewFlagSetEthereumSidecar(
			defaultServerAddress,
			defaultServerEthereumNodeAddress,
			defaultServerEthereumNetwork.String(),
			defaultServerBatchSize,
			defaultServerRequestsPerMinute,
			defaultAssetsUnlockedEndpoint,
			defaultKeyringBackend,
			defaultKeyringDir,
			defaultKeyName,
		))

	return cmd
}

func runEthereumSidecar(cmd *cobra.Command, _ []string) error {
	logger := server.GetServerContextFromCmd(cmd).Logger.With(
		"module",
		"ethereum-sidecar",
	)

	grpcAddress, _ := cmd.Flags().GetString(FlagServerAddress)
	ethNodeAddress, _ := cmd.Flags().GetString(FlagServerEthereumNodeAddress)
	network, _ := cmd.Flags().GetString(FlagServerNetwork)
	batchSize, _ := cmd.Flags().GetUint64(FlagServerBatchSize)
	requestsPerMinute, _ := cmd.Flags().GetUint64(FlagServerRequestsPerMinute)
	assetsUnlockedEndpoint, _ := cmd.Flags().GetString(FlagAssetsUnlockedEndpoint)
	keyName, _ := cmd.Flags().GetString(FlagKeyName)

	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	// The messages handled by the server contain custom types. Add codecs so
	// that the messages can be marshaled/unmarshalled.
	encoding.RegisterCodec(
		codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec(),
	)

	// Extract private key from keyring if key name is provided
	var privateKey *ecdsa.PrivateKey
	if keyName != "" {
		privateKey, err = clientkeys.ExtractPrivateKey(cmd, keyName)
		if err != nil {
			return fmt.Errorf("failed to extract private key: %w", err)
		}
		logger.Info(
			"successfully extracted private key from keyring",
			"keyName", keyName,
			"address", crypto.PubkeyToAddress(privateKey.PublicKey).Hex(),
		)
	} else {
		// If the key name is not provided, generate a random key. This key
		// won't be used by the sidecar for signing but is needed to be present.
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate private key: %w", err)
		}
	}

	sidecar.RunServer(
		logger,
		grpcAddress,
		ethNodeAddress,
		network,
		batchSize,
		requestsPerMinute,
		assetsUnlockedEndpoint,
		clientCtx.InterfaceRegistry,
		privateKey,
	)

	return nil
}
