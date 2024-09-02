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

		voteCounter := newAssetsLockedVoteCounter()

		// req.LocalLastCommit.Votes is actually a list of validators in the
		// last CometBFT validator set, with voting information regarding
		// the last block included. That means validators who did not vote
		// for the last block are not included in this list. Such votes
		// have an appropriate value of the BlockIdFlag set and their
		// vote extension is empty. This loop must take this into account.
		for _, vote := range req.LocalLastCommit.Votes {
			valConsAddr := sdk.ConsAddress(vote.Validator.Address)
			valVP := vote.Validator.Power

			isBridgeVal := slices.ContainsFunc(
				bridgeValsConsAddrs,
				func(address sdk.ConsAddress) bool {
					return address.Equals(valConsAddr)
				},
			)

			voteCounter.registerVoter(valVP, isBridgeVal)

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

			for _, event := range voteExtension.AssetsLockedEvents {
				voteCounter.addVote(&event, valVP, isBridgeVal)
			}
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

// assetsLockedVoteCounter is a helper structure that counts votes for
// AssetsLocked events. It is responsible for determining the sequence of
// canonical AssetsLocked events supported by 2/3+ of the bridge validators
// and confirmed by 2/3+ of the non-bridge validators.
type assetsLockedVoteCounter struct {
	// A map of AssetsLocked vote infos indexed by their sequence number.
	// In case validators have a different view of the world, the given
	// sequence number may have more than one item associated with it.
	votesInfo map[string][]*assetsLockedVoteInfo
	// The total voting power of validators with the bridge privilege
	// (totalVP - nonBridgeValsTotalVP).
	bridgeValsTotalVP int64
	// The total voting power of validators without the bridge privilege
	// (totalVP - bridgeValsTotalVP)
	nonBridgeValsTotalVP int64
}

// newAssetsLockedVoteCounter creates a new assetsLockedVoteCounter instance.
func newAssetsLockedVoteCounter() *assetsLockedVoteCounter {
	return &assetsLockedVoteCounter{
		votesInfo: make(map[string][]*assetsLockedVoteInfo),
		bridgeValsTotalVP: 0,
		nonBridgeValsTotalVP: 0,
	}
}

// registerVoter registers a validator's voting power. It must be called for
// each validator able to vote, regardless of whether they voted or not.
func (alvc *assetsLockedVoteCounter) registerVoter(vp int64, isBridgeVal bool) {
	if isBridgeVal {
		alvc.bridgeValsTotalVP += vp
	} else {
		alvc.nonBridgeValsTotalVP += vp
	}
}

// addVote adds the given voting power to the sum of the bridge or non-bridge
// voting power behind the given AssetsLocked event.
func (alvc *assetsLockedVoteCounter) addVote(
	event *bridgetypes.AssetsLockedEvent,
	vp int64,
	isBridgeVal bool,
) {
	sequenceKey := event.Sequence.String()

	if index := slices.IndexFunc(
		alvc.votesInfo[sequenceKey],
		func(v *assetsLockedVoteInfo) bool {
			// It's enough to compare the recipient and the amount.
			// The sequence number is used as the votesInfo map key hence
			// it is the same for all items in the given value slice.
			return event.Recipient == v.event.Recipient &&
				event.Amount.Equal(v.event.Amount)
		},
	); index != -1 {
		alvc.votesInfo[sequenceKey][index].add(vp, isBridgeVal)
	} else {
		voteInfo := newAssetsLockedVoteInfo(event)
		voteInfo.add(vp, isBridgeVal)
		alvc.votesInfo[sequenceKey] = append(alvc.votesInfo[sequenceKey], voteInfo)
	}
}

// assetsLockedVoteInfo is a helper structure that holds the bridge and
// non-bridge voting power behind a single AssetsLocked event.
type assetsLockedVoteInfo struct {
	event           *bridgetypes.AssetsLockedEvent
	bridgeValsVP    int64
	nonBridgeValsVP int64
}

// newAssetsLockedVote creates a new assetsLockedVote instance.
func newAssetsLockedVoteInfo(
	event *bridgetypes.AssetsLockedEvent,
) *assetsLockedVoteInfo {
	return &assetsLockedVoteInfo{
		event:           event,
		bridgeValsVP:    0,
		nonBridgeValsVP: 0,
	}
}

// add adds the given voting power to the sum of the bridge or non-bridge
// voting power behind this AssetsLocked event.
func (alvi *assetsLockedVoteInfo) add(vp int64, isBridgeVal bool) {
	if isBridgeVal {
		alvi.bridgeValsVP += vp
	} else {
		alvi.nonBridgeValsVP += vp
	}
}

