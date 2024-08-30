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
		// TODO: Fetched events will be finalized in the next block and the
		//       tip will be updated then. Because of that, we may fetch the same
		//       set of events twice, in two subsequent blocks. Once the full
		//       flow is implemented, we will need to check this behavior and
		//       fix it if necessary.
		//
		// TODO: Consider changing logging to debug level once this code matures.

		veh.logger.Info(
			"bridge is extending vote",
			"height", req.Height,
		)

		sequenceTip := veh.bridgeKeeper.GetAssetsLockedSequenceTip(ctx)

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

		if len(events) > AssetsLockedEventsLimit {
			// Make sure the number of events does not exceed the limit.
			return nil, fmt.Errorf("number of events exceeds the limit")
		}

		if !bridgetypes.AssetsLockedEvents(events).IsValid() {
			// Make sure all events in the slice are valid (positive sequence
			// number, positive amount, proper bech32 recipient) and form a
			// sequence strictly increasing by 1. This is important  for
			// further processing.
			return nil, fmt.Errorf("events list is not valid")
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
// - The vote extension unmarshals
// - AssetsLocked events are valid (positive sequence number, positive amount,
//   proper bech32 recipient) and form a sequence strictly increasing by 1
// - The number of AssetsLocked events does not exceed the limit
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
			//  the Ethereum sidecar returns no events.
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

		if len(voteExtension.AssetsLockedEvents) > AssetsLockedEventsLimit {
			// Make sure the number of events does not exceed the limit.
			return &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			}, fmt.Errorf("number of events exceeds the limit")
		}

		if !bridgetypes.AssetsLockedEvents(voteExtension.AssetsLockedEvents).IsValid() {
			// Make sure all events in the slice are valid (positive sequence
			// number, positive amount, proper bech32 recipient) and form a
			// sequence strictly increasing by 1. This is important  for
			// further processing.
			return &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			}, fmt.Errorf("events list is not valid")
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
