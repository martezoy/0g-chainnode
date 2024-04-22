package hard

import (
	"time"

	"github.com/0glabs/0g-chain/x/hard/keeper"
	"github.com/0glabs/0g-chain/x/hard/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker updates interest rates
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.ApplyInterestRateUpdates(ctx)
}
