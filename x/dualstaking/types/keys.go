package types

import "fmt"

const (
	ModuleName = "dualstaking"
	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_dualstaking"
)

const (
	StakingPositionPrefix = iota + 1
	DelegationPositionPrefix
)

var (
	KeyStakingPositionPrefix    = []byte{StakingPositionPrefix}
	KeyDelegationPositionPrefix = []byte{DelegationPositionPrefix}
)

func GetStakeKey(staker string, stakeId string) []byte {
	return []byte(fmt.Sprintf("staking-%s-%s", staker, stakeId))
}

func GetDelegationKey(staker string, delegationId string) []byte {
	return []byte(fmt.Sprintf("delegation-%s-%s", staker, delegationId))
}
