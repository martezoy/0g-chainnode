package keeper

import (
	"context"

	"github.com/0glabs/0g-chain/x/das/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ types.MsgServer = &Keeper{}

// RequestDAS handles MsgRequestDAS messages
func (k Keeper) RequestDAS(
	goCtx context.Context, msg *types.MsgRequestDAS,
) (*types.MsgRequestDASResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	requestID, err := k.StoreNewDASRequest(ctx, msg.StreamID, msg.BatchHeaderHash, msg.NumBlobs)
	if err != nil {
		return nil, err
	}
	k.IncrementNextRequestID(ctx)
	return &types.MsgRequestDASResponse{
		RequestID: requestID,
	}, nil
}

// ReportDASResult handles MsgReportDASResult messages
func (k Keeper) ReportDASResult(
	goCtx context.Context, msg *types.MsgReportDASResult,
) (*types.MsgReportDASResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sampler, err := sdk.ValAddressFromBech32(msg.Sampler)
	if err != nil {
		return nil, err
	}

	if _, found := k.stakingKeeperRef.GetValidator(ctx, sampler); !found {
		return nil, stakingtypes.ErrNoValidatorFound
	}

	if err := k.StoreNewDASResponse(ctx, msg.RequestID, sampler, msg.Results); err != nil {
		return nil, err
	}

	return &types.MsgReportDASResultResponse{}, nil
}
