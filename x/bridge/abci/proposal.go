package abci

import (
	"cosmossdk.io/log"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"slices"
)

// VoteExtensionDecomposer is a function that decomposes a composite app-level
// vote extension and returns the part that is relevant to the bridge.
type VoteExtensionDecomposer func(compositeVoteExtensionBytes []byte) (
	voteExtensionBytes []byte,
	err error,
)

// ProposalHandler is the bridge-specific handler for the PrepareProposal and
// ProcessProposal ABCI requests.
type ProposalHandler struct {
	logger                  log.Logger
	valStore                bridgetypes.ValidatorStore
	voteExtensionDecomposer VoteExtensionDecomposer
}

// NewProposalHandler creates a new ProposalHandler instance.
func NewProposalHandler(
	logger log.Logger,
	valStore bridgetypes.ValidatorStore,
	voteExtensionDecomposer VoteExtensionDecomposer,
) *ProposalHandler {
	return &ProposalHandler{
		logger:                  logger,
		valStore:                valStore,
		voteExtensionDecomposer: voteExtensionDecomposer,
	}
}

// TODO: Document this function.
func (ph *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(
		ctx sdk.Context,
		req *cmtabci.RequestPrepareProposal,
	) (*cmtabci.ResponsePrepareProposal, error) {
		bridgeValsConsAddrs := ph.valStore.GetValidatorsConsAddrsByPrivilege(
			bridgetypes.ValidatorPrivilege,
		)

		// The total voting power of validators with the bridge privilege
		// (totalVP - nonBridgeValsTotalVP).
		bridgeValsTotalVP := int64(0)
		// The total voting power of validators without the bridge privilege
		// (totalVP - bridgeValsTotalVP)
		nonBridgeValsTotalVP := int64(0)

		// req.LocalLastCommit.Votes is actually a list of validators in the
		// last CometBFT validator set, with voting information regarding
		// the last block included. That means validators who did not vote
		// for the last block are not included in this list. Such votes
		// have an appropriate value of the BlockIdFlag set and their
		// vote extension is empty. This loop must take this into account.
		for _, vote := range req.LocalLastCommit.Votes {
			valConsAddr := sdk.ConsAddress(vote.Validator.Address)

			isBridgeVal := slices.ContainsFunc(
				bridgeValsConsAddrs,
				func(address sdk.ConsAddress) bool {
					return address.Equals(valConsAddr)
				},
			)

			if vp := vote.Validator.Power; isBridgeVal {
				bridgeValsTotalVP += vp
			} else {
				nonBridgeValsTotalVP += vp
			}

			// The vote extension validators vote on and which is delivered
			// here is an app-level composite vote extension that contains
			// multiple vote extension parts. This bridge handler must decompose
			// it using the provided VoteExtensionDecomposer and process
			// just the vote extension part that is relevant to the bridge.
			compositeVoteExtension := vote.VoteExtension

			if len(compositeVoteExtension) == 0 {
				// TODO: Log this event.
				continue
			}

			voteExtensionBytes, err := ph.voteExtensionDecomposer(
				compositeVoteExtension,
			)
			if err != nil {
				// TODO: Log this event.
				continue
			}

			if len(voteExtensionBytes) == 0 {
				// TODO: Log this event.
				continue
			}

			var voteExtension types.VoteExtension
			if err := voteExtension.Unmarshal(voteExtensionBytes); err != nil {
				// TODO: Log this event.
				continue
			}

			if len(voteExtension.AssetsLockedEvents) == 0 {
				// TODO: Log this event.
				continue
			}

			// ABCI++ specification requires that PrepareProposal validates
			// vote extensions in the same manner as done in VerifyVoteExtension.
			// This is because extensions of votes included in the commit info
			// after the minimum of +2/3 had been reached are not verified.
			// See: https://docs.cometbft.com/v0.38/spec/abci/abci++_methods#prepareproposal
			if err := validateAssetsLockedEvents(voteExtension.AssetsLockedEvents); err != nil {
				// TODO: Log this event.
				continue
			}

			// TODO: Record asset locked events from this vote extension.
			//       Add the voting power of this validator to the non-bridge
			//       or bridge voting power counter of each asset locked event.
		}

		// TODO: Aggregate a common assets locked events sequence based
		//       on the voting power behind each event. Make sure the
		//       aggregated sequence starts just after the current sequence
		//       tip.

		// TODO: Inject the pseudo-transaction into the proposal.

		return nil, nil
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

