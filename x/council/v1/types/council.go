package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Councils []Council
type Votes []Vote

func (c Council) HasVotingEndedBy(height int64) bool {
	return height >= int64(c.StartHeight)
}

// NewVote instantiates a new instance of Vote
func NewVote(councilID uint64, voter sdk.ValAddress, ballots []*Ballot) Vote {
	return Vote{
		CouncilID: councilID,
		Voter:     voter,
		Ballots:   ballots,
	}
}
