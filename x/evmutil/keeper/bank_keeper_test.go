package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmtime "github.com/tendermint/tendermint/types/time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/0glabs/0g-chain/chaincfg"
	"github.com/0glabs/0g-chain/x/evmutil/keeper"
	"github.com/0glabs/0g-chain/x/evmutil/testutil"
	"github.com/0glabs/0g-chain/x/evmutil/types"
)

type evmBankKeeperTestSuite struct {
	testutil.Suite
}

func (suite *evmBankKeeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *evmBankKeeperTestSuite) TestGetBalance_ReturnsSpendable() {
	startingCoins := sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10))
	startingNeuron := sdkmath.NewInt(100)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	bacc := authtypes.NewBaseAccountWithAddress(suite.Addrs[0])
	vacc := vesting.NewContinuousVestingAccount(bacc, startingCoins, now.Unix(), endTime.Unix())
	suite.AccountKeeper.SetAccount(suite.Ctx, vacc)

	err := suite.App.FundAccount(suite.Ctx, suite.Addrs[0], startingCoins)
	suite.Require().NoError(err)
	err = suite.Keeper.SetBalance(suite.Ctx, suite.Addrs[0], startingNeuron)
	suite.Require().NoError(err)

	coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], chaincfg.BaseDenom)
	suite.Require().Equal(startingNeuron, coin.Amount)

	ctx := suite.Ctx.WithBlockTime(now.Add(12 * time.Hour))
	coin = suite.EvmBankKeeper.GetBalance(ctx, suite.Addrs[0], chaincfg.BaseDenom)
	suite.Require().Equal(sdkmath.NewIntFromUint64(5_000_000_000_100), coin.Amount)
}

func (suite *evmBankKeeperTestSuite) TestGetBalance_NotEvmDenom() {
	suite.Require().Panics(func() {
		suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], chaincfg.DisplayDenom)
	})
	suite.Require().Panics(func() {
		suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "busd")
	})
}

func (suite *evmBankKeeperTestSuite) TestGetBalance() {
	tests := []struct {
		name           string
		startingAmount sdk.Coins
		expAmount      sdkmath.Int
	}{
		{
			"a0gi with neuron",
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 100),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 10),
			),
			sdkmath.NewInt(10_000_000_000_100),
		},
		{
			"just neuron",
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 100),
				sdk.NewInt64Coin("busd", 100),
			),
			sdkmath.NewInt(100),
		},
		{
			"just a0gi",
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 10),
				sdk.NewInt64Coin("busd", 100),
			),
			sdkmath.NewInt(10_000_000_000_000),
		},
		{
			"no a0gi or neuron",
			sdk.NewCoins(),
			sdk.ZeroInt(),
		},
		{
			"with avaka that is more than 1 a0gi",
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 20_000_000_000_220),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 11),
			),
			sdkmath.NewInt(31_000_000_000_220),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingAmount)
			coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], chaincfg.BaseDenom)
			suite.Require().Equal(tt.expAmount, coin.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSendCoinsFromModuleToAccount() {
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin(chaincfg.BaseDenom, 200),
		sdk.NewInt64Coin(chaincfg.DisplayDenom, 100),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		startingAccBal sdk.Coins
		expAccBal      sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 12_000_000_000_010)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 10),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 12),
			),
			false,
		},
		{
			"send less than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 122)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 122),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 0),
			),
			false,
		},
		{
			"send an exact amount of a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 98_000_000_000_000)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 0o0),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 98),
			),
			false,
		},
		{
			"send no neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 0)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 0),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 0),
			),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total neuron to cover",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_001_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough a0gi to cover",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 200_000_000_000_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts receiver's neuron to a0gi if there's enough neuron after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 99_000_000_000_200)),
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 999_999_999_900),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 100),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 101),
			),
			false,
		},
		{
			"converts all of receiver's neuron to a0gi even if somehow receiver has more than 1a0gi of neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 12_000_000_000_100)),
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 5_999_999_999_990),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 90),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 19),
			),
			false,
		},
		{
			"swap 1 a0gi for neuron if module account doesn't have enough neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 99_000_000_001_000)),
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 200),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin(chaincfg.BaseDenom, 1200),
				sdk.NewInt64Coin(chaincfg.DisplayDenom, 100),
			),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingAccBal)
			suite.FundModuleAccountWithZgChain(evmtypes.ModuleName, startingModuleCoins)

			// fund our module with some a0gi to account for converting extra neuron back to a0gi
			suite.FundModuleAccountWithZgChain(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10)))

			err := suite.EvmBankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, evmtypes.ModuleName, suite.Addrs[0], tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check a0gi
			a0giSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], chaincfg.DisplayDenom)
			suite.Require().Equal(tt.expAccBal.AmountOf(chaincfg.DisplayDenom).Int64(), a0giSender.Amount.Int64())

			// check neuron
			actualNeuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expAccBal.AmountOf(chaincfg.BaseDenom).Int64(), actualNeuron.Int64())
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSendCoinsFromAccountToModule() {
	startingAccCoins := sdk.NewCoins(
		sdk.NewInt64Coin(chaincfg.BaseDenom, 200),
		sdk.NewInt64Coin(chaincfg.DisplayDenom, 100),
	)
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_000),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		expSenderCoins sdk.Coins
		expModuleCoins sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 12_000_000_000_010)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 190), sdk.NewInt64Coin(chaincfg.DisplayDenom, 88)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_010), sdk.NewInt64Coin(chaincfg.DisplayDenom, 12)),
			false,
		},
		{
			"send less than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 122)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 78), sdk.NewInt64Coin(chaincfg.DisplayDenom, 100)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_122), sdk.NewInt64Coin(chaincfg.DisplayDenom, 0)),
			false,
		},
		{
			"send an exact amount of a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 98_000_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 200), sdk.NewInt64Coin(chaincfg.DisplayDenom, 2)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_000), sdk.NewInt64Coin(chaincfg.DisplayDenom, 98)),
			false,
		},
		{
			"send no neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 0)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 200), sdk.NewInt64Coin(chaincfg.DisplayDenom, 100)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_000), sdk.NewInt64Coin(chaincfg.DisplayDenom, 0)),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin(chaincfg.BaseDenom, 12_000_000_000_000),
				sdk.NewInt64Coin(chaincfg.BaseDenom, 2_000_000_000_000),
			},
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total neuron to cover",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_001_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough a0gi to cover",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 200_000_000_000_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts 1 a0gi to neuron if not enough neuron to cover",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 99_001_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 999_000_000_200), sdk.NewInt64Coin(chaincfg.DisplayDenom, 0)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 101_000_000_000), sdk.NewInt64Coin(chaincfg.DisplayDenom, 99)),
			false,
		},
		{
			"converts receiver's neuron to a0gi if there's enough neuron after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 5_900_000_000_200)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_000_000_000), sdk.NewInt64Coin(chaincfg.DisplayDenom, 94)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 200), sdk.NewInt64Coin(chaincfg.DisplayDenom, 6)),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			suite.FundAccountWithZgChain(suite.Addrs[0], startingAccCoins)
			suite.FundModuleAccountWithZgChain(evmtypes.ModuleName, startingModuleCoins)

			err := suite.EvmBankKeeper.SendCoinsFromAccountToModule(suite.Ctx, suite.Addrs[0], evmtypes.ModuleName, tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check sender balance
			a0giSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], chaincfg.DisplayDenom)
			suite.Require().Equal(tt.expSenderCoins.AmountOf(chaincfg.DisplayDenom).Int64(), a0giSender.Amount.Int64())
			actualNeuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expSenderCoins.AmountOf(chaincfg.BaseDenom).Int64(), actualNeuron.Int64())

			// check module balance
			moduleAddr := suite.AccountKeeper.GetModuleAddress(evmtypes.ModuleName)
			a0giSender = suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, chaincfg.DisplayDenom)
			suite.Require().Equal(tt.expModuleCoins.AmountOf(chaincfg.DisplayDenom).Int64(), a0giSender.Amount.Int64())
			actualNeuron = suite.Keeper.GetBalance(suite.Ctx, moduleAddr)
			suite.Require().Equal(tt.expModuleCoins.AmountOf(chaincfg.BaseDenom).Int64(), actualNeuron.Int64())
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestBurnCoins() {
	startingA0gi := sdkmath.NewInt(100)
	tests := []struct {
		name       string
		burnCoins  sdk.Coins
		expA0gi     sdkmath.Int
		expNeuron   sdkmath.Int
		hasErr     bool
		neuronStart sdkmath.Int
	}{
		{
			"burn more than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 12_021_000_000_002)),
			sdkmath.NewInt(88),
			sdkmath.NewInt(100_000_000_000),
			false,
			sdkmath.NewInt(121_000_000_002),
		},
		{
			"burn less than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 122)),
			sdkmath.NewInt(100),
			sdkmath.NewInt(878),
			false,
			sdkmath.NewInt(1000),
		},
		{
			"burn an exact amount of a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 98_000_000_000_000)),
			sdkmath.NewInt(2),
			sdkmath.NewInt(10),
			false,
			sdkmath.NewInt(10),
		},
		{
			"burn no neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 0)),
			startingA0gi,
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if burning other coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 500), sdk.NewInt64Coin("busd", 1000)),
			startingA0gi,
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin(chaincfg.BaseDenom, 12_000_000_000_000),
				sdk.NewInt64Coin(chaincfg.BaseDenom, 2_000_000_000_000),
			},
			startingA0gi,
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if burn amount is negative",
			sdk.Coins{sdk.Coin{Denom: chaincfg.BaseDenom, Amount: sdkmath.NewInt(-100)}},
			startingA0gi,
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"errors if not enough neuron to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100_999_000_000_000)),
			sdkmath.NewInt(0),
			sdkmath.NewInt(99_000_000_000),
			true,
			sdkmath.NewInt(99_000_000_000),
		},
		{
			"errors if not enough a0gi to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 200_000_000_000_000)),
			sdkmath.NewInt(100),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"converts 1 a0gi to neuron if not enough neuron to cover",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 12_021_000_000_002)),
			sdkmath.NewInt(87),
			sdkmath.NewInt(980_000_000_000),
			false,
			sdkmath.NewInt(1_000_000_002),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			startingCoins := sdk.NewCoins(
				sdk.NewCoin(chaincfg.DisplayDenom, startingA0gi),
				sdk.NewCoin(chaincfg.BaseDenom, tt.neuronStart),
			)
			suite.FundModuleAccountWithZgChain(evmtypes.ModuleName, startingCoins)

			err := suite.EvmBankKeeper.BurnCoins(suite.Ctx, evmtypes.ModuleName, tt.burnCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check a0gi
			a0giActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, chaincfg.DisplayDenom)
			suite.Require().Equal(tt.expA0gi, a0giActual.Amount)

			// check neuron
			neuronActual := suite.Keeper.GetBalance(suite.Ctx, suite.EvmModuleAddr)
			suite.Require().Equal(tt.expNeuron, neuronActual)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestMintCoins() {
	tests := []struct {
		name       string
		mintCoins  sdk.Coins
		a0gi        sdkmath.Int
		neuron      sdkmath.Int
		hasErr     bool
		neuronStart sdkmath.Int
	}{
		{
			"mint more than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 12_021_000_000_002)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_002),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint less than 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 901_000_000_001)),
			sdk.ZeroInt(),
			sdkmath.NewInt(901_000_000_001),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint an exact amount of a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 123_000_000_000_000_000)),
			sdkmath.NewInt(123_000),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint no neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 0)),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if minting other coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.ZeroInt(),
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin(chaincfg.BaseDenom, 12_000_000_000_000),
				sdk.NewInt64Coin(chaincfg.BaseDenom, 2_000_000_000_000),
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if mint amount is negative",
			sdk.Coins{sdk.Coin{Denom: chaincfg.BaseDenom, Amount: sdkmath.NewInt(-100)}},
			sdk.ZeroInt(),
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"adds to existing neuron balance",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 12_021_000_000_002)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_102),
			false,
			sdkmath.NewInt(100),
		},
		{
			"convert neuron balance to a0gi if it exceeds 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 10_999_000_000_000)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(1_200_000_001),
			false,
			sdkmath.NewInt(1_002_200_000_001),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			suite.FundModuleAccountWithZgChain(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10)))
			suite.FundModuleAccountWithZgChain(evmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(chaincfg.BaseDenom, tt.neuronStart)))

			err := suite.EvmBankKeeper.MintCoins(suite.Ctx, evmtypes.ModuleName, tt.mintCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check a0gi
			a0giActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, chaincfg.DisplayDenom)
			suite.Require().Equal(tt.a0gi, a0giActual.Amount)

			// check neuron
			neuronActual := suite.Keeper.GetBalance(suite.Ctx, suite.EvmModuleAddr)
			suite.Require().Equal(tt.neuron, neuronActual)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestValidateEvmCoins() {
	tests := []struct {
		name      string
		coins     sdk.Coins
		shouldErr bool
	}{
		{
			"valid coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 500)),
			false,
		},
		{
			"dup coins",
			sdk.Coins{sdk.NewInt64Coin(chaincfg.BaseDenom, 500), sdk.NewInt64Coin(chaincfg.BaseDenom, 500)},
			true,
		},
		{
			"not evm coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 500)),
			true,
		},
		{
			"negative coins",
			sdk.Coins{sdk.Coin{Denom: chaincfg.BaseDenom, Amount: sdkmath.NewInt(-500)}},
			true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := keeper.ValidateEvmCoins(tt.coins)
			if tt.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestConvertOneA0giToNeuronIfNeeded() {
	neuronNeeded := sdkmath.NewInt(200)
	tests := []struct {
		name          string
		startingCoins sdk.Coins
		expectedCoins sdk.Coins
		success       bool
	}{
		{
			"not enough a0gi for conversion",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100)),
			false,
		},
		{
			"converts 1 a0gi to neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10), sdk.NewInt64Coin(chaincfg.BaseDenom, 100)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 9), sdk.NewInt64Coin(chaincfg.BaseDenom, 1_000_000_000_100)),
			true,
		},
		{
			"conversion not needed",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10), sdk.NewInt64Coin(chaincfg.BaseDenom, 200)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10), sdk.NewInt64Coin(chaincfg.BaseDenom, 200)),
			true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingCoins)
			err := suite.EvmBankKeeper.ConvertOneA0giToNeuronIfNeeded(suite.Ctx, suite.Addrs[0], neuronNeeded)
			moduleZgChain := suite.BankKeeper.GetBalance(suite.Ctx, suite.AccountKeeper.GetModuleAddress(types.ModuleName), chaincfg.DisplayDenom)
			if tt.success {
				suite.Require().NoError(err)
				if tt.startingCoins.AmountOf(chaincfg.BaseDenom).LT(neuronNeeded) {
					suite.Require().Equal(sdk.OneInt(), moduleZgChain.Amount)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().Equal(sdk.ZeroInt(), moduleZgChain.Amount)
			}

			neuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf(chaincfg.BaseDenom), neuron)
			a0gi := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], chaincfg.DisplayDenom)
			suite.Require().Equal(tt.expectedCoins.AmountOf(chaincfg.DisplayDenom), a0gi.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestConvertNeuronToA0gi() {
	tests := []struct {
		name          string
		startingCoins sdk.Coins
		expectedCoins sdk.Coins
	}{
		{
			"not enough a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 100), sdk.NewInt64Coin(chaincfg.DisplayDenom, 0)),
		},
		{
			"converts neuron for 1 a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10), sdk.NewInt64Coin(chaincfg.BaseDenom, 1_000_000_000_003)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 11), sdk.NewInt64Coin(chaincfg.BaseDenom, 3)),
		},
		{
			"converts more than 1 a0gi of neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10), sdk.NewInt64Coin(chaincfg.BaseDenom, 8_000_000_000_123)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 18), sdk.NewInt64Coin(chaincfg.BaseDenom, 123)),
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 10)))
			suite.Require().NoError(err)
			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingCoins)
			err = suite.EvmBankKeeper.ConvertNeuronToA0gi(suite.Ctx, suite.Addrs[0])
			suite.Require().NoError(err)
			neuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf(chaincfg.BaseDenom), neuron)
			a0gi := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], chaincfg.DisplayDenom)
			suite.Require().Equal(tt.expectedCoins.AmountOf(chaincfg.DisplayDenom), a0gi.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSplitNeuronCoins() {
	tests := []struct {
		name          string
		coins         sdk.Coins
		expectedCoins sdk.Coins
		shouldErr     bool
	}{
		{
			"invalid coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 500)),
			nil,
			true,
		},
		{
			"empty coins",
			sdk.NewCoins(),
			sdk.NewCoins(),
			false,
		},
		{
			"a0gi & neuron coins",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 8_000_000_000_123)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 8), sdk.NewInt64Coin(chaincfg.BaseDenom, 123)),
			false,
		},
		{
			"only neuron",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 10_123)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 10_123)),
			false,
		},
		{
			"only a0gi",
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.BaseDenom, 5_000_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin(chaincfg.DisplayDenom, 5)),
			false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			a0gi, neuron, err := keeper.SplitNeuronCoins(tt.coins)
			if tt.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.expectedCoins.AmountOf(chaincfg.DisplayDenom), a0gi.Amount)
				suite.Require().Equal(tt.expectedCoins.AmountOf(chaincfg.BaseDenom), neuron)
			}
		})
	}
}

func TestEvmBankKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(evmBankKeeperTestSuite))
}
