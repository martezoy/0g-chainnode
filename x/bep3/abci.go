package bep3

import (
	"time"

	"github.com/0glabs/0g-chain/x/bep3/keeper"
	"github.com/0glabs/0g-chain/x/bep3/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker on every block expires outdated atomic swaps and removes closed
// swap from long term storage (default storage time of 1 week)
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.UpdateTimeBasedSupplyLimits(ctx)
	k.UpdateExpiredAtomicSwaps(ctx)
	k.DeleteClosedAtomicSwapsFromLongtermStorage(ctx)
}
