package sidecar

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
)

var bitcoinBridgeAddressPath = "ethereum/bindings/portal/gen/_address/BitcoinBridge"

func networkFromString(networkStr string) ethconfig.Network {
	switch networkStr {
	case ethconfig.Mainnet.String():
		return ethconfig.Mainnet
	case ethconfig.Sepolia.String():
		return ethconfig.Sepolia
	case ethconfig.Developer.String():
		return ethconfig.Developer
	default:
		return ethconfig.Unknown
	}
}

func readBitcoinBridgeAddress() (common.Address, error) {
	// Read the file content
	content, err := os.ReadFile(bitcoinBridgeAddressPath)
	if err != nil {
		return common.Address{}, err
	}

	// Remove any leading/trailing whitespace and convert to a string
	addressStr := strings.TrimSpace(string(content))

	return common.HexToAddress(addressStr), nil
}
