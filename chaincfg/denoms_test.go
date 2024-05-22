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
			"standard to gas",
			MakeCoinForStandardDenom(99),
			GasDenom,
			MakeCoinForGasDenom(99 * (EvmDenomConversionMultiplier / GasDenomConversionMultiplier)),
			nil,
		},
		{
			"gas to standard",
			MakeCoinForGasDenom(5e7),
			StandardDenom,
			MakeCoinForStandardDenom(50),
			nil,
		},
		{
			"standard to base",
			MakeCoinForStandardDenom(22),
			EvmDenom,
			MakeCoinForEvmDenom(22 * EvmDenomConversionMultiplier),
			nil,
		},
		{
			"base to standard",
			MakeCoinForEvmDenom("97000000000000000000"),
			StandardDenom,
			MakeCoinForStandardDenom(97),
			nil,
		},
		{
			"gas to base",
			MakeCoinForGasDenom(33),
			EvmDenom,
			MakeCoinForEvmDenom(33 * GasDenomConversionMultiplier),
			nil,
		},
		{
			"base to gas",
			MakeCoinForEvmDenom("770000000000000"),
			GasDenom,
			MakeCoinForGasDenom(770000000000000 / GasDenomConversionMultiplier),
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
