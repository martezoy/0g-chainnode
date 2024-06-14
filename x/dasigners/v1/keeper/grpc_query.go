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
	request *types.QuerySignerRequest,
) (*types.QuerySignerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	n := len(request.Accounts)
	response := types.QuerySignerResponse{Signer: make([]*types.Signer, n)}
	for i := 0; i < n; i += 1 {
		signer, found, err := k.GetSigner(ctx, request.Accounts[i])
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, nil
		}
		response.Signer[i] = &signer
	}
	return &response, nil
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

func (k Keeper) QuorumCount(
	c context.Context,
	request *types.QueryQuorumCountRequest,
) (*types.QueryQuorumCountResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	quorumCount, err := k.GetQuorumCount(ctx, request.EpochNumber)
	if err != nil {
		return nil, err
	}
	return &types.QueryQuorumCountResponse{QuorumCount: quorumCount}, nil
}

func (k Keeper) EpochQuorum(c context.Context, request *types.QueryEpochQuorumRequest) (*types.QueryEpochQuorumResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	quorums, found := k.GetEpochQuorums(ctx, request.EpochNumber)
	if !found {
		return nil, types.ErrQuorumNotFound
	}
	if len(quorums.Quorums) <= int(request.QuorumId) {
		return nil, types.ErrQuorumIdOutOfBound
	}
	return &types.QueryEpochQuorumResponse{Quorum: quorums.Quorums[request.QuorumId]}, nil
}

func (k Keeper) EpochQuorumRow(c context.Context, request *types.QueryEpochQuorumRowRequest) (*types.QueryEpochQuorumRowResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	quorums, found := k.GetEpochQuorums(ctx, request.EpochNumber)
	if !found {
		return nil, types.ErrQuorumNotFound
	}
	if len(quorums.Quorums) <= int(request.QuorumId) {
		return nil, types.ErrQuorumIdOutOfBound
	}
	signers := quorums.Quorums[request.QuorumId].Signers
	if len(signers) <= int(request.RowIndex) {
		return nil, types.ErrRowIndexOutOfBound
	}
	return &types.QueryEpochQuorumRowResponse{Signer: signers[request.RowIndex]}, nil
}

func (k Keeper) AggregatePubkeyG1(c context.Context, request *types.QueryAggregatePubkeyG1Request) (*types.QueryAggregatePubkeyG1Response, error) {
	ctx := sdk.UnwrapSDKContext(c)
	quorums, found := k.GetEpochQuorums(ctx, request.EpochNumber)
	if !found {
		return nil, types.ErrQuorumNotFound
	}
	if len(quorums.Quorums) <= int(request.QuorumId) {
		return nil, types.ErrQuorumIdOutOfBound
	}
	quorum := quorums.Quorums[request.QuorumId]
	if (len(quorum.Signers)+7)/8 != len(request.QuorumBitmap) {
		return nil, types.ErrQuorumBitmapLengthMismatch
	}
	aggPubkeyG1 := new(bn254.G1Affine)
	hit := 0
	added := make(map[string]struct{})
	for i, signer := range quorum.Signers {
		if _, ok := added[signer]; ok {
			hit += 1
			continue
		}
		b := request.QuorumBitmap[i/8] & (1 << (i % 8))
		if b == 0 {
			continue
		}
		hit += 1
		added[signer] = struct{}{}
		signer, found, err := k.GetSigner(ctx, signer)
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
		Total:             uint64(len(quorum.Signers)),
		Hit:               uint64(hit),
	}, nil
}
