package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	UpgradeName = "v1"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	mk mintkeeper.Keeper,
	pk paramskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		subspace, _ := pk.GetSubspace(minttypes.ModuleName)
		params := minttypes.DefaultParams()
		subspace.Get(ctx, minttypes.KeyMintDenom, &params.MintDenom)
		subspace.Get(ctx, minttypes.KeyInflationRateChange, &params.InflationRateChange)
		subspace.Get(ctx, minttypes.KeyInflationMax, &params.InflationMax)
		subspace.Get(ctx, minttypes.KeyInflationMin, &params.InflationMin)
		subspace.Get(ctx, minttypes.KeyGoalBonded, &params.GoalBonded)
		subspace.Get(ctx, minttypes.KeyBlocksPerYear, &params.BlocksPerYear)
		ctx.Logger().Info("Mint module parameters", "params", params)
		mk.SetParams(ctx, params)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
