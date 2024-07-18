package app

import (
	v1 "github.com/0glabs/0g-chain/app/upgrades/v1"
)

// RegisterUpgradeHandlers registers the upgrade handlers for the app.
func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(v1.UpgradeName,
		v1.CreateUpgradeHandler(app.mm, app.configurator, app.mintKeeper, app.paramsKeeper))
}
