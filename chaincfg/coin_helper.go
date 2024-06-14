package chaincfg

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/shopspring/decimal"
)

func toBigInt(amount any) *big.Int {
	if amount == nil {
		return big.NewInt(0)
	}
	var val *big.Int
	switch amount.(type) {
	case int:
		val = big.NewInt(int64(amount.(int)))
	case int32:
		val = big.NewInt(int64(amount.(int32)))
	case int64:
		val = big.NewInt(amount.(int64))
	case string:
		var ok bool
		val, ok = new(big.Int).SetString(amount.(string), 0)
		if !ok {
			panic(fmt.Sprintf("invalid amount string: %s", amount.(string)))
		}
	case math.Int:
		val = amount.(math.Int).BigInt()
	case *big.Int:
		val = amount.(*big.Int)
	case float64:
		val = decimal.NewFromFloat(amount.(float64)).BigInt()
	default:
		panic(fmt.Sprintf("invalid amount type: %T", amount))
	}

	return val
}

func MakeCoinForStandardDenom(amount any) sdk.Coin {
	return makeCoin(StandardDenom, toBigInt(amount))
}

func MakeCoinForGasDenom(amount any) sdk.Coin {
	return makeCoin(GasDenom, toBigInt(amount))
}

func MakeCoinForEvmDenom(amount any) sdk.Coin {
	return makeCoin(EvmDenom, toBigInt(amount))
}

func makeCoin(denom string, amount *big.Int) sdk.Coin {
	return sdk.NewCoin(denom, math.NewIntFromBigInt(amount))
}
