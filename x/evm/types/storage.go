// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"
)

// StorageRootStrategy defines the strategy for the EVM storage root mechanism.
// Our custom StateDB implementation does not support go-ethereum storage roots,
// yet it has to comply with the original go-ethereum StateDB interface.
// To overcome that problem, we initially used the DummyHash strategy, which
// returned a dummy hash in case a storage was detected under the given address.
// However, this approach turned out to be problematic for Mezo Passport
// hence, we switched to the EmptyHash strategy which returns an empty hash
// as storage root for every account, regardless of the actual storage.
// See https://github.com/mezo-org/mezod/issues/368 for more details.
// If we ever need to support storage roots, we should implement it as a
// new strategy.
type StorageRootStrategy uint32

const (
	// StorageRootStrategyDummyHash returns a dummy hash as the storage root if
	// the account has a storage and an empty hash otherwise.
	StorageRootStrategyDummyHash StorageRootStrategy = iota
	// StorageRootStrategyEmptyHash always returns an empty hash as the storage
	// root, regardless of the actual storage.
	StorageRootStrategyEmptyHash
)

// Storage represents the account Storage map as a slice of single key value
// State pairs. This is to prevent non determinism at genesis initialization or export.
type Storage []State

// Validate performs a basic validation of the Storage fields.
func (s Storage) Validate() error {
	seenStorage := make(map[string]bool)
	for i, state := range s {
		if seenStorage[state.Key] {
			return errorsmod.Wrapf(ErrInvalidState, "duplicate state key %d: %s", i, state.Key)
		}

		if err := state.Validate(); err != nil {
			return err
		}

		seenStorage[state.Key] = true
	}
	return nil
}

// String implements the stringer interface
func (s Storage) String() string {
	var str string
	for _, state := range s {
		str += fmt.Sprintf("%s\n", state.String())
	}

	return str
}

// Copy returns a copy of storage.
func (s Storage) Copy() Storage {
	cpy := make(Storage, len(s))
	copy(cpy, s)

	return cpy
}

// Validate performs a basic validation of the State fields.
// NOTE: state value can be empty
func (s State) Validate() error {
	if strings.TrimSpace(s.Key) == "" {
		return errorsmod.Wrap(ErrInvalidState, "state key hash cannot be blank")
	}

	return nil
}

// NewState creates a new State instance
func NewState(key, value common.Hash) State {
	return State{
		Key:   key.String(),
		Value: value.String(),
	}
}
