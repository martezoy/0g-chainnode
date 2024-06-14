package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/0glabs/0g-chain/chaincfg"
	"github.com/0glabs/0g-chain/x/evmutil/types"
)

// ConversionMultiplier is the conversion multiplier between evm denom and gas denom
var ConversionMultiplier = sdkmath.NewInt(chaincfg.GasDenomConversionMultiplier)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal evm denom coin on the evm.
// x/evm consumes gas and send coins by minting and burning evm denom coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the gas denom coin and a separate evm denom balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	evmDenomKeeper Keeper
	bk             types.BankKeeper
	ak             types.AccountKeeper
}

func NewEvmBankKeeper(baseKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		evmDenomKeeper: baseKeeper,
		bk:             bk,
		ak:             ak,
	}
}

// GetBalance returns the total **spendable** balance of evm denom for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != chaincfg.EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", chaincfg.EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	gasDenomFromBank := spendableCoins.AmountOf(chaincfg.GasDenom)
	evmDenomFromBank := spendableCoins.AmountOf(chaincfg.EvmDenom)
	evmDenomFromEvmBank := k.evmDenomKeeper.GetBalance(ctx, addr)

	var total sdkmath.Int

	if gasDenomFromBank.IsPositive() {
		total = gasDenomFromBank.Mul(ConversionMultiplier).Add(evmDenomFromBank).Add(evmDenomFromEvmBank)
	} else {
		total = evmDenomFromBank.Add(evmDenomFromEvmBank)
	}
	return sdk.NewCoin(chaincfg.EvmDenom, total)
}

// SendCoins transfers evm denom coins from a AccAddress to an AccAddress.
func (k EvmBankKeeper) SendCoins(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// SendCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use SendCoins.
	panic("not implemented")
}

// SendCoinsFromModuleToAccount transfers evm denom coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	gasDenomCoin, baseDemonCnt, err := SplitEvmDenomCoins(amt)
	if err != nil {
		return err
	}

	if gasDenomCoin.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(gasDenomCoin)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneGasDenomToEvmDenomIfNeeded(ctx, senderAddr, baseDemonCnt); err != nil {
		return err
	}

	if err := k.evmDenomKeeper.SendBalance(ctx, senderAddr, recipientAddr, baseDemonCnt); err != nil {
		return err
	}

	return k.ConvertEvmDenomToGasDenom(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers evm denom coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	gasDenomCoin, evmDenomCnt, err := SplitEvmDenomCoins(amt)
	if err != nil {
		return err
	}

	if gasDenomCoin.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(gasDenomCoin)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneGasDenomToEvmDenomIfNeeded(ctx, senderAddr, evmDenomCnt); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.evmDenomKeeper.SendBalance(ctx, senderAddr, recipientAddr, evmDenomCnt); err != nil {
		return err
	}

	return k.ConvertEvmDenomToGasDenom(ctx, recipientAddr)
}

// MintCoins mints evm denom coins by minting the equivalent gas denom coins and any remaining evm denom coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	gasDenomCoin, baseDemonCnt, err := SplitEvmDenomCoins(amt)
	if err != nil {
		return err
	}

	if gasDenomCoin.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(gasDenomCoin)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.evmDenomKeeper.AddBalance(ctx, recipientAddr, baseDemonCnt); err != nil {
		return err
	}

	return k.ConvertEvmDenomToGasDenom(ctx, recipientAddr)
}

// BurnCoins burns evm denom coins by burning the equivalent gas denom coins and any remaining evm denom coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	gasDenomCoin, baseDemonCnt, err := SplitEvmDenomCoins(amt)
	if err != nil {
		return err
	}

	if gasDenomCoin.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(gasDenomCoin)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneGasDenomToEvmDenomIfNeeded(ctx, moduleAddr, baseDemonCnt); err != nil {
		return err
	}

	return k.evmDenomKeeper.RemoveBalance(ctx, moduleAddr, baseDemonCnt)
}

// ConvertOnegasDenomToEvmDenomIfNeeded converts 1 gas denom to evm denom for an address if
// its evm denom balance is smaller than the evmDenomCnt amount.
func (k EvmBankKeeper) ConvertOneGasDenomToEvmDenomIfNeeded(ctx sdk.Context, addr sdk.AccAddress, evmDenomCnt sdkmath.Int) error {
	evmDenomBal := k.evmDenomKeeper.GetBalance(ctx, addr)
	if evmDenomBal.GTE(evmDenomCnt) {
		return nil
	}

	gasDenomToStore := sdk.NewCoins(sdk.NewCoin(chaincfg.GasDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, gasDenomToStore); err != nil {
		return err
	}

	// add 1 gas denom equivalent of evm denom to addr
	evmDenomToReceive := ConversionMultiplier
	if err := k.evmDenomKeeper.AddBalance(ctx, addr, evmDenomToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertEvmDenomTogasDenom converts all available evm denom to gas denom for a given AccAddress.
func (k EvmBankKeeper) ConvertEvmDenomToGasDenom(ctx sdk.Context, addr sdk.AccAddress) error {
	totalEvmDenom := k.evmDenomKeeper.GetBalance(ctx, addr)
	gasDenomCoin, _, err := SplitEvmDenomCoins(sdk.NewCoins(sdk.NewCoin(chaincfg.EvmDenom, totalEvmDenom)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough evm denom for a single gas denom
	gasDenomToReceive := gasDenomCoin.Amount
	if !gasDenomToReceive.IsPositive() {
		return nil
	}

	// remove evm denom used for converting to gas denom
	evmDenomToBurn := gasDenomToReceive.Mul(ConversionMultiplier)
	finalBal := totalEvmDenom.Sub(evmDenomToBurn)
	if err := k.evmDenomKeeper.SetBalance(ctx, addr, finalBal); err != nil {
		return err
	}

	fromAddr := k.GetModuleAddress(types.ModuleName)
	if err := k.bk.SendCoins(ctx, fromAddr, addr, sdk.NewCoins(gasDenomCoin)); err != nil {
		return err
	}

	return nil
}

func (k EvmBankKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	addr := k.ak.GetModuleAddress(moduleName)
	if addr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}
	return addr
}

// SplitEvmDenomCoins splits evm denom coins to the equivalent gas denom coins and any remaining evm denom balance.
// An error will be returned if the coins are not valid or if the coins are not the evm denom.
func SplitEvmDenomCoins(coins sdk.Coins) (sdk.Coin, sdkmath.Int, error) {
	baseDemonCnt := sdk.ZeroInt()
	gasDenomAmt := sdk.NewCoin(chaincfg.GasDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return gasDenomAmt, baseDemonCnt, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return gasDenomAmt, baseDemonCnt, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		baseDemonCnt = remainingBalance
	}
	gasDenomAmount := coin.Amount.Quo(ConversionMultiplier)
	if gasDenomAmount.IsPositive() {
		gasDenomAmt = sdk.NewCoin(chaincfg.GasDenom, gasDenomAmount)
	}

	return gasDenomAmt, baseDemonCnt, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the evm denom.
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is evm denom
	if len(coins) != 1 || coins[0].Denom != chaincfg.EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", chaincfg.EvmDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
