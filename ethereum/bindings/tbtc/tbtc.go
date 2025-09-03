package tbtc

import (
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	mainnetgen "github.com/mezo-org/mezod/ethereum/bindings/tbtc/mainnet/gen"
	sepoliagen "github.com/mezo-org/mezod/ethereum/bindings/tbtc/sepolia/gen"
	sepoliacontract "github.com/mezo-org/mezod/ethereum/bindings/tbtc/sepolia/gen/contract"
)

// TODO: Use sepolia bindings for now.
//       In the future implement a mechanism allowing automatic selection of
//       bindings based on network.

func BridgeAddress(network ethconfig.Network) string {
	switch network {
	case ethconfig.Sepolia:
		return sepoliagen.TbtcBridgeAddress
	case ethconfig.Mainnet:
		return mainnetgen.TbtcBridgeAddress
	default:
		panic("unknown ethereum network")
	}
}

type Bridge = sepoliacontract.Bridge

var NewTbtcBridge = sepoliacontract.NewBridge
