package dasigners

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0glabs/0g-chain/x/dasigners/v1/keeper"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}
	keeper.SetEpochNumber(ctx, gs.EpochNumber)
	for _, signer := range gs.Signers {
		if err := keeper.SetSigner(ctx, *signer); err != nil {
			panic(fmt.Sprintf("failed to write genesis state into store: %s", err))
		}
	}
	for epoch, quorums := range gs.QuorumsByEpoch {
		keeper.SetEpochQuorums(ctx, uint64(epoch), *quorums)
	}
	keeper.SetParams(ctx, gs.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	params := keeper.GetParams(ctx)
	epochNumber, err := keeper.GetEpochNumber(ctx)
	if err != nil {
		panic(err)
	}
	signers := make([]*types.Signer, 0)
	keeper.IterateSigners(ctx, func(_ int64, signer types.Signer) (stop bool) {
		signers = append(signers, &signer)
		return false
	})
	epochQuorums := make([]*types.Quorums, 0)
	for i := 0; i < int(epochNumber); i += 1 {
		quorumCnt, err := keeper.GetQuorumCount(ctx, uint64(i))
		if err != nil {
			panic("historical quorums not found")
		}
		quorums := make([]*types.Quorum, quorumCnt)
		for quorumId := uint64(0); quorumId < quorumCnt; quorumId += 1 {
			quorum, err := keeper.GetEpochQuorum(ctx, uint64(i), quorumId)
			if err != nil {
				panic("failed to load historical quorum")
			}
			quorums[quorumId] = &quorum
		}
		epochQuorums = append(epochQuorums, &types.Quorums{Quorums: quorums})
	}
	return types.NewGenesisState(params, epochNumber, signers, epochQuorums)
}
