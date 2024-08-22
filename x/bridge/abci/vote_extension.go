package abci

import (
	"fmt"

	"cosmossdk.io/math"

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
	sidecarClient types.EthereumSidecarClient
	bridgeKeeper  keeper.Keeper
}

// NewVoteExtensionHandler creates a new VoteExtensionHandler instance.
func NewVoteExtensionHandler(
	sidecarClient types.EthereumSidecarClient,
	bridgeKeeper keeper.Keeper,
) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		sidecarClient: sidecarClient,
		bridgeKeeper:  bridgeKeeper,
	}
}

// ExtendVoteHandler returns the handler for the ExtendVote ABCI request.
// It fetches the AssetsLocked events from the sidecar and includes them in the
// vote extension. Events are fetched starting from the currently stored
// sequence tip + 1 and up to the sequence tip + AssetsLockedEventsLimit.
//
// Dev note: It is fine to return a nil response and an error from this
// function in case of failure. The upstream app-level vote extension handler
// will handle the error gracefully and set the bridge-specific part of the
// vote extension to be empty.
func (veh *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(
		ctx sdk.Context,
		_ *cmtabci.RequestExtendVote,
	) (*cmtabci.ResponseExtendVote, error) {
		// TODO: Fetched events will be finalized in the next block and the
		//       tip will be updated then. Because of that, we may fetch the same
		//       set of events twice, in two subsequent blocks. Once the full
		//       flow is implemented, we will need to check this behavior and
		//       fix it if necessary.

		sequenceTip := veh.bridgeKeeper.GetAssetsLockedSequenceTip(ctx)

		sequenceStart := sequenceTip.Add(math.NewInt(1))
		sequenceEnd := sequenceStart.Add(math.NewInt(AssetsLockedEventsLimit))

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

		// TODO: Sort events by sequence.

		voteExtension := types.VoteExtension{
			AssetsLockedEvents: events,
		}
		// Marshal the vote extension into bytes.
		voteExtensionBytes, err := voteExtension.Marshal()
		if err != nil {
			// If marshaling fails, we cannot recover, so return an error.
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		return &cmtabci.ResponseExtendVote{VoteExtension: voteExtensionBytes}, nil
	}
}

// TODO: Implement and document the VerifyVoteExtensionHandler function.
func (veh *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(
		_ sdk.Context,
		_ *cmtabci.RequestVerifyVoteExtension,
	) (*cmtabci.ResponseVerifyVoteExtension, error) {
		return &cmtabci.ResponseVerifyVoteExtension{Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
