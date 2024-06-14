package chaincfg

import (
	"github.com/shopspring/decimal"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var (
	maxBondedRatio, _ = sdk.NewDecFromStr("1.0")
	yMin, _           = sdk.NewDecFromStr("0.05")

	minBondedRatio, _ = sdk.NewDecFromStr("0.2")
	yMax, _           = sdk.NewDecFromStr("0.15") // 15% at min bonded ratio

	decayRate, _ = sdk.NewDecFromStr("10")
)

func InflationCalculateFn(ctx sdk.Context, minter minttypes.Minter, params minttypes.Params, bondedRatio sdk.Dec) sdk.Dec {
	logger := ctx.Logger()
	if logger == nil {
		panic("logger is nil")
	}
	return inflationCalculateFn(logger, minter, params, bondedRatio)
}

func decExp(x sdk.Dec) sdk.Dec {
	xDec := decimal.NewFromBigInt(x.BigInt(), -18)
	expDec, _ := xDec.ExpTaylor(18)
	expInt := expDec.Shift(18).BigInt()
	return sdk.NewDecFromBigIntWithPrec(expInt, 18)
}

func inflationCalculateFn(logger log.Logger, minter minttypes.Minter, params minttypes.Params, bondedRatio sdk.Dec) sdk.Dec {
	var apy sdk.Dec
	if bondedRatio.LT(minBondedRatio) {
		apy = yMax
	} else {
		exp := decayRate.Neg().Mul(maxBondedRatio.Sub(minBondedRatio))
		c := decExp(exp)
		d := yMin.Sub(yMax.Mul(c)).Quo(sdk.OneDec().Sub(c))
		expBonded := decayRate.Neg().Mul(bondedRatio.Sub(minBondedRatio))
		cBonded := decExp(expBonded)
		e := yMax.Sub(d).Mul(cBonded)
		apy = d.Add(e)
	}
	inflation := apy.Mul(bondedRatio)

	// // The target annual inflation rate is recalculated for each previsions cycle. The
	// // inflation is also subject to a rate change (positive or negative) depending on
	// // the distance from the desired ratio (67%). The maximum rate change possible is
	// // defined to be 13% per year, however the annual inflation is capped as between
	// // 7% and 20%.

	// // (1 - bondedRatio/GoalBonded) * InflationRateChange
	// inflationRateChangePerYear := sdk.OneDec().
	// 	Sub(bondedRatio.Quo(params.GoalBonded)).
	// 	Mul(params.InflationRateChange)
	// inflationRateChange := inflationRateChangePerYear.Quo(sdk.NewDec(int64(params.BlocksPerYear)))

	// // adjust the new annual inflation for this next cycle
	// inflation := minter.Inflation.Add(inflationRateChange) // note inflationRateChange may be negative
	// if inflation.GT(params.InflationMax) {
	// 	inflation = params.InflationMax
	// }
	// if inflation.LT(params.InflationMin) {
	// 	inflation = params.InflationMin
	// }

	logger.Info(
		"calculated new annual inflation",
		"bondedRatio", bondedRatio,
		"apy", apy,
		"inflation", inflation,
		"params", params,
		"minter", minter,
	)
	return inflation
}
