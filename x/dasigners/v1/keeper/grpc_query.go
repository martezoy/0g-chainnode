package keeper

import (
	"context"

	"github.com/0glabs/0g-chain/crypto/bn254util"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Signer(
	c context.Context,
	req *types.QuerySignerRequest,
) (*types.QuerySignerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	signer, found, err := k.GetSigner(ctx, req.Account)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &types.QuerySignerResponse{Signer: &signer}, nil
}

func (k Keeper) EpochNumber(
	c context.Context,
	_ *types.QueryEpochNumberRequest,
) (*types.QueryEpochNumberResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryEpochNumberResponse{EpochNumber: epochNumber}, nil
}

func (k Keeper) EpochSignerSet(c context.Context, request *types.QueryEpochSignerSetRequest) (*types.QueryEpochSignerSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	epochSignerSet := make([]*types.Signer, 0)
	signers, found := k.GetEpochSignerSet(ctx, request.EpochNumber)
	if !found {
		return &types.QueryEpochSignerSetResponse{Signers: epochSignerSet}, types.ErrEpochSignerSetNotFound
	}
	for _, account := range signers.Signers {
		signer, found, err := k.GetSigner(ctx, account)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, types.ErrSignerNotFound
		}
		epochSignerSet = append(epochSignerSet, &signer)
	}
	return &types.QueryEpochSignerSetResponse{Signers: epochSignerSet}, nil
}

func (k Keeper) AggregatePubkeyG1(c context.Context, request *types.QueryAggregatePubkeyG1Request) (*types.QueryAggregatePubkeyG1Response, error) {
	ctx := sdk.UnwrapSDKContext(c)
	signers, found := k.GetEpochSignerSet(ctx, request.EpochNumber)
	if !found {
		return nil, types.ErrEpochSignerSetNotFound
	}
	if len(request.SignersBitmap) != (len(signers.Signers)+7)/8 {
		return nil, types.ErrSignerLengthNotMatch
	}
	aggPubkeyG1 := new(bn254.G1Affine)
	for i, account := range signers.Signers {
		b := request.SignersBitmap[i/8] & (1 << (i % 8))
		if b == 0 {
			continue
		}
		signer, found, err := k.GetSigner(ctx, account)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, types.ErrSignerNotFound
		}
		aggPubkeyG1.Add(aggPubkeyG1, bn254util.DeserializeG1(signer.PubkeyG1))
	}
	return &types.QueryAggregatePubkeyG1Response{
		AggregatePubkeyG1: bn254util.SerializeG1(aggPubkeyG1),
	}, nil
}
