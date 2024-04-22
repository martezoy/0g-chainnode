package types_test

import (
	"github.com/0glabs/0g-chain/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	kavaConfig := sdk.GetConfig()
	app.SetBech32AddressPrefixes(kavaConfig)
	app.SetBip44CoinType(kavaConfig)
	kavaConfig.Seal()
}
