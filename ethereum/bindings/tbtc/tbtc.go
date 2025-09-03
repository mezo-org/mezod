package tbtc

import (
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	mainnetgen "github.com/mezo-org/mezod/ethereum/bindings/tbtc/mainnet/gen"
	mainnetcontract "github.com/mezo-org/mezod/ethereum/bindings/tbtc/mainnet/gen/contract"
	sepoliagen "github.com/mezo-org/mezod/ethereum/bindings/tbtc/sepolia/gen"
)

// The tBTC bindings were generated manually with the following commands:
//
// 1) Make binding directories:
//    mkdir -p ethereum/bindings/tbtc/mainnet/gen/{abi,contract,_address}
//
// 2) Download the package and extract it:
//    mkdir -p tmp/contracts/@keep-network/tbtc-v2@1.8.0
//    cd tmp/contracts/@keep-network/tbtc-v2@1.8.0
//
//    npm pack --silent @keep-network/tbtc-v2@1.8.0
//    mkdir -p extracted
//    tar -xzf keep-network-tbtc-v2-1.8.0.tgz -C extracted
//
//    cd ../../../../     <- Go back to the root directory
//
// 3) Enable module writing:
//    go env -w GOFLAGS=-mod=mod
//
// 4) Generate bindings:
//    ABI_JSON="tmp/contracts/@keep-network/tbtc-v2@1.8.0/extracted/package/artifacts/Bridge.json"
//    BIND_ROOT="ethereum/bindings/tbtc/mainnet/gen"
//
//    jq .abi "$ABI_JSON" > "$BIND_ROOT/abi/Bridge.abi"
//
//    go run github.com/ethereum/go-ethereum/cmd/abigen --abi "$BIND_ROOT/abi/Bridge.abi" --pkg abi --type Bridge --out "$BIND_ROOT/abi/Bridge.go"
//
//    go run github.com/keep-network/keep-common/tools/generators/ethereum "$BIND_ROOT/abi/Bridge.abi" "$BIND_ROOT/contract/Bridge.go"
//
//    cp "$ABI_JSON" "$BIND_ROOT/abi/Bridge.json"
//
// 5) Extract contract address:
//    jq -jr .address "$ABI_JSON" > "$BIND_ROOT/_address/Bridge"
//
// 6) Create the `ethereum/bindings/tbtc/mainnet/gen/gen.go` file.
//
// 7) Restore Go flags:
//    go env -u GOFLAGS
//    go mod tidy

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

// Use mainnet bindings.
type Bridge = mainnetcontract.Bridge

var NewTbtcBridge = mainnetcontract.NewBridge
