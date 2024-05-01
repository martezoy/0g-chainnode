package chaincfg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DisplayDenom defines the denomination displayed to users in client applications.
	DisplayDenom = "a0gi"
	// BaseDenom defines to the default denomination used in 0g-chain
	BaseDenom = "neuron"

	BaseDenomUnit = 18

	ConversionMultiplier = 1e18
)

// RegisterDenoms registers the base and display denominations to the SDK.
func RegisterDenoms() {
	if err := sdk.RegisterDenom(DisplayDenom, sdk.OneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(BaseDenom, sdk.NewDecWithPrec(1, BaseDenomUnit)); err != nil {
		panic(err)
	}
}
