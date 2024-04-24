package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Committees []Committee
type Votes []Vote

func (c Committee) HasVotingEndedBy(height int64) bool {
	return height >= int64(c.StartHeight)
}

// NewVote instantiates a new instance of Vote
func NewVote(committeeID uint64, voter sdk.ValAddress, ballots []*Ballot) Vote {
	return Vote{
		CommitteeID: committeeID,
		Voter:       voter,
		Ballots:     ballots,
	}
}
