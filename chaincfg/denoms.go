package chaincfg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	StandardDenom = "a0gi"

	GasDenom = "ua0gi"

	EvmDenom = "neuron"

	BondDenom = EvmDenom

	GasDenomUnit = 6

	EvmDenomUnit = 18

	GasDenomConversionMultiplier = 1e12
	EvmDenomConversionMultiplier = 1e18
)

// RegisterDenoms registers the base and gas denominations to the SDK.
func RegisterDenoms() {
	if err := sdk.RegisterDenom(StandardDenom, sdk.OneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(GasDenom, sdk.NewDecWithPrec(1, GasDenomUnit)); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(EvmDenom, sdk.NewDecWithPrec(1, EvmDenomUnit)); err != nil {
		panic(err)
	}
}
