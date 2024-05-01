package e2e_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/0glabs/0g-chain/app"
	"github.com/0glabs/0g-chain/chaincfg"
	"github.com/0glabs/0g-chain/tests/util"
)

func (suite *IntegrationTestSuite) TestEthGasPriceReturnsMinFee() {
	suite.SkipIfKvtoolDisabled()

	// read expected min fee from app.toml
	minGasPrices, err := getMinFeeFromAppToml(util.ZgChainHomePath())
	suite.NoError(err)

	// evm uses neuron, get neuron min fee
	evmMinGas := minGasPrices.AmountOf(chaincfg.BaseDenom).TruncateInt().BigInt()

	// returns eth_gasPrice, units in a0gi
	gasPrice, err := suite.ZgChain.EvmClient.SuggestGasPrice(context.Background())
	suite.NoError(err)

	suite.Equal(evmMinGas, gasPrice)
}

func (suite *IntegrationTestSuite) TestEvmRespectsMinFee() {
	suite.SkipIfKvtoolDisabled()

	// setup sender & receiver
	sender := suite.ZgChain.NewFundedAccount("evm-min-fee-test-sender", sdk.NewCoins(a0gi(big.NewInt(1e3))))
	randoReceiver := util.SdkToEvmAddress(app.RandomAddress())

	// get min gas price for evm (from app.toml)
	minFees, err := getMinFeeFromAppToml(util.ZgChainHomePath())
	suite.NoError(err)
	minGasPrice := minFees.AmountOf(chaincfg.BaseDenom).TruncateInt()

	// attempt tx with less than min gas price (min fee - 1)
	tooLowGasPrice := minGasPrice.Sub(sdk.OneInt()).BigInt()
	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(0, randoReceiver, big.NewInt(5e2), 1e5, tooLowGasPrice, nil),
		Data: "this tx should fail because it's gas price is too low",
	}
	res := sender.SignAndBroadcastEvmTx(req)

	// expect the tx to fail!
	suite.ErrorAs(res.Err, &util.ErrEvmFailedToBroadcast{})
	suite.ErrorContains(res.Err, "insufficient fee")
}

func getMinFeeFromAppToml(zgChainHome string) (sdk.DecCoins, error) {
	// read the expected min gas price from app.toml
	parsed := struct {
		MinGasPrices string `toml:"minimum-gas-prices"`
	}{}
	appToml, err := os.ReadFile(filepath.Join(zgChainHome, "config", "app.toml"))
	if err != nil {
		return nil, err
	}
	err = toml.Unmarshal(appToml, &parsed)
	if err != nil {
		return nil, err
	}

	// convert to dec coins
	return sdk.ParseDecCoins(strings.ReplaceAll(parsed.MinGasPrices, ";", ","))
}
