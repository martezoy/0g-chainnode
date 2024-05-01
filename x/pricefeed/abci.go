package pricefeed

import (
	"errors"
	"time"

	"github.com/0glabs/0g-chain/x/pricefeed/keeper"
	"github.com/0glabs/0g-chain/x/pricefeed/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker updates the current pricefeed
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Update the current price of each asset.
	for _, market := range k.GetMarkets(ctx) {
		if !market.Active {
			continue
		}

		err := k.SetCurrentPrices(ctx, market.MarketID)
		if err != nil && !errors.Is(err, types.ErrNoValidPrice) {
			panic(err)
		}
	}
}
