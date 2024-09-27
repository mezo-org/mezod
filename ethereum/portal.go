package ethereum

import (
	"fmt"

	"github.com/mezo-org/mezod/ethereum/bindings/portal/gen/contract"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
)

// Definitions of contract names.
const (
	BitcoinBridgeContractName = "BitcoinBridge"
)

// PortalChain represents a Mezo Portal chain handle.
type PortalChain struct {
	*baseChain

	bitcoinBridge *contract.BitcoinBridge
}

// newPortalChain construct a new instance of the Mezo portal Ethereum chain
// handle.
func newPortalChain(
	config ethereum.Config,
	baseChain *baseChain,
) (*PortalChain, error) {
	bitcoinBridgeAddress, err := config.ContractAddress(BitcoinBridgeContractName)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to resolve %s contract address: [%v]",
			BitcoinBridgeContractName,
			err,
		)
	}

	bitcoinBridge, err := contract.NewBitcoinBridge(
		bitcoinBridgeAddress,
		baseChain.chainID,
		baseChain.key,
		baseChain.client,
		baseChain.nonceManager,
		baseChain.miningWaiter,
		baseChain.blockCounter,
		baseChain.transactionMutex,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to attach to Portal contract: [%v]",
			err,
		)
	}

	return &PortalChain{
		baseChain:     baseChain,
		bitcoinBridge: bitcoinBridge,
	}, nil
}
