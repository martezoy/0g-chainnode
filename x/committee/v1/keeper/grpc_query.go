package keeper

import (
	"context"

	"github.com/0glabs/0g-chain/x/committee/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) CurrentCommitteeID(
	c context.Context,
	_ *types.QueryCurrentCommitteeIDRequest,
) (*types.QueryCurrentCommitteeIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	currentCommitteeID, err := k.GetCurrentCommitteeID(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryCurrentCommitteeIDResponse{CurrentCommitteeID: currentCommitteeID}, nil
}

func (k Keeper) RegisteredVoters(
	c context.Context,
	_ *types.QueryRegisteredVotersRequest,
) (*types.QueryRegisteredVotersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	voterAddrs := k.GetVoters(ctx)
	voters := make([]string, len(voterAddrs))
	for i, voterAddr := range voterAddrs {
		voters[i] = voterAddr.String()
	}
	return &types.QueryRegisteredVotersResponse{Voters: voters}, nil
}
