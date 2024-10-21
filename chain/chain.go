package chain

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/mezo-org/mezod/utils"
)

//go:embed all:mainnet
//go:embed all:testnet
var fs embed.FS

const (
	mainnetPath = "mainnet"
	testnetPath = "testnet"
	genesisFile = "genesis.json"
	seedsFile   = "seeds.txt"
)

// Config represents the chain config like genesis file and seeds.
type Config struct {
	Genesis *genutiltypes.AppGenesis
	Seeds   []string
}

func (c Config) Exists() bool {
	return c.Genesis != nil && len(c.Seeds) > 0
}

// LoadConfig loads the predefined config for the given chain ID.
// It returns true if the config was found, false otherwise.
func LoadConfig(chainID string) (Config, error) {
	var baseDir string

	//nolint:gocritic
	if utils.IsMainnet(chainID) {
		baseDir = mainnetPath
		// TODO: Remove panic once mainnet is supported.
		panic("mainnet is not supported yet")
	} else if utils.IsTestnet(chainID) {
		baseDir = testnetPath
	} else {
		return Config{}, fmt.Errorf("chain-id %s is not supported", chainID)
	}

	chainDir := fmt.Sprintf("%s/%s", baseDir, chainID)

	_, err := fs.ReadDir(chainDir)
	if err != nil {
		// If there is an error here, it means the chain directory does not exist.
		return Config{}, nil
	}

	filePath := func(file string) string {
		return fmt.Sprintf("%s/%s", chainDir, file)
	}

	genesisBytes, err := fs.ReadFile(filePath(genesisFile))
	if err != nil {
		return Config{}, fmt.Errorf("failed to read genesis file: %w", err)
	}

	genesis, err := genutiltypes.AppGenesisFromReader(
		bufio.NewReader(bytes.NewReader(genesisBytes)),
	)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse genesis file: %w", err)
	}

	seedsBytes, err := fs.ReadFile(filePath(seedsFile))
	if err != nil {
		return Config{}, fmt.Errorf("failed to read seeds file: %w", err)
	}

	var seeds []string
	seedsScanner := bufio.NewScanner(bytes.NewReader(seedsBytes))
	for seedsScanner.Scan() {
		seeds = append(seeds, seedsScanner.Text())
	}

	if err := seedsScanner.Err(); err != nil {
		return Config{}, fmt.Errorf("failed to parse seeds file: %w", err)
	}

	return Config{
		Genesis: genesis,
		Seeds:   seeds,
	}, nil
}

