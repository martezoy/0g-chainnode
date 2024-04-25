package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrUnknownCouncil          = errorsmod.Register(ModuleName, 2, "council not found")
	ErrInvalidCouncil          = errorsmod.Register(ModuleName, 3, "invalid council")
	ErrUnknownProposal         = errorsmod.Register(ModuleName, 4, "proposal not found")
	ErrProposalExpired         = errorsmod.Register(ModuleName, 5, "proposal expired")
	ErrInvalidPubProposal      = errorsmod.Register(ModuleName, 6, "invalid pubproposal")
	ErrUnknownVote             = errorsmod.Register(ModuleName, 7, "vote not found")
	ErrInvalidGenesis          = errorsmod.Register(ModuleName, 8, "invalid genesis")
	ErrNoProposalHandlerExists = errorsmod.Register(ModuleName, 9, "pubproposal has no corresponding handler")
	ErrUnknownSubspace         = errorsmod.Register(ModuleName, 10, "subspace not found")
	ErrInvalidVoteType         = errorsmod.Register(ModuleName, 11, "invalid vote type")
	ErrNotFoundProposalTally   = errorsmod.Register(ModuleName, 12, "proposal tally not found")
	ErrInvalidPublicKey        = errorsmod.Register(ModuleName, 13, "invalid public key")
	ErrInvalidValidatorAddress = errorsmod.Register(ModuleName, 14, "invalid validator address")
)
