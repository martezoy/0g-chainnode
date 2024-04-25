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
	councilID, err := k.GetCurrentCouncilID(ctx)
	if err != nil {
		// TODO: handle the case where councilID is not available
		return
	}
	council, bz := k.GetCouncil(ctx, councilID)
	if !bz {
		return
	}

	if ctx.BlockHeight() >= int64(council.StartHeight) {
		// We are ready to accept votes for the next council
		if err := k.StoreNewCouncil(ctx, council.StartHeight); err != nil {
			return
		}
	}

	if ctx.BlockHeight() < int64(council.EndHeight) {
		return
	}

	k.IncrementCurrentCouncilID(ctx)
	council, bz = k.GetCouncil(ctx, councilID+1)
	if !bz {
		return
	}

	ballots := []Ballot{}
	seen := make(map[string]struct{})
	for _, vote := range council.Votes {
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

	councilSize := k.GetParams(ctx).CouncilSize
	council.Members = make([]sdk.ValAddress, councilSize)
	for i := 0; i < int(councilSize); i = i + 1 {
		council.Members[i] = ballots[i].voter
	}

	k.SetCouncil(ctx, council)
}

func (k *Keeper) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) {
}
