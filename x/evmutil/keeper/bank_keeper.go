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

// ConversionMultiplier is the conversion multiplier between base denom and auxiliary denom
var ConversionMultiplier = sdkmath.NewInt(chaincfg.AuxiliaryDenomConversionMultiplier)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal base denom coin on the evm.
// x/evm consumes gas and send coins by minting and burning base denom coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the auxiliary denom coin and a separate base denom balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	baseDenomKeeper Keeper
	bk              types.BankKeeper
	ak              types.AccountKeeper
}

func NewEvmBankKeeper(baseKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		baseDenomKeeper: baseKeeper,
		bk:              bk,
		ak:              ak,
	}
}

// GetBalance returns the total **spendable** balance of base denom for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != chaincfg.BaseDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", chaincfg.BaseDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	auxiliaryDenomFromBank := spendableCoins.AmountOf(chaincfg.AuxiliaryDenom)
	baseDenomFromBank := spendableCoins.AmountOf(chaincfg.BaseDenom)
	baseDenomFromEvmBank := k.baseDenomKeeper.GetBalance(ctx, addr)

	var total sdkmath.Int

	if auxiliaryDenomFromBank.IsPositive() {
		total = auxiliaryDenomFromBank.Mul(ConversionMultiplier).Add(baseDenomFromBank).Add(baseDenomFromEvmBank)
	} else {
		total = baseDenomFromBank.Add(baseDenomFromEvmBank)
	}
	return sdk.NewCoin(chaincfg.BaseDenom, total)
}

// SendCoins transfers base denom coins from a AccAddress to an AccAddress.
func (k EvmBankKeeper) SendCoins(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// SendCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use SendCoins.
	panic("not implemented")
}

// SendCoinsFromModuleToAccount transfers base denom coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	auxiliaryDenomCoin, baseDemonCnt, err := SplitBaseDenomCoins(amt)
	if err != nil {
		return err
	}

	if auxiliaryDenomCoin.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(auxiliaryDenomCoin)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneAuxiliaryDenomToBaseDenomIfNeeded(ctx, senderAddr, baseDemonCnt); err != nil {
		return err
	}

	if err := k.baseDenomKeeper.SendBalance(ctx, senderAddr, recipientAddr, baseDemonCnt); err != nil {
		return err
	}

	return k.ConvertBaseDenomToAuxiliaryDenom(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers base denom coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	auxiliaryDenomCoin, baseDenomCnt, err := SplitBaseDenomCoins(amt)
	if err != nil {
		return err
	}

	if auxiliaryDenomCoin.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(auxiliaryDenomCoin)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneAuxiliaryDenomToBaseDenomIfNeeded(ctx, senderAddr, baseDenomCnt); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.baseDenomKeeper.SendBalance(ctx, senderAddr, recipientAddr, baseDenomCnt); err != nil {
		return err
	}

	return k.ConvertBaseDenomToAuxiliaryDenom(ctx, recipientAddr)
}

// MintCoins mints base denom coins by minting the equivalent auxiliary denom coins and any remaining base denom coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	auxiliaryDenomCoin, baseDemonCnt, err := SplitBaseDenomCoins(amt)
	if err != nil {
		return err
	}

	if auxiliaryDenomCoin.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(auxiliaryDenomCoin)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.baseDenomKeeper.AddBalance(ctx, recipientAddr, baseDemonCnt); err != nil {
		return err
	}

	return k.ConvertBaseDenomToAuxiliaryDenom(ctx, recipientAddr)
}

// BurnCoins burns base denom coins by burning the equivalent auxiliary denom coins and any remaining base denom coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	auxiliaryDenomCoin, baseDemonCnt, err := SplitBaseDenomCoins(amt)
	if err != nil {
		return err
	}

	if auxiliaryDenomCoin.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(auxiliaryDenomCoin)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneAuxiliaryDenomToBaseDenomIfNeeded(ctx, moduleAddr, baseDemonCnt); err != nil {
		return err
	}

	return k.baseDenomKeeper.RemoveBalance(ctx, moduleAddr, baseDemonCnt)
}

// ConvertOneauxiliaryDenomToBaseDenomIfNeeded converts 1 auxiliary denom to base denom for an address if
// its base denom balance is smaller than the baseDenomCnt amount.
func (k EvmBankKeeper) ConvertOneAuxiliaryDenomToBaseDenomIfNeeded(ctx sdk.Context, addr sdk.AccAddress, baseDenomCnt sdkmath.Int) error {
	baseDenomBal := k.baseDenomKeeper.GetBalance(ctx, addr)
	if baseDenomBal.GTE(baseDenomCnt) {
		return nil
	}

	auxiliaryDenomToStore := sdk.NewCoins(sdk.NewCoin(chaincfg.AuxiliaryDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, auxiliaryDenomToStore); err != nil {
		return err
	}

	// add 1 auxiliary denom equivalent of base denom to addr
	baseDenomToReceive := ConversionMultiplier
	if err := k.baseDenomKeeper.AddBalance(ctx, addr, baseDenomToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertBaseDenomToauxiliaryDenom converts all available base denom to auxiliary denom for a given AccAddress.
func (k EvmBankKeeper) ConvertBaseDenomToAuxiliaryDenom(ctx sdk.Context, addr sdk.AccAddress) error {
	totalBaseDenom := k.baseDenomKeeper.GetBalance(ctx, addr)
	auxiliaryDenomCoin, _, err := SplitBaseDenomCoins(sdk.NewCoins(sdk.NewCoin(chaincfg.BaseDenom, totalBaseDenom)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough base denom for a single auxiliary denom
	auxiliaryDenomToReceive := auxiliaryDenomCoin.Amount
	if !auxiliaryDenomToReceive.IsPositive() {
		return nil
	}

	// remove base denom used for converting to auxiliary denom
	baseDenomToBurn := auxiliaryDenomToReceive.Mul(ConversionMultiplier)
	finalBal := totalBaseDenom.Sub(baseDenomToBurn)
	if err := k.baseDenomKeeper.SetBalance(ctx, addr, finalBal); err != nil {
		return err
	}

	fromAddr := k.GetModuleAddress(types.ModuleName)
	if err := k.bk.SendCoins(ctx, fromAddr, addr, sdk.NewCoins(auxiliaryDenomCoin)); err != nil {
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

// SplitBaseDenomCoins splits base denom coins to the equivalent auxiliary denom coins and any remaining base denom balance.
// An error will be returned if the coins are not valid or if the coins are not the base denom.
func SplitBaseDenomCoins(coins sdk.Coins) (sdk.Coin, sdkmath.Int, error) {
	baseDemonCnt := sdk.ZeroInt()
	auxiliaryDenomAmt := sdk.NewCoin(chaincfg.AuxiliaryDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return auxiliaryDenomAmt, baseDemonCnt, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return auxiliaryDenomAmt, baseDemonCnt, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		baseDemonCnt = remainingBalance
	}
	auxiliaryDenomAmount := coin.Amount.Quo(ConversionMultiplier)
	if auxiliaryDenomAmount.IsPositive() {
		auxiliaryDenomAmt = sdk.NewCoin(chaincfg.AuxiliaryDenom, auxiliaryDenomAmount)
	}

	return auxiliaryDenomAmt, baseDemonCnt, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the base denom.
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is base denom
	if len(coins) != 1 || coins[0].Denom != chaincfg.BaseDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", chaincfg.BaseDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
