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
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/mezo-org/mezod/utils"
)

var (
	// DefaultEVMDenom defines the default EVM denomination on Mezo
	DefaultEVMDenom = utils.BaseDenom
	// DefaultAllowUnprotectedTxs rejects all unprotected txs (i.e false)
	DefaultAllowUnprotectedTxs = false
	// DefaultEnableCreate enables contract creation (i.e true)
	DefaultEnableCreate = true
	// DefaultEnableCall enables contract calls (i.e true)
	DefaultEnableCall = true
	// DefaultStorageRootStrategy defines the default strategy for the EVM storage root mechanism.
	DefaultStorageRootStrategy = StorageRootStrategyEmptyHash
	// DefaultMaxPrecompilesCallsPerExecution defines the default amount of precompiles
	// call per transaction execution allowed.
	DefaultMaxPrecompilesCallsPerExecution = 10
	// ChainFeeSplitterAddress defines the address of the fee splitter contract deployed
	// on Mezo. It will receive the fees from 'fee_collector' account collected from Mezo
	// transactions.
	DefaultChainFeeSplitterAddress = ""
	// DefaultMezoMinterAddress defines the address of the MEZO minter.
	DefaultMezoMinterAddress = ""
)

// AvailableExtraEIPs lists standard EIPs supported through ExtraEIPs. Values
// are validated with vm.ValidEip, which also permits registered custom EIPs.
// EIPs are applied in order and can override the latest hard fork instruction
// set. For more information, see:
// https://github.com/ethereum/go-ethereum/blob/master/core/vm/interpreter.go#L97
var AvailableExtraEIPs = []int64{1344, 1884, 2200, 2929, 3198, 3529}

// SelfdestructDisableEIP is the non-standard EIP (defined in
// mezo-org/go-ethereum) that disables the SELFDESTRUCT opcode. When present in
// ExtraEIPs, SELFDESTRUCT is treated as an invalid opcode.
const SelfdestructDisableEIP int64 = 90000

// MaxTxLockdownAddresses is the maximum length of a single transaction
// lockdown allowlist. The parameters are decoded for every transaction, so the
// cap bounds that cost.
const MaxTxLockdownAddresses = 256

// NewParams creates a new Params instance
func NewParams(evmDenom string, allowUnprotectedTxs, enableCreate, enableCall bool, config ChainConfig, extraEIPs []int64) Params {
	return Params{
		EvmDenom:            evmDenom,
		AllowUnprotectedTxs: allowUnprotectedTxs,
		EnableCreate:        enableCreate,
		EnableCall:          enableCall,
		ExtraEIPs:           extraEIPs,
		ChainConfig:         config,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	return Params{
		EvmDenom:                        DefaultEVMDenom,
		EnableCreate:                    DefaultEnableCreate,
		EnableCall:                      DefaultEnableCall,
		ChainConfig:                     DefaultChainConfig(),
		ExtraEIPs:                       []int64{SelfdestructDisableEIP},
		AllowUnprotectedTxs:             DefaultAllowUnprotectedTxs,
		StorageRootStrategy:             uint32(DefaultStorageRootStrategy),
		PrecompilesVersions:             DefaultPrecompilesVersions,
		MaxPrecompilesCallsPerExecution: uint32(DefaultMaxPrecompilesCallsPerExecution), //nolint:gosec
		ChainFeeSplitterAddress:         DefaultChainFeeSplitterAddress,
		MezoMinterAddress:               DefaultMezoMinterAddress,
	}
}

// Validate performs basic validation on evm parameters.
func (p Params) Validate() error {
	if err := validateEVMDenom(p.EvmDenom); err != nil {
		return err
	}

	if err := validateEIPs(p.ExtraEIPs); err != nil {
		return err
	}

	if err := validateBool(p.EnableCall); err != nil {
		return err
	}

	if err := validateBool(p.EnableCreate); err != nil {
		return err
	}

	if err := validateBool(p.AllowUnprotectedTxs); err != nil {
		return err
	}

	if err := validateTxLockdownAddresses(
		"tx lockdown sender",
		p.TxLockdownSenders,
	); err != nil {
		return err
	}

	if err := validateTxLockdownAddresses(
		"tx lockdown target",
		p.TxLockdownTargets,
	); err != nil {
		return err
	}

	return validateChainConfig(p.ChainConfig)
}

// EIPs returns the ExtraEIPS as a int slice
func (p Params) EIPs() []int {
	eips := make([]int, len(p.ExtraEIPs))
	for i, eip := range p.ExtraEIPs {
		eips[i] = int(eip)
	}
	return eips
}

func validateEVMDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter EVM denom type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateEIPs(i interface{}) error {
	eips, ok := i.([]int64)
	if !ok {
		return fmt.Errorf("invalid EIP slice type: %T", i)
	}

	for _, eip := range eips {
		if !vm.ValidEip(int(eip)) {
			return fmt.Errorf("EIP %d is not activateable, valid EIPS are: %s", eip, vm.ActivateableEips())
		}
	}

	return nil
}

// validateTxLockdownAddresses validates a single transaction lockdown
// allowlist. The name identifies the list in the error messages.
func validateTxLockdownAddresses(name string, addresses []string) error {
	if len(addresses) > MaxTxLockdownAddresses {
		return fmt.Errorf(
			"%s list holds %d entries, the maximum is %d",
			name,
			len(addresses),
			MaxTxLockdownAddresses,
		)
	}

	seen := make(map[string]struct{}, len(addresses))
	for i, address := range addresses {
		if len(address) == 0 {
			return fmt.Errorf("%s %d cannot be empty", name, i)
		}

		if !IsHexAddress(address) {
			return fmt.Errorf(
				"%s %d must be a valid hex-encoded EVM address",
				name,
				i,
			)
		}

		if IsZeroHexAddress(address) {
			return fmt.Errorf("%s %d cannot be the zero EVM address", name, i)
		}

		normalized := BytesToHexAddress(HexAddressToBytes(address))
		if _, ok := seen[normalized]; ok {
			return fmt.Errorf("%s %d is a duplicate: %s", name, i, address)
		}
		seen[normalized] = struct{}{}
	}

	return nil
}

func validateChainConfig(i interface{}) error {
	cfg, ok := i.(ChainConfig)
	if !ok {
		return fmt.Errorf("invalid chain config type: %T", i)
	}

	return cfg.Validate()
}

// IsLondon returns if london hardfork is enabled.
func IsLondon(ethConfig *params.ChainConfig, height int64) bool {
	return ethConfig.IsLondon(big.NewInt(height))
}
