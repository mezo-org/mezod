package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "poa"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName
)

var (
	// Prefix for each key to a validator
	ValidatorsKey = []byte{0x21}

	// Prefix for each key to a validator index, by pubkey
	ValidatorsByConsAddrKey = []byte{0x22}

	// Prefix for each key to a validator state index
	ValidatorStatesKey = []byte{0x23}

	// Prefix for the validator application pool
	ApplicationPoolKey = []byte{0x24}

	// Prefix for each key to a application index, by pubkey
	ApplicationByConsAddrKey = []byte{0x25}

	// Prefix for the validator kick proposal pool
	KickProposalPoolKey = []byte{0x26}
	// Prefix for the historical info
	HistoricalInfoKey = []byte{0x27}
)

// Get the key for the validator by operator address
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsKey, operatorAddr.Bytes()...)
}

// Get the key for the validator by consensus address
func GetValidatorByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ValidatorsByConsAddrKey, addr.Bytes()...)
}

// Get the key for the validator state by operator address
func GetValidatorStateKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorStatesKey, operatorAddr.Bytes()...)
}

// Get the key for a validator candidate application by operator address
func GetApplicationKey(operatorAddr sdk.ValAddress) []byte {
	return append(ApplicationPoolKey, operatorAddr.Bytes()...)
}

// Get the key for a validator candidate application by consensus address
func GetApplicationByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ApplicationByConsAddrKey, addr.Bytes()...)
}

// Get the key for a kick proposal by operator address
func GetKickProposalKey(operatorAddr sdk.ValAddress) []byte {
	return append(KickProposalPoolKey, operatorAddr.Bytes()...)
}

// GetHistoricalInfoKey returns a key prefix for indexing HistoricalInfo objects.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKey, []byte(strconv.FormatInt(height, 10))...)
}
