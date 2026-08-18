package assetsbridge

import (
	"fmt"

	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// RemoveBridgeOutChainMethodName is the name of the removeBridgeOutChain
// method. It matches the name of the method in the contract ABI.
const RemoveBridgeOutChainMethodName = "removeBridgeOutChain"

// RemoveBridgeOutChainMethod is the implementation of the removeBridgeOutChain
// method.
type RemoveBridgeOutChainMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newRemoveBridgeOutChainMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *RemoveBridgeOutChainMethod {
	return &RemoveBridgeOutChainMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *RemoveBridgeOutChainMethod) MethodName() string {
	return RemoveBridgeOutChainMethodName
}

func (m *RemoveBridgeOutChainMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *RemoveBridgeOutChainMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *RemoveBridgeOutChainMethod) Payable() bool {
	return false
}

func (m *RemoveBridgeOutChainMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	chain, err := extractBridgeOutChainInput(inputs)
	if err != nil {
		return nil, nil, err
	}

	err = m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	if !m.bridgeKeeper.IsBridgeOutChainEnabled(context.SdkCtx(), chain) {
		return nil, nil, fmt.Errorf("chain is not enabled for bridge-outs")
	}

	m.bridgeKeeper.DisableBridgeOutChain(context.SdkCtx(), chain)

	err = context.EventEmitter().Emit(
		NewBridgeOutChainRemovedEvent(chain),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to emit BridgeOutChainRemoved event: [%w]",
			err,
		)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// AddBridgeOutChainMethodName is the name of the addBridgeOutChain method.
// It matches the name of the method in the contract ABI.
const AddBridgeOutChainMethodName = "addBridgeOutChain"

// AddBridgeOutChainMethod is the implementation of the addBridgeOutChain
// method.
type AddBridgeOutChainMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newAddBridgeOutChainMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *AddBridgeOutChainMethod {
	return &AddBridgeOutChainMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *AddBridgeOutChainMethod) MethodName() string {
	return AddBridgeOutChainMethodName
}

func (m *AddBridgeOutChainMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *AddBridgeOutChainMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *AddBridgeOutChainMethod) Payable() bool {
	return false
}

func (m *AddBridgeOutChainMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	chain, err := extractBridgeOutChainInput(inputs)
	if err != nil {
		return nil, nil, err
	}

	err = m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	if m.bridgeKeeper.IsBridgeOutChainEnabled(context.SdkCtx(), chain) {
		return nil, nil, fmt.Errorf("chain is already enabled for bridge-outs")
	}

	m.bridgeKeeper.EnableBridgeOutChain(context.SdkCtx(), chain)

	err = context.EventEmitter().Emit(
		NewBridgeOutChainAddedEvent(chain),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to emit BridgeOutChainAdded event: [%w]",
			err,
		)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// GetBridgeOutChainsMethodName is the name of the getBridgeOutChains method.
// It matches the name of the method in the contract ABI.
const GetBridgeOutChainsMethodName = "getBridgeOutChains"

// GetBridgeOutChainsMethod is the implementation of the getBridgeOutChains
// method.
type GetBridgeOutChainsMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetBridgeOutChainsMethod(
	bridgeKeeper BridgeKeeper,
) *GetBridgeOutChainsMethod {
	return &GetBridgeOutChainsMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetBridgeOutChainsMethod) MethodName() string {
	return GetBridgeOutChainsMethodName
}

func (m *GetBridgeOutChainsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetBridgeOutChainsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetBridgeOutChainsMethod) Payable() bool {
	return false
}

func (m *GetBridgeOutChainsMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	chains := m.bridgeKeeper.GetBridgeOutChains(context.SdkCtx())

	return precompile.MethodOutputs{chains}, nil, nil
}

// extractBridgeOutChainInput extracts the single target chain input and checks
// it against the supported target chains.
func extractBridgeOutChainInput(inputs precompile.MethodInputs) (uint8, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return 0, err
	}

	chainRaw, ok := inputs[0].(uint8)
	if !ok {
		return 0, fmt.Errorf("invalid chain: %v", inputs[0])
	}

	chain := TargetChain(chainRaw)

	if _, ok := chain.Validate(); !ok {
		return 0, fmt.Errorf("unsupported chain: %v", chain)
	}

	return chainRaw, nil
}

// BridgeOutChainRemovedEventName is the name of the BridgeOutChainRemoved
// event. It matches the name of the event in the contract ABI.
const BridgeOutChainRemovedEventName = "BridgeOutChainRemoved"

// BridgeOutChainRemovedEvent is the implementation of the BridgeOutChainRemoved
// event that contains the following arguments:
// - chain (non-indexed): the target chain removed from the set of chains
// enabled for bridge-outs.
type BridgeOutChainRemovedEvent struct {
	chain uint8
}

func NewBridgeOutChainRemovedEvent(chain uint8) *BridgeOutChainRemovedEvent {
	return &BridgeOutChainRemovedEvent{
		chain: chain,
	}
}

func (e *BridgeOutChainRemovedEvent) EventName() string {
	return BridgeOutChainRemovedEventName
}

func (e *BridgeOutChainRemovedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: false,
			Value:   e.chain,
		},
	}
}

// BridgeOutChainAddedEventName is the name of the BridgeOutChainAdded event.
// It matches the name of the event in the contract ABI.
const BridgeOutChainAddedEventName = "BridgeOutChainAdded"

// BridgeOutChainAddedEvent is the implementation of the BridgeOutChainAdded
// event that contains the following arguments:
// - chain (non-indexed): the target chain added to the set of chains enabled
// for bridge-outs.
type BridgeOutChainAddedEvent struct {
	chain uint8
}

func NewBridgeOutChainAddedEvent(chain uint8) *BridgeOutChainAddedEvent {
	return &BridgeOutChainAddedEvent{
		chain: chain,
	}
}

func (e *BridgeOutChainAddedEvent) EventName() string {
	return BridgeOutChainAddedEventName
}

func (e *BridgeOutChainAddedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: false,
			Value:   e.chain,
		},
	}
}
