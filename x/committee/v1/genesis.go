package committee

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0glabs/0g-chain/x/committee/v1/keeper"
	"github.com/0glabs/0g-chain/x/committee/v1/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	params := gs.Params
	err := keeper.SetParams(ctx, params)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}

	keeper.SetCurrentCommitteeID(ctx, gs.CurrentCommitteeID)

	for _, p := range gs.Committees {
		keeper.SetCommittee(ctx, p)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	startHeight, err := keeper.GetVotingStartHeight(ctx)
	if err != nil {
		panic(err)
	}

	period, err := keeper.GetVotingPeriod(ctx)
	if err != nil {
		panic(err)
	}

	currentID, err := keeper.GetCurrentCommitteeID(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(
		keeper.GetParams(ctx),
		startHeight,
		period,
		currentID,
		keeper.GetCommittees(ctx),
	)
}
