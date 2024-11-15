package abci

import (
	"fmt"
	"slices"

	"cosmossdk.io/log"

	"golang.org/x/exp/maps"

	bridgeabci "github.com/mezo-org/mezod/x/bridge/abci"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
)

// IModuleManager is an interface representing the module manager.
type IModuleManager interface {
	// PreBlock returns the pre-block logic for the module manager.
	PreBlock(ctx sdk.Context) (*sdk.ResponsePreBlock, error)
}

// IPreBlockHandler is an interface representing the pre-block handler
// (part of the FinalizeBlock ABCI request).
type IPreBlockHandler interface {
	// PreBlocker returns the pre-block handler (part of the FinalizeBlock ABCI request).
	PreBlocker() sdk.PreBlocker
}

// PreBlockHandler is a pre-block handler (part of the FinalizeBlock ABCI request).
// It is designed to be used with multiple sub-handlers, each responsible for a
// specific part of the pre-block logic.
type PreBlockHandler struct {
	logger      log.Logger
	subHandlers map[VoteExtensionPart]IPreBlockHandler
}

// NewPreBlockHandler returns a new PreBlockHandler.
func NewPreBlockHandler(
	logger log.Logger,
	bridgeSubHandler *bridgeabci.PreBlockHandler,
) *PreBlockHandler {
	subHandlers := map[VoteExtensionPart]IPreBlockHandler{
		VoteExtensionPartBridge: bridgeSubHandler,
	}

	return &PreBlockHandler{
		logger:      logger,
		subHandlers: subHandlers,
	}
}

// PreBlocker returns the pre-block handler (part of the FinalizeBlock ABCI request).
// This function starts by invoking the module manager's PreBlock method to
// invoke module-level pre-block logic. Next, it iterates over app-level
// pre-block sub-handlers, invoking each sub-handler's PreBlocker method.
// The app-level pre-block sub-handlers (unlike the module-level pre-block handlers)
// have access to the corresponding pseudo-transactions injected during the
// proposal phase and can perform state changes based on these transactions.
//
// Dev note: Returning an error from this function will make the FinalizeBlock
// ABCI request fail and lead to an unrecoverable consensus failure.
// Make sure you know what you are doing before returning an error.
//
// Invariants summary:
//  1. Calls only the module manager's PreBlock if vote extensions are not enabled.
//  2. Returns error if the module manager's PreBlock call fails.
//  3. Returns error if any of the sub-handlers fails.
func (pbh *PreBlockHandler) PreBlocker(mm IModuleManager) sdk.PreBlocker {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestFinalizeBlock,
	) (*sdk.ResponsePreBlock, error) {
		// TODO: Consider changing logging to debug level once this code matures.

		res, err := mm.PreBlock(ctx)
		if err != nil {
			return nil, err
		}

		if !isVoteExtensionsEnabled(ctx, req.Height) {
			// Short-circuit if vote extensions are not enabled.
			return res, nil
		}

		// If the transaction vector for this block is not empty, the
		// first transaction MAY be the app-level pseudo-transaction injected
		// during PrepareProposal. If the first transaction does not unmarshal
		// as an injected pseudo-transaction, that means the PrepareProposal
		// handler did not inject it due to all proposal sub-handlers failing.
		// In this case, the first transaction is a regular application-specific
		// transaction. Moreover, an empty transaction vector is also a valid
		// case and indicates that no pseudo-transaction was injected and there
		// are no regular application-specific txs in the vector.
		var injectedTx types.InjectedTx
		var injectedTxOk bool
		var regularTxs [][]byte
		if len(req.Txs) > 0 {
			if err := injectedTx.Unmarshal(req.Txs[0]); err != nil {
				// If the first transaction does not unmarshal as an injected
				// pseudo-transaction, that means all transactions in the
				// vector are regular chain transactions.
				injectedTxOk = false
				regularTxs = req.Txs
			} else {
				// If the first transaction unmarshals as an injected pseudo-transaction,
				// regular transactions occur after the injected pseudo-transaction.
				injectedTxOk = true
				regularTxs = req.Txs[1:]
			}
		}

		// Unlike in app-level VoteExtensionHandler and ProposalHandler, we
		// cannot think about running the pre-block sub-handlers in parallel.
		// This is because pre-block sub-handlers may change the state and
		// state changes have to occur sequentially. To avoid race, we guarantee
		// the sub-handlers are executed in a deterministic order.
		subHandlerKeys := maps.Keys(pbh.subHandlers)
		slices.Sort(subHandlerKeys)

		for _, part := range subHandlerKeys {
			pbh.logger.Info(
				"running sub-handler to execute pre-block",
				"height", req.Height,
				"part", part,
			)

			// No need to check if the sub-handler exists. We know it exists
			// because we iterate over the keys of the sub-handler map.
			subHandler := pbh.subHandlers[part]

			// Replace the app-level pseudo-transaction with the part-specific
			// pseudo-transaction. Note that the part-specific pseudo-transaction
			// may be zero-length, which is a valid case. The sub-handler is
			// responsible for handling this case.
			var injectedTxPart []byte
			if injectedTxOk && injectedTx.Parts != nil {
				injectedTxPart = injectedTx.Parts[uint32(part)]
			}
			subTxs := append([][]byte{injectedTxPart}, regularTxs...)

			// We ignore the exact response from the sub-handler. Construction
			// of the sdk.ResponsePreBlock implies that it should come from
			// the module manager. We do not want to interfere with that
			// by including responses from sub-handlers. The only goal here is
			// that sub-handlers introduce their state changes, based on
			// the txs they have access to.
			_, err = subHandler.PreBlocker()(ctx, &cmtabci.RequestFinalizeBlock{
				Txs:                subTxs,
				DecidedLastCommit:  req.DecidedLastCommit,
				Misbehavior:        req.Misbehavior,
				Hash:               req.Hash,
				Height:             req.Height,
				Time:               req.Time,
				NextValidatorsHash: req.NextValidatorsHash,
				ProposerAddress:    req.ProposerAddress,
			})
			if err != nil {
				// The vote extension and proposal phases should guarantee the
				// pre-block stage does not encounter any errors. If it does,
				// something went really wrong unexpectedly. The error should
				// bubble up and cause an unrecoverable consensus failure.
				return nil, fmt.Errorf(
					"pre-blocker for part %v failed: %w",
					part,
					err,
				)
			}

			pbh.logger.Info(
				"sub-handler executed pre-block",
				"height", req.Height,
				"part", part,
			)
		}

		pbh.logger.Info(
			"pre-block executed",
			"height", req.Height,
		)

		return res, nil
	}
}
