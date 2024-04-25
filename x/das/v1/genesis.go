package das

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0glabs/0g-chain/x/das/v1/keeper"
	"github.com/0glabs/0g-chain/x/das/v1/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	keeper.SetNextRequestID(ctx, gs.NextRequestID)
	for _, req := range gs.Requests {
		keeper.SetDASRequest(ctx, req)
	}
	for _, resp := range gs.Responses {
		keeper.SetDASResponse(ctx, resp)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	nextRequestID, err := keeper.GetNextRequestID(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(
		nextRequestID,
		keeper.GetDASRequests(ctx),
		keeper.GetDASResponses(ctx),
	)
}
