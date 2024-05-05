package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/0glabs/0g-chain/x/evmutil/types"
)

const (
	// EvmDenom is the gas denom used by the evm
	EvmDenom = "neuron"

	// CosmosDenom is the gas denom used by the 0g-chain app
	CosmosDenom = "ua0gi"
)

// ConversionMultiplier is the conversion multiplier between neuron and ua0gi
var ConversionMultiplier = sdkmath.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal neuron coin on the evm.
// x/evm consumes gas and send coins by minting and burning neuron coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the a0gi coin and a separate neuron balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	neuronKeeper Keeper
	bk           types.BankKeeper
	ak           types.AccountKeeper
}

func NewEvmBankKeeper(neuronKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		neuronKeeper: neuronKeeper,
		bk:           bk,
		ak:           ak,
	}
}

// GetBalance returns the total **spendable** balance of neuron for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	cosmosDenomFromBank := spendableCoins.AmountOf(CosmosDenom)
	evmDenomFromBank := spendableCoins.AmountOf(EvmDenom)
	evmDenomFromEvmBank := k.neuronKeeper.GetBalance(ctx, addr)

	var total sdkmath.Int

	if cosmosDenomFromBank.IsPositive() {
		total = cosmosDenomFromBank.Mul(ConversionMultiplier).Add(evmDenomFromBank).Add(evmDenomFromEvmBank)
	} else {
		total = evmDenomFromBank.Add(evmDenomFromEvmBank)
	}
	return sdk.NewCoin(EvmDenom, total)
}

// SendCoins transfers neuron coins from a AccAddress to an AccAddress.
func (k EvmBankKeeper) SendCoins(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// SendCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use SendCoins.
	panic("not implemented")
}

// SendCoinsFromModuleToAccount transfers neuron coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	ua0gi, neuron, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if ua0gi.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(ua0gi)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneUa0giToNeuronIfNeeded(ctx, senderAddr, neuron); err != nil {
		return err
	}

	if err := k.neuronKeeper.SendBalance(ctx, senderAddr, recipientAddr, neuron); err != nil {
		return err
	}

	return k.ConvertNeuronToUa0gi(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers neuron coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	ua0gi, neuronNeeded, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if ua0gi.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(ua0gi)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneUa0giToNeuronIfNeeded(ctx, senderAddr, neuronNeeded); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.neuronKeeper.SendBalance(ctx, senderAddr, recipientAddr, neuronNeeded); err != nil {
		return err
	}

	return k.ConvertNeuronToUa0gi(ctx, recipientAddr)
}

// MintCoins mints neuron coins by minting the equivalent a0gi coins and any remaining neuron coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ua0gi, neuron, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if ua0gi.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(ua0gi)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.neuronKeeper.AddBalance(ctx, recipientAddr, neuron); err != nil {
		return err
	}

	return k.ConvertNeuronToUa0gi(ctx, recipientAddr)
}

// BurnCoins burns neuron coins by burning the equivalent a0gi coins and any remaining neuron coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ua0gi, neuron, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if ua0gi.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(ua0gi)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneUa0giToNeuronIfNeeded(ctx, moduleAddr, neuron); err != nil {
		return err
	}

	return k.neuronKeeper.RemoveBalance(ctx, moduleAddr, neuron)
}

// ConvertOneUa0giToNeuronIfNeeded converts 1 a0gi to neuron for an address if
// its neuron balance is smaller than the neuronNeeded amount.
func (k EvmBankKeeper) ConvertOneUa0giToNeuronIfNeeded(ctx sdk.Context, addr sdk.AccAddress, neuronNeeded sdkmath.Int) error {
	neuronBal := k.neuronKeeper.GetBalance(ctx, addr)
	if neuronBal.GTE(neuronNeeded) {
		return nil
	}

	ua0giToStore := sdk.NewCoins(sdk.NewCoin(CosmosDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, ua0giToStore); err != nil {
		return err
	}

	// add 1a0gi equivalent of neuron to addr
	neuronToReceive := ConversionMultiplier
	if err := k.neuronKeeper.AddBalance(ctx, addr, neuronToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertNeuronToA0gi converts all available neuron to a0gi for a given AccAddress.
func (k EvmBankKeeper) ConvertNeuronToUa0gi(ctx sdk.Context, addr sdk.AccAddress) error {
	totalNeuron := k.neuronKeeper.GetBalance(ctx, addr)
	ua0gi, _, err := SplitNeuronCoins(sdk.NewCoins(sdk.NewCoin(EvmDenom, totalNeuron)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough neuron for a single a0gi
	ua0giToReceive := ua0gi.Amount
	if !ua0giToReceive.IsPositive() {
		return nil
	}

	// remove neuron used for converting to ua0gi
	neuronToBurn := ua0giToReceive.Mul(ConversionMultiplier)
	finalBal := totalNeuron.Sub(neuronToBurn)
	if err := k.neuronKeeper.SetBalance(ctx, addr, finalBal); err != nil {
		return err
	}

	fromAddr := k.GetModuleAddress(types.ModuleName)
	if err := k.bk.SendCoins(ctx, fromAddr, addr, sdk.NewCoins(ua0gi)); err != nil {
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

// SplitNeuronCoins splits neuron coins to the equivalent a0gi coins and any remaining neuron balance.
// An error will be returned if the coins are not valid or if the coins are not the neuron denom.
func SplitNeuronCoins(coins sdk.Coins) (sdk.Coin, sdkmath.Int, error) {
	neuron := sdk.ZeroInt()
	ua0gi := sdk.NewCoin(CosmosDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return ua0gi, neuron, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return ua0gi, neuron, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		neuron = remainingBalance
	}
	ua0giAmount := coin.Amount.Quo(ConversionMultiplier)
	if ua0giAmount.IsPositive() {
		ua0gi = sdk.NewCoin(CosmosDenom, ua0giAmount)
	}

	return ua0gi, neuron, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the EvmDenom (neuron).
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is neuron
	if len(coins) != 1 || coins[0].Denom != EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", EvmDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
