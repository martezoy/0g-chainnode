package keeper_test

import (
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/0glabs/0g-chain/x/evmutil/keeper"
	"github.com/0glabs/0g-chain/x/evmutil/testutil"
	"github.com/0glabs/0g-chain/x/evmutil/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/suite"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type evmBankKeeperTestSuite struct {
	testutil.Suite
}

func (suite *evmBankKeeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *evmBankKeeperTestSuite) TestGetBalance_ReturnsSpendable() {
	startingCoins := sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10))
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

	coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "neuron")
	suite.Require().Equal(startingNeuron, coin.Amount)

	ctx := suite.Ctx.WithBlockTime(now.Add(12 * time.Hour))
	coin = suite.EvmBankKeeper.GetBalance(ctx, suite.Addrs[0], "neuron")
	suite.Require().Equal(sdkmath.NewIntFromUint64(5_000_000_000_100), coin.Amount)
}
func (suite *evmBankKeeperTestSuite) TestGetBalance_NotEvmDenom() {
	suite.Require().Panics(func() {
		suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ua0gi")
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
			"ua0gi with neuron",
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 100),
				sdk.NewInt64Coin("ua0gi", 10),
			),
			sdk.NewIntFromBigInt(makeBigIntByString("10000000000100")),
		},
		{
			"just neuron",
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 100),
				sdk.NewInt64Coin("busd", 100),
			),
			sdkmath.NewInt(100),
		},
		{
			"just ua0gi",
			sdk.NewCoins(
				sdk.NewInt64Coin("ua0gi", 10),
				sdk.NewInt64Coin("busd", 100),
			),
			sdk.NewIntFromBigInt(makeBigIntByString("10000000000000")),
		},
		{
			"no ua0gi or neuron",
			sdk.NewCoins(),
			sdk.ZeroInt(),
		},
		{
			"with avaka that is more than 1 ua0gi",
			sdk.NewCoins(
				sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("20000000000220"))),
				sdk.NewInt64Coin("ua0gi", 11),
			),
			sdk.NewIntFromBigInt(makeBigIntByString("31000000000220")),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingAmount)
			coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "neuron")
			suite.Require().Equal(tt.expAmount, coin.Amount)
		})
	}
}
func (suite *evmBankKeeperTestSuite) TestSendCoinsFromModuleToAccount() {
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin("neuron", 200),
		sdk.NewInt64Coin("ua0gi", 100),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		startingAccBal sdk.Coins
		expAccBal      sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("12000000000010")))),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 10),
				sdk.NewInt64Coin("ua0gi", 12),
			),
			false,
		},
		{
			"send less than 1 ua0gi",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 122)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 122),
				sdk.NewInt64Coin("ua0gi", 0),
			),
			false,
		},
		{
			"send an exact amount of ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("98000000000000")))),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 0),
				sdk.NewInt64Coin("ua0gi", 98),
			),
			false,
		},
		{
			"send no neuron",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 0)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 0),
				sdk.NewInt64Coin("ua0gi", 0),
			),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total neuron to cover",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("100000000001000")))),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough ua0gi to cover",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("200000000000000")))),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts receiver's neuron to ua0gi if there's enough neuron after the transfer",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("99000000000200")))),
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 999_999_999_900),
				sdk.NewInt64Coin("ua0gi", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 100),
				sdk.NewInt64Coin("ua0gi", 101),
			),
			false,
		},
		{
			"converts all of receiver's neuron to ua0gi even if somehow receiver has more than 1a0gi of neuron",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("12000000000100")))),
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 5_999_999_999_990),
				sdk.NewInt64Coin("ua0gi", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 90),
				sdk.NewInt64Coin("ua0gi", 19),
			),
			false,
		},
		{
			"swap 1 ua0gi for neuron if module account doesn't have enough neuron",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("99000000001000")))),
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 200),
				sdk.NewInt64Coin("ua0gi", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("neuron", 1200),
				sdk.NewInt64Coin("ua0gi", 100),
			),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingAccBal)
			suite.FundModuleAccountWithZgChain(evmtypes.ModuleName, startingModuleCoins)

			// fund our module with some ua0gi to account for converting extra neuron back to ua0gi
			suite.FundModuleAccountWithZgChain(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10)))

			err := suite.EvmBankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, evmtypes.ModuleName, suite.Addrs[0], tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ua0gi
			a0giSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ua0gi")
			suite.Require().Equal(tt.expAccBal.AmountOf("ua0gi").Int64(), a0giSender.Amount.Int64())

			// check neuron
			actualNeuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expAccBal.AmountOf("neuron").Int64(), actualNeuron.Int64())
		})
	}
}
func (suite *evmBankKeeperTestSuite) TestSendCoinsFromAccountToModule() {
	startingAccCoins := sdk.NewCoins(
		sdk.NewInt64Coin("neuron", 200),
		sdk.NewInt64Coin("ua0gi", 100),
	)
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin("neuron", 100_000_000_000),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		expSenderCoins sdk.Coins
		expModuleCoins sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("12000000000010")))),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 190), sdk.NewInt64Coin("ua0gi", 88)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100_000_000_010), sdk.NewInt64Coin("ua0gi", 12)),
			false,
		},
		{
			"send less than 1 ua0gi",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 122)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 78), sdk.NewInt64Coin("ua0gi", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100_000_000_122), sdk.NewInt64Coin("ua0gi", 0)),
			false,
		},
		{
			"send an exact amount of ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("98000000000000")))),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 200), sdk.NewInt64Coin("ua0gi", 2)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100_000_000_000), sdk.NewInt64Coin("ua0gi", 98)),
			false,
		},
		{
			"send no neuron",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 0)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 200), sdk.NewInt64Coin("ua0gi", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100_000_000_000), sdk.NewInt64Coin("ua0gi", 0)),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("neuron", 12_000_000_000_000),
				sdk.NewInt64Coin("neuron", 2_000_000_000_000),
			},
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total neuron to cover",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("100000000001000")))),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough ua0gi to cover",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("200000000000000")))),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts 1 ua0gi to neuron if not enough neuron to cover",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("99001000000000")))),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 999_000_000_200), sdk.NewInt64Coin("ua0gi", 0)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 101_000_000_000), sdk.NewInt64Coin("ua0gi", 99)),
			false,
		},
		{
			"converts receiver's neuron to ua0gi if there's enough neuron after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 5_900_000_000_200)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100_000_000_000), sdk.NewInt64Coin("ua0gi", 94)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 200), sdk.NewInt64Coin("ua0gi", 6)),
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
			a0giSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ua0gi")
			suite.Require().Equal(tt.expSenderCoins.AmountOf("ua0gi").Int64(), a0giSender.Amount.Int64())
			actualNeuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expSenderCoins.AmountOf("neuron").Int64(), actualNeuron.Int64())

			// check module balance
			moduleAddr := suite.AccountKeeper.GetModuleAddress(evmtypes.ModuleName)
			a0giSender = suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "ua0gi")
			suite.Require().Equal(tt.expModuleCoins.AmountOf("ua0gi").Int64(), a0giSender.Amount.Int64())
			actualNeuron = suite.Keeper.GetBalance(suite.Ctx, moduleAddr)
			suite.Require().Equal(tt.expModuleCoins.AmountOf("neuron").Int64(), actualNeuron.Int64())
		})
	}
}
func (suite *evmBankKeeperTestSuite) TestBurnCoins() {
	startingA0gi := sdkmath.NewInt(100)
	tests := []struct {
		name        string
		burnCoins   sdk.Coins
		expA0gi     sdkmath.Int
		expNeuron   sdkmath.Int
		hasErr      bool
		neuronStart sdkmath.Int
	}{
		{
			"burn more than 1 ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("12021000000002")))),
			sdkmath.NewInt(88),
			sdkmath.NewInt(100_000_000_000),
			false,
			sdk.NewIntFromBigInt(makeBigIntByString("121000000002")),
		},
		{
			"burn less than 1 ua0gi",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 122)),
			sdkmath.NewInt(100),
			sdkmath.NewInt(878),
			false,
			sdkmath.NewInt(1000),
		},
		{
			"burn an exact amount of ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("98000000000000")))),
			sdkmath.NewInt(2),
			sdkmath.NewInt(10),
			false,
			sdkmath.NewInt(10),
		},
		{
			"burn no neuron",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 0)),
			startingA0gi,
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if burning other coins",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 500), sdk.NewInt64Coin("busd", 1000)),
			startingA0gi,
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("neuron", 12_000_000_000_000),
				sdk.NewInt64Coin("neuron", 2_000_000_000_000),
			},
			startingA0gi,
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if burn amount is negative",
			sdk.Coins{sdk.Coin{Denom: "neuron", Amount: sdkmath.NewInt(-100)}},
			startingA0gi,
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"errors if not enough neuron to cover burn",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("100999000000000")))),
			sdkmath.NewInt(0),
			sdkmath.NewInt(99_000_000_000),
			true,
			sdkmath.NewInt(99_000_000_000),
		},
		{
			"errors if not enough ua0gi to cover burn",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("200000000000000")))),
			sdkmath.NewInt(100),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"converts 1 ua0gi to neuron if not enough neuron to cover",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("12021000000002")))),
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
				sdk.NewCoin("ua0gi", startingA0gi),
				sdk.NewCoin("neuron", tt.neuronStart),
			)
			suite.FundModuleAccountWithZgChain(evmtypes.ModuleName, startingCoins)

			err := suite.EvmBankKeeper.BurnCoins(suite.Ctx, evmtypes.ModuleName, tt.burnCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ua0gi
			a0giActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, "ua0gi")
			suite.Require().Equal(tt.expA0gi, a0giActual.Amount)

			// check neuron
			neuronActual := suite.Keeper.GetBalance(suite.Ctx, suite.EvmModuleAddr)
			suite.Require().Equal(tt.expNeuron, neuronActual)
		})
	}
}
func (suite *evmBankKeeperTestSuite) TestMintCoins() {
	tests := []struct {
		name        string
		mintCoins   sdk.Coins
		ua0gi       sdkmath.Int
		neuron      sdkmath.Int
		hasErr      bool
		neuronStart sdkmath.Int
	}{
		{
			"mint more than 1 ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("12021000000002")))),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_002),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint less than 1 ua0gi",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 901_000_000_001)),
			sdk.ZeroInt(),
			sdkmath.NewInt(901_000_000_001),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint an exact amount of ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("123000000000000000")))),
			sdkmath.NewInt(123_000),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint no neuron",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 0)),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if minting other coins",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.ZeroInt(),
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("neuron", 12_000_000_000_000),
				sdk.NewInt64Coin("neuron", 2_000_000_000_000),
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if mint amount is negative",
			sdk.Coins{sdk.Coin{Denom: "neuron", Amount: sdkmath.NewInt(-100)}},
			sdk.ZeroInt(),
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"adds to existing neuron balance",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("12021000000002")))),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_102),
			false,
			sdkmath.NewInt(100),
		},
		{
			"convert neuron balance to ua0gi if it exceeds 1 ua0gi",
			sdk.NewCoins(sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("10999000000000")))),
			sdkmath.NewInt(12),
			sdkmath.NewInt(1_200_000_001),
			false,
			sdkmath.NewIntFromBigInt(makeBigIntByString("1002200000001")),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			suite.FundModuleAccountWithZgChain(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10)))
			suite.FundModuleAccountWithZgChain(evmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("neuron", tt.neuronStart)))

			err := suite.EvmBankKeeper.MintCoins(suite.Ctx, evmtypes.ModuleName, tt.mintCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ua0gi
			a0giActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, "ua0gi")
			suite.Require().Equal(tt.ua0gi, a0giActual.Amount)

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
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 500)),
			false,
		},
		{
			"dup coins",
			sdk.Coins{sdk.NewInt64Coin("neuron", 500), sdk.NewInt64Coin("neuron", 500)},
			true,
		},
		{
			"not evm coins",
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 500)),
			true,
		},
		{
			"negative coins",
			sdk.Coins{sdk.Coin{Denom: "neuron", Amount: sdkmath.NewInt(-500)}},
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
			"not enough ua0gi for conversion",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100)),
			false,
		},
		{
			"converts 1 ua0gi to neuron",
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10), sdk.NewInt64Coin("neuron", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 9), sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("1000000000100")))),
			true,
		},
		{
			"conversion not needed",
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10), sdk.NewInt64Coin("neuron", 200)),
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10), sdk.NewInt64Coin("neuron", 200)),
			true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingCoins)
			err := suite.EvmBankKeeper.ConvertOneUa0giToNeuronIfNeeded(suite.Ctx, suite.Addrs[0], neuronNeeded)
			moduleZgChain := suite.BankKeeper.GetBalance(suite.Ctx, suite.AccountKeeper.GetModuleAddress(types.ModuleName), "ua0gi")
			if tt.success {
				suite.Require().NoError(err)
				if tt.startingCoins.AmountOf("neuron").LT(neuronNeeded) {
					suite.Require().Equal(sdk.OneInt(), moduleZgChain.Amount)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().Equal(sdk.ZeroInt(), moduleZgChain.Amount)
			}

			neuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf("neuron"), neuron)
			ua0gi := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ua0gi")
			suite.Require().Equal(tt.expectedCoins.AmountOf("ua0gi"), ua0gi.Amount)
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
			"not enough ua0gi",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 100), sdk.NewInt64Coin("ua0gi", 0)),
		},
		{
			"converts neuron for 1 ua0gi",
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10), sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("1000000000003")))),
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 11), sdk.NewInt64Coin("neuron", 3)),
		},
		{
			"converts more than 1 ua0gi of neuron",
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10), sdk.NewCoin("neuron", sdk.NewIntFromBigInt(makeBigIntByString("8000000000123")))),
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 18), sdk.NewInt64Coin("neuron", 123)),
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 10)))
			suite.Require().NoError(err)
			suite.FundAccountWithZgChain(suite.Addrs[0], tt.startingCoins)
			err = suite.EvmBankKeeper.ConvertNeuronToUa0gi(suite.Ctx, suite.Addrs[0])
			suite.Require().NoError(err)
			neuron := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf("neuron"), neuron)
			ua0gi := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ua0gi")
			suite.Require().Equal(tt.expectedCoins.AmountOf("ua0gi"), ua0gi.Amount)
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
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 500)),
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
			"ua0gi & neuron coins",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 8_000_000_000_123)),
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 8), sdk.NewInt64Coin("neuron", 123)),
			false,
		},
		{
			"only neuron",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 10_123)),
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 10_123)),
			false,
		},
		{
			"only ua0gi",
			sdk.NewCoins(sdk.NewInt64Coin("neuron", 5_000_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin("ua0gi", 5)),
			false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ua0gi, neuron, err := keeper.SplitNeuronCoins(tt.coins)
			if tt.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.expectedCoins.AmountOf("ua0gi"), ua0gi.Amount)
				suite.Require().Equal(tt.expectedCoins.AmountOf("neuron"), neuron)
			}
		})
	}
}

func TestEvmBankKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(evmBankKeeperTestSuite))
}

func makeBigIntByString(s string) *big.Int {
	i, _ := new(big.Int).SetString(s, 10)
	return i
}
