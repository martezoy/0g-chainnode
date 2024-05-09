package chaincfg

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestRegisterDenoms(t *testing.T) {
	RegisterDenoms()
	tests := []struct {
		name        string
		from        sdk.Coin
		targetDenom string
		expCoin     sdk.Coin
		expErr      error
	}{
		{
			"standard to auxiliary",
			MakeCoinForStandardDenom(99),
			AuxiliaryDenom,
			MakeCoinForAuxiliaryDenom(99 * (BaseDenomConversionMultiplier / AuxiliaryDenomConversionMultiplier)),
			nil,
		},
		{
			"auxiliary to standard",
			MakeCoinForAuxiliaryDenom(5e7),
			StandardDenom,
			MakeCoinForStandardDenom(50),
			nil,
		},
		{
			"standard to base",
			MakeCoinForStandardDenom(22),
			BaseDenom,
			MakeCoinForBaseDenom(22 * BaseDenomConversionMultiplier),
			nil,
		},
		{
			"base to standard",
			MakeCoinForBaseDenom("97000000000000000000"),
			StandardDenom,
			MakeCoinForStandardDenom(97),
			nil,
		},
		{
			"auxiliary to base",
			MakeCoinForAuxiliaryDenom(33),
			BaseDenom,
			MakeCoinForBaseDenom(33 * AuxiliaryDenomConversionMultiplier),
			nil,
		},
		{
			"base to auxiliary",
			MakeCoinForBaseDenom("770000000000000"),
			AuxiliaryDenom,
			MakeCoinForAuxiliaryDenom(770000000000000 / AuxiliaryDenomConversionMultiplier),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := sdk.ConvertCoin(tt.from, tt.targetDenom)
			if tt.expErr != nil {
				if err == nil {
					t.Errorf("expErr is not nil, but got nil")
					return
				}
			} else {
				if err != nil {
					t.Errorf("expErr is nil, but got %v", err)
					return
				}
			}

			assert.Equal(t, tt.expCoin, ret)
		})
	}
}
