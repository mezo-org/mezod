package gen

import (
	_ "embed"
	"strings"
)

// Bindings are committed; prevent `go generate` from rebuilding here.
//go:generate true

var (
	//go:embed _address/Bridge
	tbtcBridgeAddressFileContent string
	TbtcBridgeAddress            = strings.TrimSpace(tbtcBridgeAddressFileContent)
)
