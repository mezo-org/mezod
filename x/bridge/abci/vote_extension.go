package abci

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	"github.com/mezo-org/mezod/x/bridge/keeper"
)

// AssetsLockedEventsLimit is the maximum number of AssetsLocked events to
// fetch in a single request and include in the vote extension.
const AssetsLockedEventsLimit = 10

// VoteExtensionHandler is the bridge-specific handler for the ExtendVote and
// VerifyVoteExtension ABCI requests.
type VoteExtensionHandler struct {
	logger        log.Logger
	sidecarClient EthereumSidecarClient
	bridgeKeeper  keeper.Keeper
}

// NewVoteExtensionHandler creates a new VoteExtensionHandler instance.
func NewVoteExtensionHandler(
	logger log.Logger,
	sidecarClient EthereumSidecarClient,
	bridgeKeeper keeper.Keeper,
) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		logger:        logger,
		sidecarClient: sidecarClient,
		bridgeKeeper:  bridgeKeeper,
	}
}

// ExtendVoteHandler returns the handler for the ExtendVote ABCI request.
// It fetches the AssetsLocked events from the sidecar and includes them in the
// vote extension, in their natural order (by sequence asc).
// Events are fetched from a half-open range [start, end), where `start` is the
// currently stored sequence tip + 1, and `end` is `start` + AssetsLockedEventsLimit.
// It is guaranteed that the number of events included in the vote extension
// will not exceed AssetsLockedEventsLimit.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. The upstream app-level vote extension handler
// will handle the error gracefully and won't include the bridge-specific part
// in the app-level vote extension.
func (veh *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestExtendVote,
	) (*cmtabci.ResponseExtendVote, error) {
		// TODO: Consider changing logging to debug level once this code matures.

		veh.logger.Info(
			"bridge is extending vote",
			"height", req.Height,
		)

		// Try to extract the sequence tip from AssetsLocked events that are
		// included in this block's proposal. Those events are very likely
		// to be processed upon block finalization, and they will move
		// ahead the sequence tip in the bridge state. By determining the sequence
		// tip based on the proposal's events, we can avoid fetching the same
		// events from the sidecar twice and burning vote extension cycles on them.
		var sequenceTip math.Int
		if len(req.Txs) > 0 && len(req.Txs[0]) > 0 {
			var injectedTx types.InjectedTx
			if err := injectedTx.Unmarshal(req.Txs[0]); err != nil {
				// If the transaction vector and the first tx are not empty, the
				// first transaction must be the injected bridge-specific
				// pseudo-transaction and unmarshaling must succeed.
				// If it fails, we cannot recover, so return an error.
				return nil, fmt.Errorf("failed to unmarshal injected tx: %w", err)
			}

			events := injectedTx.AssetsLockedEvents
			if len(events) == 0 {
				// This should not happen because the proposal phase guarantees
				// the presence of AssetsLocked events in the injected
				// bridge-specific pseudo-transaction.
				return nil, fmt.Errorf("no AssetsLocked events in the injected tx")
			}

			sequenceTip = events[len(events)-1].Sequence
		}

		// If the sequence tip is not determined from the proposal, fetch it from
		// the bridge state.
		if sequenceTip.IsNil() {
			sequenceTip = veh.bridgeKeeper.GetAssetsLockedSequenceTip(ctx)
		}

		veh.logger.Info(
			"assets locked sequence tip fetched",
			"height", req.Height,
			"sequence_tip", sequenceTip,
		)

		sequenceStart := sequenceTip.Add(math.NewInt(1))
		sequenceEnd := sequenceStart.Add(math.NewInt(AssetsLockedEventsLimit))

		veh.logger.Info(
			"fetching assets locked events from the sidecar",
			"height", req.Height,
			"sequence_start", sequenceStart,
			"sequence_end", sequenceEnd,
		)

		events, err := veh.sidecarClient.GetAssetsLockedEvents(
			ctx,
			&sequenceStart,
			&sequenceEnd,
		)
		if err != nil {
			// If fetching events fails, we cannot recover, so return an error.
			return nil, fmt.Errorf(
				"failed to fetch AssetsLocked events from the sidecar: %w",
				err,
			)
		}

		veh.logger.Info(
			"sidecar returned assets locked events",
			"height", req.Height,
			"events_count", len(events),
		)

		// NOTE: Despite the sidecar client should abide the contract defined in
		// the EthereumSidecarClient interface, we are doing validation of the
		// returned data to maintain parity with the VerifyVoteExtension logic.
		// The ExtendVote handler is used by honest validators so, it must
		// guarantee that the produced vote extension is accepted by the
		// VerifyVoteExtension handler.

		if err := validateAssetsLockedEvents(events); err != nil {
			return nil, err
		}

		voteExtension := types.VoteExtension{
			AssetsLockedEvents: events,
		}
		// Marshal the vote extension into bytes. Note that if len(events) == 0,
		// the Marshal method will return an empty byte slice so an empty
		// vote extension part will be returned from this handler.
		voteExtensionBytes, err := voteExtension.Marshal()
		if err != nil {
			// If marshaling fails, we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		veh.logger.Info(
			"bridge extended vote",
			"height", req.Height,
		)

		return &cmtabci.ResponseExtendVote{VoteExtension: voteExtensionBytes}, nil
	}
}

// VerifyVoteExtensionHandler returns the handler for the VerifyVoteExtension
// ABCI request. It verifies the vote extension by checking that:
//   - The vote extension unmarshals
//   - AssetsLocked events are valid (positive sequence number, positive amount,
//     proper bech32 recipient) and form a sequence strictly increasing by 1
//   - The number of AssetsLocked events does not exceed the limit
//
// If the vote extension is valid, it is accepted. Empty vote extensions are
// accepted by default.
//
// Dev note: In case the vote extension is invalid, we REJECT it explicitly
// and return an error describing the reason. Due to the limitations of the
// Cosmos interface, REJECT without an error does not provide any details about
// the reason. Conversely, error without REJECT is confusing as it should rather
// denote a failure of the handler itself. The upstream app-level vote extension
// handler will handle all non-ACCEPT cases gracefully and reject the app-level
// vote extension.
//
// See Skip's price oracle VerifyVoteExtension handler for a similar pattern:
// https://github.com/skip-mev/connect/blob/8c9ac8bf5b5bf239caa11086db34f88f30efe2c5/abci/ve/vote_extension.go#L213
func (veh *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(
		_ sdk.Context,
		req *cmtabci.RequestVerifyVoteExtension,
	) (*cmtabci.ResponseVerifyVoteExtension, error) {
		from := sdk.ConsAddress(req.ValidatorAddress).String()

		veh.logger.Debug(
			"bridge is verifying vote extension",
			"height", req.Height,
			"from", from,
		)

		if len(req.VoteExtension) == 0 {
			// Accept empty bridge-specific vote extensions. This is necessary
			// given that this handler's ExtendVote produces empty ones when
			// the Ethereum sidecar returns no events.
			veh.logger.Debug(
				"bridge accepted empty vote extension",
				"height", req.Height,
				"from", from,
			)

			return &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			}, nil
		}

		// Unmarshal the vote extension. Note that at this point, the
		// len(voteExtension.AssetsLockedEvents) > 0 because we short-circuit
		// len(req.VoteExtension) == 0 above. In practice, there is no
		// possibility that len(req.VoteExtension) > 0 produces
		// len(voteExtension.AssetsLockedEvents) == 0 with the current
		// protobuf implementation. This may change in the future but
		// even then, the case of len(voteExtension.AssetsLockedEvents) == 0
		// will not harm the downstream logic of this function and such a vote
		// extension should be accepted properly.
		var voteExtension types.VoteExtension
		if err := voteExtension.Unmarshal(req.VoteExtension); err != nil {
			// If the vote extension cannot be unmarshalled, we cannot recover.
			return &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			}, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		if err := validateAssetsLockedEvents(voteExtension.AssetsLockedEvents); err != nil {
			return &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			}, err
		}

		veh.logger.Debug(
			"bridge accepted vote extension",
			"height", req.Height,
			"from", from,
		)

		return &cmtabci.ResponseVerifyVoteExtension{
			Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
		}, nil
	}
}

// validateAssetsLockedEvents validates the given list of AssetsLocked events
// in the context of the bridge vote extension.
//
// The given list is considered valid if:
//   - The number of events does not exceed the AssetsLockedEventsLimit
//   - All events in the slice are valid (positive sequence number, positive
//     amount, proper bech32 recipient) and form a sequence strictly increasing
//     by 1
//
// If the validation passes, the function returns nil. Otherwise, it returns
// an error describing the reason.
func validateAssetsLockedEvents(events []bridgetypes.AssetsLockedEvent) error {
	if len(events) > AssetsLockedEventsLimit {
		return fmt.Errorf("number of events exceeds the limit")
	}

	if !bridgetypes.AssetsLockedEvents(events).IsValid() {
		return fmt.Errorf("events list is not valid")
	}

	return nil
}
