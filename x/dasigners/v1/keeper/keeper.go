package keeper

import (
	"encoding/hex"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/0glabs/0g-chain/chaincfg"
	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
)

var BondedConversionRate = math.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(chaincfg.GasDenomUnit), nil))

type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new das Keeper instance
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	stakingKeeper types.StakingKeeper,
) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		stakingKeeper: stakingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
}

func (k Keeper) GetEpochNumber(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		return 0, types.ErrEpochNumberNotSet
	}
	return sdk.BigEndianToUint64(bz), nil
}

func (k Keeper) SetEpochNumber(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.EpochNumberKey, sdk.Uint64ToBigEndian(epoch))
}

func (k Keeper) GetQuorumCount(ctx sdk.Context, epoch uint64) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.QuorumCountKeyPrefix)
	bz := store.Get(types.GetQuorumCountKey(epoch))
	if bz == nil {
		return 0, types.ErrQuorumNotFound
	}
	return sdk.BigEndianToUint64(bz), nil
}

func (k Keeper) SetQuorumCount(ctx sdk.Context, epoch uint64, quorums uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.QuorumCountKeyPrefix)
	store.Set(types.GetQuorumCountKey(epoch), sdk.Uint64ToBigEndian(quorums))
}

func (k Keeper) GetSigner(ctx sdk.Context, account string) (types.Signer, bool, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SignerKeyPrefix)
	key, err := types.GetSignerKeyFromAccount(account)
	if err != nil {
		return types.Signer{}, false, err
	}
	bz := store.Get(key)
	if bz == nil {
		return types.Signer{}, false, nil
	}
	var signer types.Signer
	k.cdc.MustUnmarshal(bz, &signer)
	return signer, true, nil
}

func (k Keeper) SetSigner(ctx sdk.Context, signer types.Signer) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SignerKeyPrefix)
	bz := k.cdc.MustMarshal(&signer)
	key, err := types.GetSignerKeyFromAccount(signer.Account)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateSigner,
			sdk.NewAttribute(types.AttributeKeySigner, signer.Account),
			sdk.NewAttribute(types.AttributeKeySocket, signer.Socket),
			sdk.NewAttribute(types.AttributeKeyPublicKeyG1, hex.EncodeToString(signer.PubkeyG1)),
			sdk.NewAttribute(types.AttributeKeyPublicKeyG2, hex.EncodeToString(signer.PubkeyG2)),
		),
	)
	return nil
}

// iterate through the signers set and perform the provided function
func (k Keeper) IterateSigners(ctx sdk.Context, fn func(index int64, signer types.Signer) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	prefix := types.SignerKeyPrefix
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		var signer types.Signer
		k.cdc.MustUnmarshal(iterator.Value(), &signer)
		stop := fn(i, signer)

		if stop {
			break
		}
		i++
	}
}

func (k Keeper) GetEpochQuorums(ctx sdk.Context, epoch uint64) (types.Quorums, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.EpochQuorumsKeyPrefix)
	bz := store.Get(types.GetEpochQuorumsKeyFromEpoch(epoch))
	if bz == nil {
		return types.Quorums{}, false
	}
	var quorums types.Quorums
	k.cdc.MustUnmarshal(bz, &quorums)
	return quorums, true
}

func (k Keeper) SetEpochQuorums(ctx sdk.Context, epoch uint64, quorums types.Quorums) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.EpochQuorumsKeyPrefix)
	bz := k.cdc.MustMarshal(&quorums)
	store.Set(types.GetEpochQuorumsKeyFromEpoch(epoch), bz)
	k.SetQuorumCount(ctx, epoch, uint64(len(quorums.Quorums)))
}

func (k Keeper) GetRegistration(ctx sdk.Context, epoch uint64, account string) ([]byte, bool, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetEpochRegistrationKeyPrefix(epoch))
	key, err := types.GetRegistrationKey(account)
	if err != nil {
		return nil, false, err
	}
	signature := store.Get(key)
	if signature == nil {
		return nil, false, nil
	}
	return signature, true, nil
}

// iterate through the registrations set and perform the provided function
func (k Keeper) IterateRegistrations(ctx sdk.Context, epoch uint64, fn func(account string, signature []byte) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	prefix := types.GetEpochRegistrationKeyPrefix(epoch)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		stop := fn(hex.EncodeToString((iterator.Key())[len(prefix):]), iterator.Value())

		if stop {
			break
		}
		i++
	}
}

func (k Keeper) SetRegistration(ctx sdk.Context, epoch uint64, account string, signature []byte) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetEpochRegistrationKeyPrefix(epoch))
	key, err := types.GetRegistrationKey(account)
	if err != nil {
		return err
	}
	store.Set(key, signature)
	return nil
}

func (k Keeper) GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) math.Int {
	bonded := sdk.ZeroDec()

	cnt := 0
	k.stakingKeeper.IterateDelegatorDelegations(ctx, delegator, func(delegation stakingtypes.Delegation) bool {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := k.stakingKeeper.GetValidator(ctx, validatorAddr)
		if found {
			shares := delegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares)
			bonded = bonded.Add(tokens)
		}
		cnt += 1
		return cnt > 10
	})
	return bonded.RoundInt()
}

func (k Keeper) CheckDelegations(ctx sdk.Context, account string) error {
	accAddr, err := sdk.AccAddressFromHexUnsafe(account)
	if err != nil {
		return err
	}
	bonded := k.GetDelegatorBonded(ctx, accAddr)
	params := k.GetParams(ctx)
	tokensPerVote := sdk.NewIntFromUint64(params.TokensPerVote)
	if bonded.Quo(BondedConversionRate).Quo(tokensPerVote).Abs().BigInt().Cmp(big.NewInt(0)) <= 0 {
		return types.ErrInsufficientBonded
	}
	return nil
}
