package gen

import (
	_ "embed"
	"strings"
)

//go:generate make

var (
	//go:embed _address/BitcoinBridge
	bitcoinBridgeAddressFileContent string

	// BitcoinBridgeAddress is a Bitcoin Bridge contract's address read from the NPM package.
	BitcoinBridgeAddress = strings.TrimSpace(bitcoinBridgeAddressFileContent)
)
