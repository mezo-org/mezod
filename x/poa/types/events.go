package types

// poa module event types
const (
	EventTypeSubmitApplication   = "submit_application"
	EventTypeAppendValidator     = "append_validator"
	EventTypeProposeKick         = "propose_kick"
	EventTypeKickValidator       = "kick_validator"
	EventTypeLeaveValidatorSet   = "leave_validator_set"
	EventTypeApproveApplication  = "approve_application"
	EventTypeRejectApplication   = "reject_application"
	EventTypeRejectValidator     = "reject_validator"
	EventTypeApproveKickProposal = "approve_kick_proposal"
	EventTypeRejectKickProposal  = "reject_kick_proposal"
	EventTypeKeepValidator       = "keep_validator"

	AttributeKeyValidator = "validator"
	AttributeKeyCandidate = "candidate"
	AttributeKeyVoter     = "voter"
	AttributeKeyProposer  = "proposer"

	AttributeValueCategory = ModuleName
)
