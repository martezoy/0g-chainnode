package chaincfg

import (
	"github.com/shopspring/decimal"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var (
	Xmax, _ = sdk.NewDecFromStr("1.0")  // upper limit on staked supply (as % of circ supply)
	Ymin, _ = sdk.NewDecFromStr("0.05") // target APY at upper limit

	Xmin, _ = sdk.NewDecFromStr("0.2")  // lower limit on staked supply (as % of circ supply)
	Ymax, _ = sdk.NewDecFromStr("0.15") // target APY at lower limit

	decayRate, _ = sdk.NewDecFromStr("10")
)

func decExp(x sdk.Dec) sdk.Dec {
	xDec := decimal.NewFromBigInt(x.BigInt(), -18)
	expDec, _ := xDec.ExpTaylor(18)
	expInt := expDec.Shift(18).BigInt()
	return sdk.NewDecFromBigIntWithPrec(expInt, 18)
}

func NextInflationRate(ctx sdk.Context, minter minttypes.Minter, params minttypes.Params, bondedRatio sdk.Dec, circulatingRatio sdk.Dec) sdk.Dec {
	X := bondedRatio.Quo(circulatingRatio)

	var apy sdk.Dec
	if X.LT(Xmin) {
		apy = Ymax
	} else {
		exp := decayRate.Neg().Mul(Xmax.Sub(Xmin))
		c := decExp(exp)
		d := Ymin.Sub(Ymax.Mul(c)).Quo(sdk.OneDec().Sub(c))
		expBonded := decayRate.Neg().Mul(X.Sub(Xmin))
		cBonded := decExp(expBonded)
		e := Ymax.Sub(d).Mul(cBonded)
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

	ctx.Logger().Info(
		"nextInflationRate",
		"bondedRatio", bondedRatio,
		"circulatingRatio", circulatingRatio,
		"apy", apy,
		"inflation", inflation,
		"params", params,
		"minter", minter,
	)
	return inflation
}
