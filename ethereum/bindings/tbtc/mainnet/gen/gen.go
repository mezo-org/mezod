package gen

import (
	_ "embed"
	"strings"
)

// Bindings for tBTC Bridge were generated manually - no `go:generate` needed.

var (
	//go:embed _address/Bridge
	tbtcBridgeAddressFileContent string
	TbtcBridgeAddress            = strings.TrimSpace(tbtcBridgeAddressFileContent)
)
