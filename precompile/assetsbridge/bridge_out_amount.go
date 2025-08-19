package assetsbridge

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// SetMinBridgeOutAmountMethodName is the name of the setMinBridgeOutAmount
// method. It matches the name of the method in the contract ABI.
//
//nolint:gosec
const SetMinBridgeOutAmountMethodName = "setMinBridgeOutAmount"

// SetMinBridgeOutAmountMethod is the implementation of the setMinBridgeOutAmount
// method.
type SetMinBridgeOutAmountMethod struct {
	poaKeeper    PoaKeeper
	bridgeKeeper BridgeKeeper
}

func newSetMinBridgeOutAmountMethod(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
) *SetMinBridgeOutAmountMethod {
	return &SetMinBridgeOutAmountMethod{
		poaKeeper:    poaKeeper,
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *SetMinBridgeOutAmountMethod) MethodName() string {
	return SetMinBridgeOutAmountMethodName
}

func (m *SetMinBridgeOutAmountMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *SetMinBridgeOutAmountMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *SetMinBridgeOutAmountMethod) Payable() bool {
	return false
}

func (m *SetMinBridgeOutAmountMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	mezoToken, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("mezo token must be common.Address")
	}

	minAmount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid minimum amount: %v", inputs[1])
	}

	if minAmount == nil {
		return nil, fmt.Errorf("minimum amount is required")
	}

	if minAmount.Sign() < 0 {
		return nil, fmt.Errorf("minimum amount cannot be negative")
	}

	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	err = m.bridgeKeeper.SetMinBridgeOutAmount(
		context.SdkCtx(),
		mezoToken.Bytes(),
		sdkmath.NewIntFromBigInt(minAmount),
	)
	if err != nil {
		return nil, err
	}

	err = context.EventEmitter().Emit(
		NewMinBridgeOutAmountSetEvent(mezoToken, minAmount),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to emit MinBridgeOutAmountSet event: [%w]",
			err,
		)
	}

	return precompile.MethodOutputs{true}, nil
}

// MinBridgeOutAmountSetEventName is the name of the MinBridgeOutAmountSet event.
// It matches the name of the event in the contract ABI.
//
//nolint:gosec
const MinBridgeOutAmountSetEventName = "MinBridgeOutAmountSet"

// MinBridgeOutAmountSetEvent is the implementation of the MinBridgeOutAmountSet
// event that contains the following arguments:
// - token (indexed): the address of the token on Mezo chain.
// - minAmount (non-indexed): the new minimum bridgeable amount for the token.
type MinBridgeOutAmountSetEvent struct {
	mezoToken common.Address
	minAmount *big.Int
}

func NewMinBridgeOutAmountSetEvent(
	mezoToken common.Address,
	minAmount *big.Int,
) *MinBridgeOutAmountSetEvent {
	return &MinBridgeOutAmountSetEvent{
		mezoToken: mezoToken,
		minAmount: minAmount,
	}
}

func (e *MinBridgeOutAmountSetEvent) EventName() string {
	return MinBridgeOutAmountSetEventName
}

func (e *MinBridgeOutAmountSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.mezoToken,
		},
		{
			Indexed: false,
			Value:   e.minAmount,
		},
	}
}
