package poa

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	clientSdk "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	poatypes "github.com/mezo-org/mezod/x/poa/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/precompile/validatorpool"
	validatorpoolgen "github.com/mezo-org/mezod/precompile/validatorpool/gen"
)

type Validator struct {
	OperatorBech32   string               `json:"operator_bech32"`
	ConsPubKeyBech32 string               `json:"cons_pub_key_bech32"`
	Description      poatypes.Description `json:"validator"`
}

type Data struct {
	Memo      string    `json:"memo"`
	Validator Validator `json:"validator"`
}

const (
	localNodeURL = "http://127.0.0.1:8545"
)

func NewSubmitApplicationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "submit-application [key name]",
		Short:        "Application Submission to become PoA validator",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runSubmitApplication(cmd, args)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(NewFlagSetSubmitApplication())

	return cmd
}

func runSubmitApplication(cmd *cobra.Command, args []string) error {
	clientCtx, err := clientSdk.GetClientTxContext(cmd)
	if err != nil {
		return fmt.Errorf("unable to get client transaction context: %v", err)
	}

	moniker, identity, website, securityContact, details, rpcURL, err := getFlags(cmd)
	if err != nil {
		fmt.Println("Error getting flags: ", err)
		return err
	}

	consPubKeyArray, err := getConsensusPublicKey(clientCtx, sdk.GetConfig())
	if err != nil {
		return err
	}

	client, networkID, err := connectToEthereumNetwork(rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	instance, err := loadContract(client)
	if err != nil {
		return err
	}

	privateKey, err := extractPrivateKey(cmd, args[0])
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, networkID)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %v", err)
	}

	description := validatorpoolgen.Description{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}

	return submitApplication(instance, auth, consPubKeyArray, description)
}

func getFlags(cmd *cobra.Command) (string, string, string, string, string, string, error) {
	serverCtx := server.GetServerContextFromCmd(cmd)
	config := serverCtx.Config
	moniker := config.Moniker
	if monikerFlag, _ := cmd.Flags().GetString(FlagMoniker); monikerFlag != "" {
		moniker = monikerFlag
	}

	identity, err := cmd.Flags().GetString(FlagIdentity)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("unable to get identity flag: %v", err)
	}

	website, err := cmd.Flags().GetString(FlagWebsite)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("unable to get website flag: %v", err)
	}

	securityContact, err := cmd.Flags().GetString(FlagSecurityContact)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("unable to get security contact flag: %v", err)
	}

	details, err := cmd.Flags().GetString(FlagDetails)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("unable to get details flag: %v", err)
	}

	rpcURL, err := cmd.Flags().GetString(FlagRPCURL)
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("unable to get RPC URL flag: %v", err)
	}

	return moniker, identity, website, securityContact, details, rpcURL, nil
}

func getConsensusPublicKey(clientCtx clientSdk.Context, config *sdk.Config) ([32]byte, error) {
	homeDir := clientCtx.HomeDir
	pattern := filepath.Join(homeDir, "config", "genval", "genval-*.json")
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return [32]byte{}, fmt.Errorf("failed to search for files or no files found: %v", err)
	}

	filePath := files[0]
	content, err := os.ReadFile(filePath)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to read the file %s: %v", filePath, err)
	}

	var data Data
	if err := json.Unmarshal(content, &data); err != nil {
		return [32]byte{}, fmt.Errorf("failed to parse JSON from %s: %v", filePath, err)
	}

	prefix := config.GetBech32ConsensusPubPrefix()
	consPubKeyBytes, err := sdk.GetFromBech32(data.Validator.ConsPubKeyBech32, prefix)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to convert consensus pub key from Bech32: %v", err)
	}

	var consPubKeyArray [32]byte
	copy(consPubKeyArray[:], consPubKeyBytes[:32])
	return consPubKeyArray, nil
}

func connectToEthereumNetwork(rpcURL string) (*ethclient.Client, *big.Int, error) {
	url := localNodeURL
	if rpcURL != "" {
		url = rpcURL
	}

	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to Ethereum Server: %s [%v]", url, err)
	}

	networkID, err := client.NetworkID(context.Background())
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to get network ID: %v", err)
	}

	return client, networkID, nil
}

func loadContract(client *ethclient.Client) (*validatorpoolgen.Validatorpool, error) {
	contractAddress := common.HexToAddress(validatorpool.EvmAddress)
	instance, err := validatorpoolgen.NewValidatorpool(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to load contract: %v", err)
	}
	return instance, nil
}

func extractPrivateKey(cmd *cobra.Command, keyName string) (*ecdsa.PrivateKey, error) {
	clientCtx, err := clientSdk.GetClientTxContext(cmd)
	if err != nil {
		return nil, fmt.Errorf("unable to get client transaction context: %v", err)
	}

	decryptPassword := ""
	inBuf := bufio.NewReader(cmd.InOrStdin())
	if clientCtx.Keyring.Backend() == keyring.BackendFile {
		decryptPassword, err = input.GetPassword("Exporting private key. \nEnter key password:", inBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to get password: %v", err)
		}
	}

	armor, err := clientCtx.Keyring.ExportPrivKeyArmor(keyName, decryptPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to export private key: %v", err)
	}

	privKey, algo, err := crypto.UnarmorDecryptPrivKey(armor, decryptPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}
	if algo != ethsecp256k1.KeyType {
		return nil, fmt.Errorf("invalid key algorithm, got %s, expected %s", algo, ethsecp256k1.KeyType)
	}

	ethPrivKey, ok := privKey.(*ethsecp256k1.PrivKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key type %T, expected %T", privKey, &ethsecp256k1.PrivKey{})
	}

	return ethPrivKey.ToECDSA()
}

func submitApplication(instance *validatorpoolgen.Validatorpool, auth *bind.TransactOpts, consPubKeyArray [32]byte, description validatorpoolgen.Description) error {
	tx, err := instance.SubmitApplication(auth, consPubKeyArray, description)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	return nil
}
