package keeper

import (
	"context"

	"github.com/0glabs/0g-chain/x/council/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) CurrentCouncilID(
	c context.Context,
	_ *types.QueryCurrentCouncilIDRequest,
) (*types.QueryCurrentCouncilIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	currentCouncilID, err := k.GetCurrentCouncilID(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryCurrentCouncilIDResponse{CurrentCouncilID: currentCouncilID}, nil
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
