package gen

import (
	_ "embed"
	"strings"
)

//go:generate make

var (
	//go:embed _address/Portal
	portalAddressFileContent string

	// PortalAddress is a Portal contract's address read from the NPM package.
	PortalAddress = strings.TrimSpace(portalAddressFileContent)
)
