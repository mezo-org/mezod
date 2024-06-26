package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "poa"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for querier msgs
	QuerierRoute = ModuleName
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
)

// Get the key for the validator with address
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsKey, operatorAddr.Bytes()...)
}

// Get the key for the validator with pubkey
func GetValidatorByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ValidatorsByConsAddrKey, addr.Bytes()...)
}

// Get the key for the validator state with address
func GetValidatorStateKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorStatesKey, operatorAddr.Bytes()...)
}

// Get the key for a validator canditate application with address
func GetApplicationKey(operatorAddr sdk.ValAddress) []byte {
	return append(ApplicationPoolKey, operatorAddr.Bytes()...)
}

// Get the key for a validator canditate application with pubkey
func GetApplicationByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ApplicationByConsAddrKey, addr.Bytes()...)
}

// Get the key for a validator canditate application with pubkey
func GetKickProposalKey(operatorAddr sdk.ValAddress) []byte {
	return append(KickProposalPoolKey, operatorAddr.Bytes()...)
}
