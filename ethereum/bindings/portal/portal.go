package portal

import (
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	mainnetgen "github.com/mezo-org/mezod/ethereum/bindings/portal/mainnet/gen"
	mainnetabi "github.com/mezo-org/mezod/ethereum/bindings/portal/mainnet/gen/abi"
	sepoliagen "github.com/mezo-org/mezod/ethereum/bindings/portal/sepolia/gen"
)

// TODO: The current bindings implementation uses a simplified approach for handling multiple Ethereum networks.
//       A more robust solution would involve creating a facade component within the `portal` package
//       that dynamically loads the appropriate contract bindings at runtime, abstracting this complexity
//       from client code.
//
//       Currently, we're using Mainnet contract bindings for both Sepolia and Mainnet networks in the
//       `mezod/ethereum` package. While this introduces some technical debt, it's an intentional trade-off
//       since the contract interfaces are currently identical across networks. The only difference is in the
//       contract addresses, which are properly configured per environment.
//
//       As long as this approach is in place, we optimize the binary size by generating bindings only for Mainnet,
//       while Sepolia bindings are limited to address definitions.
//
//       This approach is sustainable as long as there are no differences between Mainnet and Sepolia contracts.
//       Once such differences emerge, we need to revisit it.

func MezoBridgeAddress(network ethconfig.Network) string {
	switch network {
	case ethconfig.Sepolia:
		return sepoliagen.MezoBridgeAddress
	case ethconfig.Mainnet:
		return mainnetgen.MezoBridgeAddress
	default:
		panic("unknown ethereum network")
	}
}

type (
	MezoBridge                     = mainnetabi.MezoBridge
	MezoBridgeAssetsLocked         = mainnetabi.MezoBridgeAssetsLocked
	MezoBridgeAssetsLockedIterator = mainnetabi.MezoBridgeAssetsLockedIterator
)

var NewMezoBridge = mainnetabi.NewMezoBridge
