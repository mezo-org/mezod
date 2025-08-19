package gen

import (
	_ "embed"
	"strings"
)

// TODO: The `make gen_address` command is done on purpose to generate only contract addresses
//       for Sepolia. See explanation in the `ethereum/bindings/portal/portal.go` file.

// TODO for TET-1197: Switch back to `make gen_address` once we have a stable release.
//                    Remember to remove the sepolia/gen/abi and sepolia/gen/contract
//                    directories as well.
//go:generate make

var (
	//go:embed _address/MezoBridge
	mezoBridgeAddressFileContent string

	// MezoBridgeAddress is the MezoBridge contract's address read from the NPM package.
	MezoBridgeAddress = strings.TrimSpace(mezoBridgeAddressFileContent)
)
