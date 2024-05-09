package chaincfg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	StandardDenom = "a0gi"

	AuxiliaryDenom = "ua0gi"

	BaseDenom = "neuron"

	BondDenom = BaseDenom

	AuxiliaryDenomUnit = 6

	BaseDenomUnit = 18

	AuxiliaryDenomConversionMultiplier = 1e12
	BaseDenomConversionMultiplier      = 1e18
)

// RegisterDenoms registers the base and auxiliary denominations to the SDK.
func RegisterDenoms() {
	if err := sdk.RegisterDenom(StandardDenom, sdk.OneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(AuxiliaryDenom, sdk.NewDecWithPrec(1, AuxiliaryDenomUnit)); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(BaseDenom, sdk.NewDecWithPrec(1, BaseDenomUnit)); err != nil {
		panic(err)
	}
}
