package maintenance

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetMinGasPriceMethodName is the name of the setMinGasPrice method.
// It matches the name of the method in the contract ABI.
const SetMinGasPriceMethodName = "setMinGasPrice"

type setMinGasPriceMethod struct {
	poaKeeper       PoaKeeper
	feeMarketKeeper FeeMarketKeeper
}

func newSetMinGasPriceMethod(poaKeeper PoaKeeper, feeMarketKeeper FeeMarketKeeper) *setMinGasPriceMethod {
	return &setMinGasPriceMethod{
		poaKeeper:       poaKeeper,
		feeMarketKeeper: feeMarketKeeper,
	}
}

func (m *setMinGasPriceMethod) MethodName() string {
	return SetMinGasPriceMethodName
}

func (m *setMinGasPriceMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setMinGasPriceMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setMinGasPriceMethod) Payable() bool {
	return false
}

func (m *setMinGasPriceMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, nil, err
	}

	minGasPrice, ok := inputs[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("invalid min gas price type")
	}

	// This method is restricted to the validator pool owner
	err := m.poaKeeper.CheckOwner(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	if minGasPrice.Sign() <= 0 {
		return nil, nil, fmt.Errorf("min gas price must be positive")
	}

	params := m.feeMarketKeeper.GetParams(context.SdkCtx())
	params.MinGasPrice = sdkmath.LegacyNewDecFromBigInt(minGasPrice)
	err = m.feeMarketKeeper.SetParams(context.SdkCtx(), params)
	if err != nil {
		return nil, nil, err
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// GetMinGasPriceMethodName is the name of the getMinGasPrice method.
// It matches the name of the method in the contract ABI.
const GetMinGasPriceMethodName = "getMinGasPrice"

type getMinGasPriceMethod struct {
	feeMarketKeeper FeeMarketKeeper
}

func newGetMinGasPriceMethod(feeMarketKeeper FeeMarketKeeper) *getMinGasPriceMethod {
	return &getMinGasPriceMethod{
		feeMarketKeeper: feeMarketKeeper,
	}
}

func (m *getMinGasPriceMethod) MethodName() string {
	return GetMinGasPriceMethodName
}

func (m *getMinGasPriceMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getMinGasPriceMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *getMinGasPriceMethod) Payable() bool {
	return false
}

func (m *getMinGasPriceMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	params := m.feeMarketKeeper.GetParams(context.SdkCtx())

	return precompile.MethodOutputs{params.MinGasPrice.TruncateInt().BigInt()}, nil, nil
}
