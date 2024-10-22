package sidecar

import (
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
)

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
