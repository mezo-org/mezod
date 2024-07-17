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

// Standalone keys and key prefixes for the poa module.
var (
	OwnerKey          = []byte{0x10} // standalone key for the owner of the validator pool
	CandidateOwnerKey = []byte{0x11} // standalone key for the candidate owner of the validator pool

	ApplicationKeyPrefix           = []byte{0x20} // prefix for each key to a validator application
	ApplicationByConsAddrKeyPrefix = []byte{0x21} // prefix for each key to a validator application, by consensus address

	ValidatorKeyPrefix           = []byte{0x30} // prefix for each key to a validator
	ValidatorByConsAddrKeyPrefix = []byte{0x31} // prefix for each key to a validator index, by consensus address
	ValidatorStateKeyPrefix      = []byte{0x32} // prefix for each key to a validator state

	KickProposalKeyPrefix = []byte{0x40} // prefix for each key to a kick proposal

	HistoricalInfoKeyPrefix = []byte{0x50} // prefix for each key to a historical info
)

// GetApplicationKey gets the key for a validator application by operator address.
func GetApplicationKey(operatorAddr sdk.ValAddress) []byte {
	return append(ApplicationKeyPrefix, operatorAddr.Bytes()...)
}

// GetApplicationByConsAddrKey gets the key for a validator application by consensus address.
func GetApplicationByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ApplicationByConsAddrKeyPrefix, addr.Bytes()...)
}

// GetValidatorKey gets the key for the validator by operator address.
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorKeyPrefix, operatorAddr.Bytes()...)
}

// GetValidatorByConsAddrKey gets the key for the validator by consensus address.
func GetValidatorByConsAddrKey(consAddr sdk.ConsAddress) []byte {
	return append(ValidatorByConsAddrKeyPrefix, consAddr.Bytes()...)
}

// GetValidatorStateKey gets the key for the validator state by operator address.
func GetValidatorStateKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorStateKeyPrefix, operatorAddr.Bytes()...)
}

// GetKickProposalKey gets the key for a kick proposal by operator address.
func GetKickProposalKey(operatorAddr sdk.ValAddress) []byte {
	return append(KickProposalKeyPrefix, operatorAddr.Bytes()...)
}

// GetHistoricalInfoKey gets the key for a historical info by height.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKeyPrefix, []byte(strconv.FormatInt(height, 10))...)
}
