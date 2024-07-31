package btctoken

import (
	"bytes"
	"fmt"

	sdkmath "cosmossdk.io/math"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
)

// AllowanceMethodName is the name of the Allowance method that should match the
// name in the contract ABI.
const (
	AllowanceMethodName = "allowance"
)

type allowanceMethod struct {
	authzkeeper authzkeeper.Keeper
}

func newAllowanceMethod(
	authzkeeper authzkeeper.Keeper,
) *allowanceMethod {
	return &allowanceMethod{
		authzkeeper: authzkeeper,
	}
}

func (am *allowanceMethod) MethodName() string {
	return AllowanceMethodName
}

func (am *allowanceMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (am *allowanceMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (am *allowanceMethod) Payable() bool {
	return false
}

// Run returns the allowance of the spender for the owner.
func (am *allowanceMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	owner, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid owner address: %v", inputs[0])
	}

	if isZeroAddress(owner) {
		return nil, fmt.Errorf("owner address cannot be empty")
	}

	spender, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid spender address: %v", inputs[1])
	}

	if isZeroAddress(spender) {
		return nil, fmt.Errorf("spender address cannot be empty")
	}

	// Return the max uint256 when the owner and spender are the same.
	if bytes.Equal(owner.Bytes(), spender.Bytes()) {
		return precompile.MethodOutputs{abi.MaxUint256}, nil
	}

	var allowance sdkmath.Int
	authorization, _ := am.authzkeeper.GetAuthorization(context.SdkCtx(), precompile.TypesConverter.Address.ToSDK(spender), precompile.TypesConverter.Address.ToSDK(owner), SendMsgURL)
	if authorization == nil {
		allowance = sdkmath.ZeroInt()
	} else {
		sendAuth, ok := authorization.(*banktypes.SendAuthorization)
		if !ok {
			return nil, fmt.Errorf(
				"expected authorization to be a %T", banktypes.SendAuthorization{},
			)
		}

		allowance = sendAuth.SpendLimit.AmountOfNoDenomValidation(evm.DefaultEVMDenom)
	}

	return precompile.MethodOutputs{
		precompile.TypesConverter.BigInt.FromSDK(allowance),
	}, nil
}
