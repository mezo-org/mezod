package precompile

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type AddressConverter struct {}

func (ac AddressConverter) ToSDK(address common.Address) sdk.AccAddress {
	return address.Bytes()
}

func (ac AddressConverter) FromSDK(address sdk.AccAddress) common.Address {
	return common.BytesToAddress(address)
}

type BigIntConverter struct {}

func (bic BigIntConverter) ToSDK(value *big.Int) (sdkmath.Int, error){
	// Validate the value's bit length against the maximum bit length
	// supported by the SDK. Otherwise, the sdk.NewIntFromBigInt may panic.
	if value.BitLen() > sdk.MaxBitLen {
		return sdkmath.Int{}, fmt.Errorf(
			"value is exceeding the maximum bit length: [%d]",
			sdk.MaxBitLen,
		)
	}

	return sdk.NewIntFromBigInt(value), nil
}

func (bic BigIntConverter) FromSDK(value sdkmath.Int) *big.Int {
	return value.BigInt()
}