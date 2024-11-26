package precompile

import "github.com/ethereum/go-ethereum/common"

// VersionMap is a map of precompile versions.
type VersionMap struct {
	versions map[int]*Contract
	rules    func(height int64) int
}

// NewSingleVersionMap creates a new VersionMap with a single contract.
// This should be the default for most cases.
func NewSingleVersionMap(contract *Contract) *VersionMap {
	return newVersionMap([]*Contract{contract}, nil)
}

// NewMultiVersionMap creates a new VersionMap with multiple contracts.
// Order of the passed contracts is important. The first contract is version 1,
// the second contract is version 2, and so on. The rules function should return
// the version number based on the height.
//
// BEWARE: Use this only when you know what you are doing. The typical use
// cases is ensuring backwards compatibility when breaking changes
// are introduced to the precompile. Double-check if you are not
// introducing a consensus-level regression by the way.
func NewMultiVersionMap(
	contracts []*Contract,
	rules func(height int64) int,
) *VersionMap {
	if len(contracts) < 2 {
		panic("multiple contracts must be provided")
	}

	if rules == nil {
		panic("rules function must be provided")
	}

	return newVersionMap(contracts, rules)
}

func newVersionMap(
	contracts []*Contract,
	rules func(height int64) int,
) *VersionMap {
	if len(contracts) == 0 {
		panic("no contracts provided")
	}

	versions := make(map[int]*Contract)
	address := contracts[0].Address()

	for i, contract := range contracts {
		if address != contract.Address() {
			panic("all contracts must have the same address")
		}

		versions[i+1] = contract
	}

	return &VersionMap{
		versions: versions,
		rules:    rules,
	}
}

// GetByHeight returns the precompile version based on the chain height.
func (v *VersionMap) GetByHeight(height int64) *Contract {
	if v.rules != nil {
		version := v.rules(height)
		return v.GetByVersion(version)
	}

	return v.GetLatest()
}

// GetByVersion returns the precompile version based on the version number.
func (v *VersionMap) GetByVersion(version int) *Contract {
	contract, ok := v.versions[version]
	if !ok {
		panic("version not found")
	}

	return contract
}

// GetLatest returns the latest precompile version.
func (v *VersionMap) GetLatest() *Contract {
	return v.GetByVersion(len(v.versions))
}

// Address returns the address of the precompile. All versions share the same
// address so the latest version's address is returned.
func (v *VersionMap) Address() common.Address {
	return v.GetLatest().Address()
}