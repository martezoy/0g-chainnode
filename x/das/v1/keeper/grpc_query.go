package keeper

import (
	"context"

	"github.com/0glabs/0g-chain/x/das/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) NextRequestID(
	c context.Context,
	_ *types.QueryNextRequestIDRequest,
) (*types.QueryNextRequestIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	nextRequestID, err := k.GetNextRequestID(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryNextRequestIDResponse{NextRequestID: nextRequestID}, nil
}
