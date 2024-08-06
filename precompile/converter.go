package precompile

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// TypesConverter is a singleton that provides functions to convert types
// between EVM and Cosmos SDK.
var TypesConverter = struct {
	// Address provides functions to convert between EVM and Cosmos SDK addresses.
	Address addressConverter
	// BigInt provides functions to convert between EVM and Cosmos SDK big integers.
	BigInt bigIntConverter
}{
	Address: addressConverter{},
	BigInt:  bigIntConverter{},
}

type addressConverter struct{}

// ToSDK converts the given EVM address to the Cosmos SDK address.
func (ac addressConverter) ToSDK(address common.Address) sdk.AccAddress {
	return address.Bytes()
}

// FromSDK converts the given Cosmos SDK address to the EVM address.
func (ac addressConverter) FromSDK(address sdk.AccAddress) common.Address {
	return common.BytesToAddress(address)
}

type bigIntConverter struct{}

// ToSDK converts the given big integer to the Cosmos SDK integer.
func (bic bigIntConverter) ToSDK(value *big.Int) (sdkmath.Int, error) {
	// Validate the value's bit length against the maximum bit length
	// supported by the SDK. Otherwise, the sdk.NewIntFromBigInt may panic.
	if value.BitLen() > sdkmath.MaxBitLen {
		return sdkmath.Int{}, fmt.Errorf(
			"value is exceeding the maximum bit length: [%d]",
			sdkmath.MaxBitLen,
		)
	}

	return sdkmath.NewIntFromBigInt(value), nil
}

// FromSDK converts the given Cosmos SDK integer to the big integer.
func (bic bigIntConverter) FromSDK(value sdkmath.Int) *big.Int {
	return value.BigInt()
}
