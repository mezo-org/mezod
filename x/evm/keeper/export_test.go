package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// ApplySetCodeAuthorization re-exports applySetCodeAuthorization so the
// keeper_test package can drive EIP-7702 authorization processing in
// isolation from applyMessageWithConfig.
var ApplySetCodeAuthorization = applySetCodeAuthorization

// EIP-7702 validation sentinel errors re-exported so keeper_test can
// match the exact failure mode via errors.Is, instead of asserting only
// that some error was returned. Exposing these by identity prevents a
// regression that collapses two validation branches into the same error
// (or swaps their source order) from going undetected.
var (
	ErrSetCodeAuthorizationWrongChainID       = errSetCodeAuthorizationWrongChainID
	ErrSetCodeAuthorizationNonceOverflow      = errSetCodeAuthorizationNonceOverflow
	ErrSetCodeAuthorizationInvalidSignature   = errSetCodeAuthorizationInvalidSignature
	ErrSetCodeAuthorizationDestinationHasCode = errSetCodeAuthorizationDestinationHasCode
	ErrSetCodeAuthorizationNonceMismatch      = errSetCodeAuthorizationNonceMismatch
)

// ApplyMessageWithStateDB executes a message via applyMessageWithConfig
// using the caller-supplied *statedb.StateDB so tests can inspect
// transient state (notably the access list) after the call returns. The
// public ApplyMessage / ApplyMessageWithConfig entry points construct
// their own StateDB internally and discard it; that is incompatible with
// post-call access-list assertions because access lists do not survive
// Commit().
func (k *Keeper) ApplyMessageWithStateDB(
	ctx sdk.Context,
	wrapper MessageWrapper,
	tracer *tracers.Tracer,
	commit bool,
	cfg *statedb.EVMConfig,
	txConfig statedb.TxConfig,
	stateDB *statedb.StateDB,
) (*types.MsgEthereumTxResponse, []statedb.StateChange, error) {
	return k.applyMessageWithConfig(ctx, wrapper, tracer, commit, cfg, txConfig, stateDB, nil)
}
