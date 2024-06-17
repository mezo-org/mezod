package precompile

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

var TypesConverter = struct {
	Address addressConverter
	BigInt  bigIntConverter
}{
	Address: addressConverter{},
	BigInt:  bigIntConverter{},
}

type addressConverter struct{}

func (ac addressConverter) ToSDK(address common.Address) sdk.AccAddress {
	return address.Bytes()
}

func (ac addressConverter) FromSDK(address sdk.AccAddress) common.Address {
	return common.BytesToAddress(address)
}

type bigIntConverter struct{}

func (bic bigIntConverter) ToSDK(value *big.Int) (sdkmath.Int, error) {
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

func (bic bigIntConverter) FromSDK(value sdkmath.Int) *big.Int {
	return value.BigInt()
}
