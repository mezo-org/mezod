package abci

import (
	"bytes"
	"cosmossdk.io/log"
	"fmt"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
	bridgeabci "github.com/mezo-org/mezod/x/bridge/abci"
)

// IProposalHandler is an interface representing a proposal handler.
type IProposalHandler interface {
	// PrepareProposalHandler returns the handler for the PrepareProposal ABCI request.
	PrepareProposalHandler() sdk.PrepareProposalHandler
	// ProcessProposalHandler returns the handler for the ProcessProposal ABCI request.
	ProcessProposalHandler() sdk.ProcessProposalHandler
}

// ProposalHandler is a handler for the PrepareProposal and
// ProcessProposal ABCI requests. It is designed to be used with
// multiple sub-handlers, each responsible for a specific part of the
// proposal. The handler itself is responsible for combining the
// results of the sub-handlers into a single proposal.
type ProposalHandler struct {
	logger      log.Logger
	subHandlers map[VoteExtensionPart]IProposalHandler
}

// NewProposalHandler creates a new ProposalHandler instance.
func NewProposalHandler(
	logger log.Logger,
	bridgeSubHandler *bridgeabci.ProposalHandler,
) *ProposalHandler {
	subHandlers := map[VoteExtensionPart]IProposalHandler{
		VoteExtensionPartBridge: bridgeSubHandler,
	}

	return &ProposalHandler{
		logger:      logger,
		subHandlers: subHandlers,
	}
}

// SetHandlers sets the PrepareProposal and ProcessProposal handlers on the
// provided base app.
func (ph *ProposalHandler) SetHandlers(baseApp *baseapp.BaseApp) {
	baseApp.SetPrepareProposal(ph.PrepareProposalHandler())
	baseApp.SetProcessProposal(ph.ProcessProposalHandler())
}

// PrepareProposalHandler returns the handler for the PrepareProposal ABCI request.
// It triggers the PrepareProposal sub-handlers that inject part-specific pseudo-txs
// and aggregates all the injections into a single app-level pseudo-tx that
// is put at the beginning of the proposal's transactions list. If any
// sub-handler fails, its error is logged and the part it was responsible for
// is not included in the final app-level pseudo-tx. This way, handlers can fail
// independently of each other. A sub-handler can decide to not inject the
// part-specific pseudo-tx and such an empty part is still included in the final
// app-level pseudo-tx. This way, presence of the given part in the app-level
// pseudo-tx indicates the success/failure of the corresponding sub-handler and
// may be used for debugging purposes.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. Cosmos SDK will log the error and return the
// original unmodified proposal's transactions list to the CometBFT engine.
//
// Invariants summary:
//  1. Sub-handler's response is considered valid if no error is returned,
//     regardless of whether the pseudo-tx was injected into the response
//     or not.
//  2. Returns an error if all sub-handlers return an invalid response.
//  3. Injects a non-empty app-level pseudo-tx if at least one sub-handler
//     returns a valid response.
//  4. Returns an error if the app-level pseudo-tx cannot be marshaled.
//  5. Returns the original proposal's transactions list if vote extensions
//     are not enabled.
//  6. Guarantees that the final transactions list size does not exceed the
//     maximum allowed block size.
func (ph *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestPrepareProposal,
	) (*cmtabci.ResponsePrepareProposal, error) {
		if !isVoteExtensionsEnabled(ctx, req.Height) {
			// Short-circuit if vote extensions are not enabled.
			return &cmtabci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		injectedTxParts := make(map[uint32][]byte)

		for part, subHandler := range ph.subHandlers {
			// Trigger the PrepareProposal sub-handler for the given part.
			//
			// The sub-handler is responsible for:
			// - Validating the signatures of the commit's vote extensions (using baseapp.ValidateVoteExtensions)
			// - Extracting its respective parts from the commit's vote extensions
			// - Validating whether the extracted parts are valid according to the sub-handler's rules
			// - Making sure valid parts are backed by the super-majority of the validators
			res, err := subHandler.PrepareProposalHandler()(ctx, req)
			if err != nil {
				// Just log the error and continue execution in case there
				// are other sub-handlers in the queue. We do not want
				// make sub-handlers dependent on each other.
				ph.logger.Error(
					"sub-handler failed to prepare proposal",
					"height", req.Height,
					"part", part,
					"err", err,
				)
				continue
			}

			// Note that len(extractInjectedTx(req, res)) may be 0 and this
			// is a valid case as the handler may decide to not inject anything.
			// Such an empty part is still included in the app-level
			// injected pseudo-transaction. Missing part indicates an error
			// of the given sub-handler.
			injectedTxParts[uint32(part)] = extractInjectedTx(req, res)
		}

		if len(injectedTxParts) == 0 {
			// Short-circuit if all sub-handlers failed.
			return nil, fmt.Errorf("all sub-handlers failed to prepare proposal")
		}

		// Construct the final injected pseudo-transaction from the parts.
		injectedTx := types.InjectedTx{Parts: injectedTxParts}
		// Marshal the injected pseudo-transaction into bytes.
		injectedTxBytes, err := injectedTx.Marshal()
		if err != nil {
			// If marshaling fails, we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to marshal injected tx: %w", err)
		}

		// Inject the pseudo-transaction at the beginning of the original
		// transaction list being part of the proposal.
		draftTxs := append([][]byte{injectedTxBytes}, req.Txs...)
		// We need to ensure the final transactions list size does not exceed
		// the maximum allowed block size.
		txs := make([][]byte, 0)
		txsBytes := int64(0)
		for _, tx := range draftTxs {
			txsBytes += int64(len(tx))
			if txsBytes > req.MaxTxBytes {
				break
			}
			txs = append(txs, tx)
		}

		return &cmtabci.ResponsePrepareProposal{Txs: txs}, nil
	}
}

// ProcessProposalHandler returns the handler for the ProcessProposal ABCI request.
// It decomposes the app-level pseudo-tx injected during the PrepareProposal phase
// into part-specific pseudo-txs and triggers the corresponding ProcessProposal
// sub-handler for each part.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. Cosmos SDK will log the error and REJECT
// the app-level proposal. Rejecting the proposal is a serious event and has
// liveness implications for the CometBFT engine. Make sure you know what you
// are doing before returning an error from this function.
//
// Invariants summary:
//  1. Accepts the app-level proposal if vote extensions are not enabled.
//  1. Accepts app-level proposal with an empty transactions vector.
//  2. Rejects app-level proposal with an injected pseudo-tx that cannot be unmarshaled.
//  4. Rejects app-level proposal with an injected pseudo-tx containing no parts.
//  5. Accepts app-level proposal only if each part of their injected pseudo-tx
//     corresponds to a known sub-handler.
//  6. Accepts app-level proposal only if each part of their injected pseudo-tx
//     is accepted by the corresponding sub-handler.
//
// TODO: Currently, this function either ACCEPT the proposal or return
// an error - it does not REJECT anything explicitly. This is fine as
// Cosmos SDK handles the error gracefully by logging it on the error
// level and rejecting the proposal. This is fine for now, but we
// may consider a more granular approach and start distinguish between
// a case when the validation completes successfully but the proposal
// is invalid and a case when the validation fails (sub-handlers can
// distinguish between these cases by returning either REJECT + error
// holding the reason or nil + error, respectively). Having this distinction
// could allow us to log rejections on warn level and leave log errors
// for actual errors. The main motivation here is validator experience.
// Error logs should ideally lead to action items for the given validator
// while rejection warnings should stay informative and highlight potential
// misbehavior of other validators.
func (ph *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestProcessProposal,
	) (*cmtabci.ResponseProcessProposal, error) {
		if !isVoteExtensionsEnabled(ctx, req.Height) {
			// Short-circuit if vote extensions are not enabled.
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			}, nil
		}

		if len(req.Txs) == 0 {
			// Short-circuit if proposal has no transactions. This case is
			// possible when the injected transaction was not included in the
			// proposal due to the maximum allowed block size constraint.
			return &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			}, nil
		}

		var injectedTx types.InjectedTx
		if err := injectedTx.Unmarshal(req.Txs[0]); err != nil {
			// If unmarshaling fails, we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to unmarshal injected tx: %w", err)
		}

		if len(injectedTx.Parts) == 0 {
			// Reject pseudo-transactions with no parts. This is done in order
			// to ensure parity with the PrepareProposalHandler, which always
			// produces a pseudo-transaction with at least one part.
			return nil, fmt.Errorf("injected tx has no parts")
		}

		for partUint, partBytes := range injectedTx.Parts {
			part := VoteExtensionPart(partUint)

			subHandler, ok := ph.subHandlers[part]
			if !ok {
				// Make sure the pseudo-transaction part is recognized. If not,
				// reject the proposal as something is clearly wrong.
				return nil, fmt.Errorf("unknown injected tx part: %d", part)
			}

			// The transactions vector passed to the sub-handler must be
			// modified to contain the sub-handler's pseudo-transaction part
			// at the beginning. We cannot pass the whole app-level
			// pseudo-transaction which is originally at the beginning of the
			// transactions vector as the sub-handler won't be able to
			// unmarshal it. Note that it's safe to do req.Txs[1:] as the
			// req.Txs slice is guaranteed to have at least one element
			// at this point (see the `len(injectedTx.Parts) == 0` check above).
			subTxs := append([][]byte{partBytes}, req.Txs[1:]...)

			// Trigger the ProcessProposal sub-handler for the given part.
			//
			// The sub-handler is responsible for re-executing the same
			// logic as in the PrepareProposal phase in order to make sure the
			// proposed pseudo-transaction part is actually valid and the
			// block proposer behaves correctly. Please note that an important
			// part of this logic is re-validating the signatures of the
			// vote extensions used to build the pseudo-transaction part
			// (using baseapp.ValidateVoteExtension). The block proposer should
			// always include those vote extensions in the pseudo-transaction
			// part to enable the aforementioned signature validation.
			res, err := subHandler.ProcessProposalHandler()(
				ctx,
				&cmtabci.RequestProcessProposal{
					Txs:                subTxs,
					ProposedLastCommit: req.ProposedLastCommit,
					Misbehavior:        req.Misbehavior,
					Hash:               req.Hash,
					Height:             req.Height,
					Time:               req.Time,
					NextValidatorsHash: req.NextValidatorsHash,
					ProposerAddress:    req.ProposerAddress,
				},
			)
			if err != nil {
				// If a sub-handler fails to process its pseudo-transaction part,
				// reject the whole proposal.
				return nil, fmt.Errorf(
					"sub-handler failed to process injexted tx part %v: %w",
					part,
					err,
				)
			}
			if res.Status != cmtabci.ResponseProcessProposal_ACCEPT {
				// If a sub-handler rejects its pseudo-transaction part,
				// reject the whole proposal.
				return nil, fmt.Errorf(
					"sub-handler rejected injexted tx part %v",
					part,
				)
			}
		}

		return &cmtabci.ResponseProcessProposal{
			Status: cmtabci.ResponseProcessProposal_ACCEPT,
		}, nil
	}
}

// extractInjectedTx returns the injected transaction from the PrepareProposal
// response. An injected transaction is a transaction that was added to the
// response by a sub-handler during the PrepareProposal phase. By convention,
// the first transaction in the response is considered the injected transaction.
// The function returns nil if no injected transaction was found.
func extractInjectedTx(
	req *cmtabci.RequestPrepareProposal,
	res *cmtabci.ResponsePrepareProposal,
) []byte {
	// Helper to safely extract the first transaction from a slice of
	// transactions without panicking if the slice is empty.
	safeFirstTx := func(txs [][]byte) []byte {
		if len(txs) == 0 {
			return nil
		}

		return txs[0]
	}

	firstReqTx := safeFirstTx(req.Txs)
	firstResTx := safeFirstTx(res.Txs)

	if len(firstResTx) == 0 || bytes.Equal(firstResTx, firstReqTx) {
		// If the first transaction in the response doesn't exist or is the
		// same as the first transaction in the request, nothing has
		// been injected for sure.
		return nil
	}

	return firstResTx
}
