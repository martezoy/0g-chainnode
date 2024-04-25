package ante_test

import (
	"strings"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		EvmDenom: chaincfg.BaseDenom,
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
			"zero a0gi gas price",
			mustParseDecCoins("0a0gi"),
			mustParseDecCoins("0a0gi"),
		},
		{
			"non-zero a0gi gas price",
			mustParseDecCoins("0.001a0gi"),
			mustParseDecCoins("0.001a0gi"),
		},
		{
			"zero a0gi gas price, min neuron price",
			mustParseDecCoins("0a0gi;100000neuron"),
			mustParseDecCoins("0a0gi"), // neuron is removed
		},
		{
			"zero a0gi gas price, min neuron price, other token",
			mustParseDecCoins("0a0gi;100000neuron;0.001other"),
			mustParseDecCoins("0a0gi;0.001other"), // neuron is removed
		},
		{
			"non-zero a0gi gas price, min neuron price",
			mustParseDecCoins("0.25a0gi;100000neuron;0.001other"),
			mustParseDecCoins("0.25a0gi;0.001other"), // neuron is removed
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
