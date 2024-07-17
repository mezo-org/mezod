package types

// poa module event types
const (
	EventTypeProposeKick         = "propose_kick"
	EventTypeKickValidator       = "kick_validator"
	EventTypeLeaveValidatorSet   = "leave_validator_set"
	EventTypeApproveKickProposal = "approve_kick_proposal"
	EventTypeRejectKickProposal  = "reject_kick_proposal"
	EventTypeKeepValidator       = "keep_validator"
	AttributeKeyValidator = "validator"
	AttributeKeyProposer  = "proposer"
	AttributeValueCategory = ModuleName
)
