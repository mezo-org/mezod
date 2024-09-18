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

func (ph *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestProcessProposal,
	) (*cmtabci.ResponseProcessProposal, error) {
		// TODO: Implement the ProcessProposalHandler.
		return nil, nil
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
