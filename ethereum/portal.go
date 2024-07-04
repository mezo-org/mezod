package ethereum

import (
	"fmt"

	"github.com/evmos/evmos/v12/ethereum/bindings/portal/gen/contract"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
)

// Definitions of contract names.
const (
	PortalContractName = "Portal"
)

// PortalChain represents a Mezo portal chain handle.
type PortalChain struct {
	*baseChain

	portal *contract.Portal
}

// newPortalChain construct a new instance of the Mezo portal Ethereum chain
// handle.
func newPortalChain(
	config ethereum.Config,
	baseChain *baseChain,
) (*PortalChain, error) {
	portalAddress, err := config.ContractAddress(PortalContractName)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to resolve %s contract address: [%v]",
			PortalContractName,
			err,
		)
	}

	portal, err := contract.NewPortal(
		portalAddress,
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
		baseChain: baseChain,
		portal:    portal,
	}, nil
}
