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
	ParamsKey = []byte{0x10} // standalone key for module params

	OwnerKey          = []byte{0x20} // standalone key for the owner of the validator pool
	CandidateOwnerKey = []byte{0x21} // standalone key for the candidate owner of the validator pool

	ApplicationKeyPrefix           = []byte{0x30} // prefix for each key to a validator application
	ApplicationByConsAddrKeyPrefix = []byte{0x31} // prefix for each key to a validator application index, by consensus address

	ValidatorKeyPrefix             = []byte{0x40} // prefix for each key to a validator
	ValidatorByConsAddrKeyPrefix   = []byte{0x41} // prefix for each key to a validator index, by consensus address
	ValidatorStateKeyPrefix        = []byte{0x42} // prefix for each key to a validator state
	ValidatorsByPrivilegeKeyPrefix = []byte{0x43} // prefix for each key to a validators sub-set, by privilege

	HistoricalInfoKeyPrefix = []byte{0x50} // prefix for each key to a historical info
)

// GetApplicationKey gets the key for a validator application by operator address.
func GetApplicationKey(operator sdk.ValAddress) []byte {
	return append(ApplicationKeyPrefix, operator.Bytes()...)
}

// GetApplicationByConsAddrKey gets the key for a validator application by consensus address.
func GetApplicationByConsAddrKey(cons sdk.ConsAddress) []byte {
	return append(ApplicationByConsAddrKeyPrefix, cons.Bytes()...)
}

// GetValidatorKey gets the key for the validator by operator address.
func GetValidatorKey(operator sdk.ValAddress) []byte {
	return append(ValidatorKeyPrefix, operator.Bytes()...)
}

// GetValidatorByConsAddrKey gets the key for the validator by consensus address.
func GetValidatorByConsAddrKey(cons sdk.ConsAddress) []byte {
	return append(ValidatorByConsAddrKeyPrefix, cons.Bytes()...)
}

// GetValidatorStateKey gets the key for the validator state by operator address.
func GetValidatorStateKey(operator sdk.ValAddress) []byte {
	return append(ValidatorStateKeyPrefix, operator.Bytes()...)
}

// GetValidatorsByPrivilegeKey gets the key for a validators sub-set having the given privilege.
// A value is a set of consensus addresses of validators that have the given privilege.
// We use the consensus address not the operator address because consensus-address-based
// lookups occur more frequently than operator-address-based lookups.
func GetValidatorsByPrivilegeKey(privilege string) []byte {
	return append(ValidatorsByPrivilegeKeyPrefix, []byte(privilege)...)
}

// GetHistoricalInfoKey gets the key for a historical info by height.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKeyPrefix, []byte(strconv.FormatInt(height, 10))...)
}
