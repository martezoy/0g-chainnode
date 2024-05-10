package ante_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/0glabs/0g-chain/app"
	"github.com/0glabs/0g-chain/app/ante"
	"github.com/0glabs/0g-chain/chaincfg"
)

func mustParseDecCoins(value string) sdk.DecCoins {
	coins, err := sdk.ParseDecCoins(strings.ReplaceAll(value, ";", ","))
	if err != nil {
		panic(err)
	}

	return coins
}

func TestEvmMinGasFilter(t *testing.T) {
	tApp := app.NewTestApp()
	handler := ante.NewEvmMinGasFilter(tApp.GetEvmKeeper())

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	tApp.GetEvmKeeper().SetParams(ctx, evmtypes.Params{
		EvmDenom: chaincfg.EvmDenom,
	})

	testCases := []struct {
		name                 string
		minGasPrices         sdk.DecCoins
		expectedMinGasPrices sdk.DecCoins
	}{
		{
			"no min gas prices",
			mustParseDecCoins(""),
			mustParseDecCoins(""),
		},
		{
			"zero ua0gi gas price",
			mustParseDecCoins("0ua0gi"),
			mustParseDecCoins("0ua0gi"),
		},
		{
			"non-zero ua0gi gas price",
			mustParseDecCoins("0.001ua0gi"),
			mustParseDecCoins("0.001ua0gi"),
		},
		{
			"zero ua0gi gas price, min neuron price",
			mustParseDecCoins("0ua0gi;100000neuron"),
			mustParseDecCoins("0ua0gi"), // neuron is removed
		},
		{
			"zero ua0gi gas price, min neuron price, other token",
			mustParseDecCoins("0ua0gi;100000neuron;0.001other"),
			mustParseDecCoins("0ua0gi;0.001other"), // neuron is removed
		},
		{
			"non-zero ua0gi gas price, min neuron price",
			mustParseDecCoins("0.25ua0gi;100000neuron;0.001other"),
			mustParseDecCoins("0.25ua0gi;0.001other"), // neuron is removed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

			ctx = ctx.WithMinGasPrices(tc.minGasPrices)
			mmd := MockAnteHandler{}

			_, err := handler.AnteHandle(ctx, nil, false, mmd.AnteHandle)
			require.NoError(t, err)
			require.True(t, mmd.WasCalled)

			assert.NoError(t, mmd.CalledCtx.MinGasPrices().Validate())
			assert.Equal(t, tc.expectedMinGasPrices, mmd.CalledCtx.MinGasPrices())
		})
	}
}
