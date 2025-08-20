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

	if (mezoToken == common.Address{}) {
		return nil, fmt.Errorf("mezo token cannot be the zero address")
	}

	minAmount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid minimum amount: %v", inputs[1])
	}

	if minAmount == nil {
		return nil, fmt.Errorf("minimum amount is required")
	}

	if minAmount.Sign() <= 0 {
		return nil, fmt.Errorf("minimum amount must be negative")
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
	token     common.Address
	minAmount *big.Int
}

func NewMinBridgeOutAmountSetEvent(
	token common.Address,
	minAmount *big.Int,
) *MinBridgeOutAmountSetEvent {
	return &MinBridgeOutAmountSetEvent{
		token:     token,
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
			Value:   e.token,
		},
		{
			Indexed: false,
			Value:   e.minAmount,
		},
	}
}

// GetMinBridgeOutAmountMethodName is the name of the getMinBridgeOutAmount
// method. It matches the name of the method in the contract ABI.
//
//nolint:gosec
const GetMinBridgeOutAmountMethodName = "getMinBridgeOutAmount"

// GetMinBridgeOutAmountMethod is the implementation of the getMinBridgeOutAmount
// method.
type GetMinBridgeOutAmountMethod struct {
	bridgeKeeper BridgeKeeper
}

func newGetMinBridgeOutAmountMethod(
	bridgeKeeper BridgeKeeper,
) *GetMinBridgeOutAmountMethod {
	return &GetMinBridgeOutAmountMethod{
		bridgeKeeper: bridgeKeeper,
	}
}

func (m *GetMinBridgeOutAmountMethod) MethodName() string {
	return GetMinBridgeOutAmountMethodName
}

func (m *GetMinBridgeOutAmountMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetMinBridgeOutAmountMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetMinBridgeOutAmountMethod) Payable() bool {
	return false
}

func (m *GetMinBridgeOutAmountMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	token, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("token must be common.Address")
	}

	minAmount, found := m.bridgeKeeper.GetMinBridgeOutAmount(
		context.SdkCtx(),
		token.Bytes(),
	)

	if !found {
		return precompile.MethodOutputs{
			big.NewInt(0),
			false,
		}, nil
	}

	return precompile.MethodOutputs{
		minAmount.BigInt(),
		true,
	}, nil
}
