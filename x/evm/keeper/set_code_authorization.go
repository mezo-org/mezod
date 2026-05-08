// Reimplementation of geth's unexported validateAuthorization /
// applyAuthorization on *statedb.StateDB. Re-implemented locally because
// rebuilding a geth stateTransition to reach the upstream methods would mean
// abandoning mezod's MinGasMultiplier / precompile / simulate paths. Re-audit
// on every geth bump; pins are by function name against
// mezo-org/go-ethereum@v1.16.9-mezo0
// (commit 859c41bdcb7141276fc9c7a7c3486190025f9ec1).
package keeper

import (
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/mezo-org/mezod/x/evm/statedb"
)

var (
	errSetCodeAuthorizationWrongChainID       = errors.New("set-code authorization chain id mismatch")
	errSetCodeAuthorizationNonceOverflow      = errors.New("set-code authorization nonce overflow")
	errSetCodeAuthorizationInvalidSignature   = errors.New("set-code authorization invalid signature")
	errSetCodeAuthorizationDestinationHasCode = errors.New("set-code authorization destination has non-delegation code")
	errSetCodeAuthorizationNonceMismatch      = errors.New("set-code authorization nonce mismatch")
	errSetCodeAuthorizationTargetIsPrecompile = errors.New("set-code authorization target is a precompile")
)

// validateSetCodeAuthorization validates an EIP-7702 authorization against the
// provided state. Mirrors geth's `validateAuthorization` in
// core/state_transition.go (mezo-org/go-ethereum@v1.16.9-mezo0,
// commit 859c41bdcb7141276fc9c7a7c3486190025f9ec1) with one Mezo-specific
// deviation: any authorization whose target is a precompile (stock Ethereum
// or a Mezo custom precompile) is rejected via isPrecompile. Mezo's custom
// precompiles register stored facade bytecode at their addresses, so a
// delegation there would actually execute in the authority's context;
// rejecting all precompile targets keeps the semantics surface narrow and
// matches stock geth's no-op behavior on its precompiles. Re-audit on every
// geth bump and preserve this check.
//
// Note: the authority is added to the access list even if a later check fails
// — this matches upstream's deliberate ordering (the AddAddressToAccessList
// call sits before the destination-code and nonce-mismatch checks).
func validateSetCodeAuthorization(
	stateDB *statedb.StateDB,
	chainID *big.Int,
	auth *ethtypes.SetCodeAuthorization,
	isPrecompile func(common.Address) bool,
) (common.Address, error) {
	var authority common.Address
	if !auth.ChainID.IsZero() && auth.ChainID.CmpBig(chainID) != 0 {
		return authority, errSetCodeAuthorizationWrongChainID
	}
	if auth.Nonce+1 < auth.Nonce {
		return authority, errSetCodeAuthorizationNonceOverflow
	}
	if isPrecompile(auth.Address) {
		return authority, errSetCodeAuthorizationTargetIsPrecompile
	}
	authority, err := auth.Authority()
	if err != nil {
		return authority, fmt.Errorf("%w: %v", errSetCodeAuthorizationInvalidSignature, err)
	}
	stateDB.AddAddressToAccessList(authority)
	code := stateDB.GetCode(authority)
	if _, ok := ethtypes.ParseDelegation(code); len(code) != 0 && !ok {
		return authority, errSetCodeAuthorizationDestinationHasCode
	}
	if have := stateDB.GetNonce(authority); have != auth.Nonce {
		return authority, errSetCodeAuthorizationNonceMismatch
	}
	return authority, nil
}

// applySetCodeAuthorization applies an EIP-7702 code delegation to the
// provided state. Mirrors geth's `applyAuthorization` in
// core/state_transition.go (mezo-org/go-ethereum@v1.16.9-mezo0,
// commit 859c41bdcb7141276fc9c7a7c3486190025f9ec1).
//
// On success the authority's nonce is bumped, its code is set to either the
// 0xef0100||target delegation marker or cleared (when auth.Address is zero),
// and the new-account intrinsic-gas overcharge is refunded if the authority
// already exists in state.
func applySetCodeAuthorization(
	stateDB *statedb.StateDB,
	chainID *big.Int,
	auth *ethtypes.SetCodeAuthorization,
	isPrecompile func(common.Address) bool,
) error {
	authority, err := validateSetCodeAuthorization(stateDB, chainID, auth, isPrecompile)
	if err != nil {
		return err
	}

	if stateDB.Exist(authority) {
		stateDB.AddRefund(params.CallNewAccountGas - params.TxAuthTupleGas)
	}

	stateDB.SetNonce(authority, auth.Nonce+1, tracing.NonceChangeAuthorization)
	if auth.Address == (common.Address{}) {
		stateDB.SetCode(authority, nil, tracing.CodeChangeAuthorizationClear)
		return nil
	}
	stateDB.SetCode(authority, ethtypes.AddressToDelegation(auth.Address), tracing.CodeChangeAuthorization)
	return nil
}

// applySetCodeAuthorizations applies the EIP-7702 authorization tuples carried
// by msg to stateDB and warms the resolved delegation target of msg.To.
//
// The msg.SetCodeAuthorizations loop mirrors the per-tuple application step
// upstream geth performs in core/state_transition.go (the loop following
// `applyAuthorization` in mezo-org/go-ethereum@v1.16.9-mezo0,
// commit 859c41bdcb7141276fc9c7a7c3486190025f9ec1). Errors from
// applySetCodeAuthorization are intentionally swallowed per EIP-7702 (invalid
// tuples must not abort the enclosing tx) but are emitted at debug level so
// operators can diagnose silently-rejected tuples.
//
// After the loop, the resolved delegation target of msg.To is added to the
// access list. This covers the case where a delegation pointing to msg.To was
// just installed by this same tx — without warming the target, the subsequent
// evm.Call would pay cold-access gas for the delegated address.
//
// Tolerates msg.SetCodeAuthorizations == nil. msg.To must be non-nil; callers
// only invoke this on the call path (contractCreation == false).
func applySetCodeAuthorizations(
	logger log.Logger,
	stateDB *statedb.StateDB,
	chainID *big.Int,
	msg core.Message,
	isPrecompile func(common.Address) bool,
) {
	if msg.SetCodeAuthorizations != nil {
		for i := range msg.SetCodeAuthorizations {
			if err := applySetCodeAuthorization(stateDB, chainID, &msg.SetCodeAuthorizations[i], isPrecompile); err != nil {
				logger.Debug(
					"set-code authorization rejected",
					"index", i,
					"error", err,
				)
			}
		}
	}
	// Convenience warming of msg.To's resolved delegation target — covers
	// the case where a delegation to msg.To was just installed by this tx.
	if addr, ok := ethtypes.ParseDelegation(stateDB.GetCode(*msg.To)); ok {
		stateDB.AddAddressToAccessList(addr)
	}
}
