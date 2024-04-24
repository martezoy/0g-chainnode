package keeper

import (
	"sort"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Ballot struct {
	voter   sdk.ValAddress
	content string
}

func (k *Keeper) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	committeeID, err := k.GetCurrentCommitteeID(ctx)
	if err != nil {
		// TODO: handle the case where committeeID is not available
		return
	}
	committee, bz := k.GetCommittee(ctx, committeeID)
	if !bz {
		return
	}

	if ctx.BlockHeight() >= int64(committee.StartHeight) {
		// We are ready to accept votes for the next committee
		if err := k.StoreNewCommittee(ctx, committee.StartHeight); err != nil {
			return
		}
	}

	if ctx.BlockHeight() < int64(committee.EndHeight) {
		return
	}

	k.IncrementCurrentCommitteeID(ctx)
	committee, bz = k.GetCommittee(ctx, committeeID+1)
	if !bz {
		return
	}

	ballots := []Ballot{}
	seen := make(map[string]struct{})
	for _, vote := range committee.Votes {
		for _, ballot := range vote.Ballots {
			ballot := Ballot{
				voter:   vote.Voter,
				content: string(ballot.Content),
			}
			if _, ok := seen[ballot.content]; ok {
				continue
			}
			ballots = append(ballots, ballot)
			seen[ballot.content] = struct{}{}
		}
	}
	sort.Slice(ballots, func(i, j int) bool {
		return ballots[i].content < ballots[j].content
	})

	committeeSize := k.GetParams(ctx).CommitteeSize
	committee.Members = make([]sdk.ValAddress, committeeSize)
	for i := 0; i < int(committeeSize); i = i + 1 {
		committee.Members[i] = ballots[i].voter
	}

	k.SetCommittee(ctx, committee)
}

func (k *Keeper) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) {
}
