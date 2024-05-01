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

// ConversionMultiplier is the conversion multiplier between neuron and a0gi
var ConversionMultiplier = sdkmath.NewInt(chaincfg.ConversionMultiplier)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal neuron coin on the evm.
// x/evm consumes gas and send coins by minting and burning neuron coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the a0gi coin and a separate neuron balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	baseKeeper Keeper
	bk         types.BankKeeper
	ak         types.AccountKeeper
}

func NewEvmBankKeeper(baseKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		baseKeeper: baseKeeper,
		bk:         bk,
		ak:         ak,
	}
}

// GetBalance returns the total **spendable** balance of neuron for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != chaincfg.BaseDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", chaincfg.BaseDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	a0gi := spendableCoins.AmountOf(chaincfg.DisplayDenom)
	neuron := k.baseKeeper.GetBalance(ctx, addr)
	total := a0gi.Mul(ConversionMultiplier).Add(neuron)
	return sdk.NewCoin(chaincfg.BaseDenom, total)
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
	a0gi, neuron, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if a0gi.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(a0gi)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneA0giToNeuronIfNeeded(ctx, senderAddr, neuron); err != nil {
		return err
	}

	if err := k.baseKeeper.SendBalance(ctx, senderAddr, recipientAddr, neuron); err != nil {
		return err
	}

	return k.ConvertNeuronToA0gi(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers neuron coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	a0gi, neuronNeeded, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if a0gi.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(a0gi)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneA0giToNeuronIfNeeded(ctx, senderAddr, neuronNeeded); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.baseKeeper.SendBalance(ctx, senderAddr, recipientAddr, neuronNeeded); err != nil {
		return err
	}

	return k.ConvertNeuronToA0gi(ctx, recipientAddr)
}

// MintCoins mints neuron coins by minting the equivalent a0gi coins and any remaining neuron coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	a0gi, neuron, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if a0gi.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(a0gi)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.baseKeeper.AddBalance(ctx, recipientAddr, neuron); err != nil {
		return err
	}

	return k.ConvertNeuronToA0gi(ctx, recipientAddr)
}

// BurnCoins burns neuron coins by burning the equivalent a0gi coins and any remaining neuron coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	a0gi, neuron, err := SplitNeuronCoins(amt)
	if err != nil {
		return err
	}

	if a0gi.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(a0gi)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneA0giToNeuronIfNeeded(ctx, moduleAddr, neuron); err != nil {
		return err
	}

	return k.baseKeeper.RemoveBalance(ctx, moduleAddr, neuron)
}

// ConvertOneA0giToNeuronIfNeeded converts 1 a0gi to neuron for an address if
// its neuron balance is smaller than the neuronNeeded amount.
func (k EvmBankKeeper) ConvertOneA0giToNeuronIfNeeded(ctx sdk.Context, addr sdk.AccAddress, neuronNeeded sdkmath.Int) error {
	neuronBal := k.baseKeeper.GetBalance(ctx, addr)
	if neuronBal.GTE(neuronNeeded) {
		return nil
	}

	a0giToStore := sdk.NewCoins(sdk.NewCoin(chaincfg.DisplayDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, a0giToStore); err != nil {
		return err
	}

	// add 1a0gi equivalent of neuron to addr
	neuronToReceive := ConversionMultiplier
	if err := k.baseKeeper.AddBalance(ctx, addr, neuronToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertNeuronToA0gi converts all available neuron to a0gi for a given AccAddress.
func (k EvmBankKeeper) ConvertNeuronToA0gi(ctx sdk.Context, addr sdk.AccAddress) error {
	totalNeuron := k.baseKeeper.GetBalance(ctx, addr)
	a0gi, _, err := SplitNeuronCoins(sdk.NewCoins(sdk.NewCoin(chaincfg.BaseDenom, totalNeuron)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough neuron for a single a0gi
	a0giToReceive := a0gi.Amount
	if !a0giToReceive.IsPositive() {
		return nil
	}

	// remove neuron used for converting to a0gi
	neuronToBurn := a0giToReceive.Mul(ConversionMultiplier)
	finalBal := totalNeuron.Sub(neuronToBurn)
	if err := k.baseKeeper.SetBalance(ctx, addr, finalBal); err != nil {
		return err
	}

	fromAddr := k.GetModuleAddress(types.ModuleName)
	if err := k.bk.SendCoins(ctx, fromAddr, addr, sdk.NewCoins(a0gi)); err != nil {
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
	a0gi := sdk.NewCoin(chaincfg.DisplayDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return a0gi, neuron, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return a0gi, neuron, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		neuron = remainingBalance
	}
	a0giAmount := coin.Amount.Quo(ConversionMultiplier)
	if a0giAmount.IsPositive() {
		a0gi = sdk.NewCoin(chaincfg.DisplayDenom, a0giAmount)
	}

	return a0gi, neuron, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the chaincfg.BaseDenom (neuron).
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is neuron
	if len(coins) != 1 || coins[0].Denom != chaincfg.BaseDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", chaincfg.BaseDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
