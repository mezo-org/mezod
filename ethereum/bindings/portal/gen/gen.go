package gen

import (
	_ "embed"
	"strings"
)

//go:generate make

var (
	//go:embed _address/MezoBridge
	mezoBridgeAddressFileContent string

	// MezoBridgeAddress is the MezoBridge contract's address read from the NPM package.
	MezoBridgeAddress = strings.TrimSpace(mezoBridgeAddressFileContent)
)
