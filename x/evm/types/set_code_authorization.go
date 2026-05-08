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
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// AuthorizationList is the EIP-7702 authorization list represented as a slice
// of the protobuf SetCodeAuthorization tuples.
type AuthorizationList []SetCodeAuthorization

// NewAuthorizationList creates a protobuf-compatible AuthorizationList from
// an ethereum-side []SetCodeAuthorization slice.
func NewAuthorizationList(auths []ethtypes.SetCodeAuthorization) AuthorizationList {
	if auths == nil {
		return nil
	}

	al := make(AuthorizationList, 0, len(auths))
	for _, a := range auths {
		al = append(al, NewSetCodeAuthorization(a))
	}
	return al
}

// ToEthAuthorizationList converts the protobuf AuthorizationList back into
// the geth-side []SetCodeAuthorization slice.
func (al AuthorizationList) ToEthAuthorizationList() []ethtypes.SetCodeAuthorization {
	if al == nil {
		return nil
	}

	out := make([]ethtypes.SetCodeAuthorization, len(al))
	for i, a := range al {
		out[i] = a.ToEthAuthorization()
	}
	return out
}

// NewSetCodeAuthorization converts a single ethereum-side authorization
// tuple into its protobuf form.
func NewSetCodeAuthorization(auth ethtypes.SetCodeAuthorization) SetCodeAuthorization {
	chainID := sdkmath.NewIntFromBigInt(auth.ChainID.ToBig())
	return SetCodeAuthorization{
		ChainID: &chainID,
		Address: auth.Address.Hex(),
		Nonce:   auth.Nonce,
		V:       []byte{auth.V},
		R:       auth.R.Bytes(),
		S:       auth.S.Bytes(),
	}
}

// ToEthAuthorization converts the protobuf authorization tuple back into the
// geth-side SetCodeAuthorization.
func (a SetCodeAuthorization) ToEthAuthorization() ethtypes.SetCodeAuthorization {
	var chainID uint256.Int
	if cid := a.GetChainID(); cid != nil {
		chainID = *uint256.MustFromBig(cid)
	}

	var v uint8
	if len(a.V) > 0 {
		v = a.V[0]
	}

	r := new(uint256.Int)
	if len(a.R) > 0 {
		r.SetBytes(a.R)
	}

	s := new(uint256.Int)
	if len(a.S) > 0 {
		s.SetBytes(a.S)
	}

	return ethtypes.SetCodeAuthorization{
		ChainID: chainID,
		Address: common.HexToAddress(a.Address),
		Nonce:   a.Nonce,
		V:       v,
		R:       *r,
		S:       *s,
	}
}

// GetChainID returns the authorization tuple chain id (nil-safe).
func (a SetCodeAuthorization) GetChainID() *big.Int {
	if a.ChainID == nil {
		return nil
	}
	return a.ChainID.BigInt()
}
