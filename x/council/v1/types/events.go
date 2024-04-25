package types

// Module event types
const (
	EventTypeRegister = "register"
	EventTypeVote     = "vote"

	AttributeValueCategory          = "council"
	AttributeKeyCouncilID           = "council_id"
	AttributeKeyProposalID          = "proposal_id"
	AttributeKeyVotingStartHeight   = "voting_start_height"
	AttributeKeyVotingEndHeight     = "voting_end_height"
	AttributeKeyProposalCloseStatus = "status"
	AttributeKeyVoter               = "voter"
	AttributeKeyBallots             = "ballots"
	AttributeKeyPublicKey           = "public_key"
	AttributeKeyProposalOutcome     = "proposal_outcome"
	AttributeKeyProposalTally       = "proposal_tally"
)
