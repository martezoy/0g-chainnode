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
	for epoch, signers := range gs.SignersByEpoch {
		keeper.SetEpochSignerSet(ctx, uint64(epoch), *signers)
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
	epochSignerSets := make([]*types.EpochSignerSet, 0)
	for i := 0; i < int(epochNumber); i += 1 {
		epochSignerSet, found := keeper.GetEpochSignerSet(ctx, uint64(i))
		if !found {
			panic("historical epoch signer set not found")
		}
		epochSignerSets = append(epochSignerSets, &epochSignerSet)
	}
	return types.NewGenesisState(params, epochNumber, signers, epochSignerSets)
}
