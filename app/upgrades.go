package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	UpgradeName_Mainnet = "v0.25.0"
	UpgradeName_Testnet = "v0.25.0-alpha.0"
	UpgradeName_E2ETest = "v0.25.0-testing"
)

var (
	// KAVA to ukava - 6 decimals
	kavaConversionFactor = sdk.NewInt(1000_000)
	secondsPerYear       = sdk.NewInt(365 * 24 * 60 * 60)

	// 10 Million KAVA per year in staking rewards, inflation disable time 2024-01-01T00:00:00 UTC
	// CommunityParams_Mainnet = communitytypes.NewParams(
	// 	time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	// 	// before switchover
	// 	sdkmath.LegacyZeroDec(),
	// 	// after switchover - 10M KAVA to ukava per year / seconds per year
	// 	sdkmath.LegacyNewDec(10_000_000).
	// 		MulInt(kavaConversionFactor).
	// 		QuoInt(secondsPerYear),
	// )

	// Testnet -- 15 Trillion KAVA per year in staking rewards, inflation disable time 2023-11-16T00:00:00 UTC
	// CommunityParams_Testnet = communitytypes.NewParams(
	// 	time.Date(2023, 11, 16, 0, 0, 0, 0, time.UTC),
	// 	// before switchover
	// 	sdkmath.LegacyZeroDec(),
	// 	// after switchover
	// 	sdkmath.LegacyNewDec(15_000_000).
	// 		MulInt64(1_000_000). // 15M * 1M = 15T
	// 		MulInt(kavaConversionFactor).
	// 		QuoInt(secondsPerYear),
	// )

	// CommunityParams_E2E = communitytypes.NewParams(
	// 	time.Now().Add(10*time.Second).UTC(), // relative time for testing
	// 	sdkmath.LegacyNewDec(0),              // stakingRewardsPerSecond
	// 	sdkmath.LegacyNewDec(1000),           // upgradeTimeSetstakingRewardsPerSecond
	// )

	// ValidatorMinimumCommission is the new 5% minimum commission rate for validators
	ValidatorMinimumCommission = sdk.NewDecWithPrec(5, 2)
)

// RegisterUpgradeHandlers registers the upgrade handlers for the app.
func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler("v1",
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			params := minttypes.DefaultParams()
			app.mintKeeper.SetParams(ctx, params)
			moduleName := "mint"
			currentVersion, ok := fromVM[moduleName]
			if !ok {
				return nil, fmt.Errorf("module %s not found in version map", moduleName)
			}

			fromVM[moduleName] = currentVersion + 1

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})
}
