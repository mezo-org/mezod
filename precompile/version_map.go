package precompile

import (
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"
)

// VersionMap is a map of precompile versions.
type VersionMap struct {
	versions map[int]*Contract
}

// NewVersionMap creates a new version map for the precompile.
func NewVersionMap(versions map[int]*Contract) *VersionMap {
	if len(versions) == 0 {
		panic("no contracts provided")
	}

	keys := maps.Keys(versions)
	slices.Sort(keys)

	if !isStrictlyIncreasingVersions(keys) {
		panic("versions must be strictly increasing by 1")
	}

	address := versions[keys[0]].Address()

	for _, contract := range versions {
		if contract == nil {
			panic("nil contract provided")
		}

		if address != contract.Address() {
			panic("all contracts must have the same address")
		}
	}

	return &VersionMap{
		versions: versions,
	}
}

// isStrictlyIncreasingVersions returns true if the versions slice form a
// sequence strictly increasing by 1. Such a sequence guarantees that there are
// no gaps between the sequence numbers of the events and that each event is
// unique by its sequence number. Returns true for empty and single-element slices.
func isStrictlyIncreasingVersions(versions []int) bool {
	for i := 0; i < len(versions)-1; i++ {
		expectedNextVersion := versions[i] + 1
		if expectedNextVersion != versions[i+1] {
			return false
		}
	}

	return true
}

// GetByVersion returns the precompile version based on the version number.
// The boolean return value indicates if the version exists.
func (v *VersionMap) GetByVersion(version int) (*Contract, bool) {
	contract, ok := v.versions[version]
	return contract, ok
}

// GetLatestVersion returns the latest precompile version number.
func (v *VersionMap) GetLatestVersion() int {
	keys := maps.Keys(v.versions)
	slices.Sort(keys)
	return keys[len(keys)-1]
}

// GetLatest returns the latest precompile version.
func (v *VersionMap) GetLatest() *Contract {
	latest, ok := v.GetByVersion(v.GetLatestVersion())
	if !ok {
		panic("invalid precompile version map")
	}

	return latest
}

// Address returns the address of the precompile. All versions share the same
// address so the latest version's address is returned.
func (v *VersionMap) Address() common.Address {
	return v.GetLatest().Address()
}
