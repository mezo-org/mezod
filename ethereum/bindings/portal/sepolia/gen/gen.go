package gen

import (
	_ "embed"
	"strings"
)

// TODO: The `make gen_address` command is done on purpose to generate only contract addresses
//       for Sepolia. See explanation in the `ethereum/bindings/portal/portal.go` file.

//go:generate make gen_address

var (
	//go:embed _address/MezoBridge
	mezoBridgeAddressFileContent string

	// MezoBridgeAddress is the MezoBridge contract's address read from the NPM package.
	MezoBridgeAddress = strings.TrimSpace(mezoBridgeAddressFileContent)
)
