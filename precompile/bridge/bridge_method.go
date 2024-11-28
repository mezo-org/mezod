package bridge

import (
	"fmt"

	"github.com/mezo-org/mezod/precompile"
)

// BridgeMethodName is the name of the bridge method. It matches the name
// of the method in the contract ABI.
const BridgeMethodName = "bridge"

// BridgeMethod is the implementation of the bridge method that is used to
// enable asset bridging observability in tools such as block explorers.
// nolint:revive // `BridgeMethod` is intentionally named this way for clarity.
type BridgeMethod struct{}

func newBridgeMethod() *BridgeMethod {
	return &BridgeMethod{}
}

func (bm *BridgeMethod) MethodName() string {
	return BridgeMethodName
}

func (bm *BridgeMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (bm *BridgeMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (bm *BridgeMethod) Payable() bool {
	return false
}

func (bm *BridgeMethod) Run(_ *precompile.RunContext, _ precompile.MethodInputs) (precompile.MethodOutputs, error) {
	// The bridge method is only used to enable bridging assets observability
	// in tools such as block explorers.
	return precompile.MethodOutputs{false}, fmt.Errorf(
		"bridge action is done by the supermajority of validators and cannot " +
			"be called directly",
	)
}
