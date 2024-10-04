package abci

import (
	"fmt"

	"cosmossdk.io/log"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	"github.com/mezo-org/mezod/x/bridge/keeper"
)

// PreBlockHandler is the bridge-specific pre-block handler (part of the
// FinalizeBlock ABCI request)
type PreBlockHandler struct {
	logger       log.Logger
	bridgeKeeper keeper.Keeper
}

// NewPreBlockHandler returns a new PreBlockHandler.
func NewPreBlockHandler(
	logger log.Logger,
	bridgeKeeper keeper.Keeper,
) *PreBlockHandler {
	return &PreBlockHandler{
		logger:       logger,
		bridgeKeeper: bridgeKeeper,
	}
}

// PreBlocker returns the pre-block handler (part of the FinalizeBlock ABCI request).
// This function:
//   - Extracts the sequence of canonical AssetsLocked events from the injected
//     pseudo-transaction. Minimum validation is performed to avoid unexpected
//     panics but no error is expected here as this handler is invoked after
//     the proposal phase which guarantees the pseudo-transaction is well-formed
//     and contain a valid sequence of canonical AssetsLocked events.
//   - Accepts the AssetsLocked events by calling the bridge keeper's AcceptAssetsLocked
//     method. This call updates the internal state of the bridge module to reflect
//     the new AssetsLocked events and mints new BTC as result.
//
// Dev note: In case of success, this function should return a non-nil pointer value
// (&sdk.ResponsePreBlock{}) as ResponsePreBlock and a nil error. The specific
// value of ResponsePreBlock is not relevant for the app-level handler but, we
// try to adhere to the Go best practices and return a non-nil pointer value
// in case of success. Conversely, in case of failure, this function should return
// a nil ResponsePreBlock and a non-nil error. BEWARE: An error returned from this
// function will be bubbled up by the app-level handler and will cause the FinalizeBlock
// ABCI request to fail. This, in turn, will lead to an unrecoverable consensus failure.
// Make sure you know what you are doing before returning an error.
func (pbh *PreBlockHandler) PreBlocker() sdk.PreBlocker {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestFinalizeBlock,
	) (*sdk.ResponsePreBlock, error) {
		// TODO: Consider changing logging to debug level once this code matures.

		pbh.logger.Info(
			"bridge is executing pre-block",
			"height", req.Height,
		)

		if len(req.Txs) == 0 {
			// The app-level handler always passes a transaction vector
			// with at least one element (representing the possibly empty
			// bridge-specific pseudo-transaction). If the vector is empty,
			// something went wrong upstream and we cannot recover so, return
			// an error.
			return nil, fmt.Errorf("empty transaction vector in the block")
		}

		if len(req.Txs[0]) == 0 {
			// As said above, the app-level handler always passes a transaction
			// vector with at least one element (representing the possibly empty
			// bridge-specific pseudo-transaction). The first element can be an
			// empty byte slice if the bridge-level PrepareProposal handler
			// decided to not inject the pseudo-transaction. In this case, we
			// do not need to do anything, and we can short-circuit the processing.
			pbh.logger.Info(
				"bridge skipped pre-block processing; "+
					"no AssetsLocked events sequence in the block",
				"height", req.Height,
			)

			return &sdk.ResponsePreBlock{}, nil
		}

		var injectedTx types.InjectedTx
		if err := injectedTx.Unmarshal(req.Txs[0]); err != nil {
			// If unmarshaling of the injected pseudo-transaction fails,
			// we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to unmarshal injected tx: %w", err)
		}

		events := injectedTx.AssetsLockedEvents

		eventsSequenceNumbers := make([]string, 0)
		for _, event := range events {
			if !event.Sequence.IsNil() {
				eventsSequenceNumbers = append(eventsSequenceNumbers, event.Sequence.String())
			}
		}

		pbh.logger.Info(
			"AssetsLocked events sequence extracted from block",
			"height", req.Height,
			"events_sequence_numbers", eventsSequenceNumbers,
		)

		// We do not validate the `events` slice as we assume all requirements
		// of AcceptAssetsLocked were ensured during the proposal phase.
		err := pbh.bridgeKeeper.AcceptAssetsLocked(ctx, events)
		if err != nil {
			return nil, fmt.Errorf("cannot accept AssetsLocked events: %w", err)
		}

		pbh.logger.Info(
			"bridge executed pre-block",
			"height", req.Height,
			"sequence_tip", pbh.bridgeKeeper.GetAssetsLockedSequenceTip(ctx),
		)

		return &sdk.ResponsePreBlock{}, nil
	}
}
