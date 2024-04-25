// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"

	"github.com/0glabs/0g-chain/x/council/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ types.MsgServer = &Keeper{}

// Register handles MsgRegister messages
func (k Keeper) Register(goCtx context.Context, msg *types.MsgRegister) (*types.MsgRegisterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, err
	}

	_, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, stakingtypes.ErrNoValidatorFound
	}

	if err := k.AddVoter(ctx, valAddr, msg.Key); err != nil {
		return nil, err
	}

	return &types.MsgRegisterResponse{}, nil
}

// Vote handles MsgVote messages
func (k Keeper) Vote(goCtx context.Context, msg *types.MsgVote) (*types.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	voter, err := sdk.ValAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, err
	}

	if err := k.AddVote(ctx, msg.CouncilID, voter, msg.Ballots); err != nil {
		return nil, err
	}

	return &types.MsgVoteResponse{}, nil
}
