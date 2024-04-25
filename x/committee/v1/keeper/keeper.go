package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0glabs/0g-chain/x/committee/v1/types"
)

// Keeper of the inflation store
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates a new mint Keeper instance
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

// ------------------------------------------
//				Committees
// ------------------------------------------

func (k Keeper) SetCurrentCommitteeID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.CurrentCommitteeIDKey, types.GetKeyFromID(id))
}

func (k Keeper) GetCurrentCommitteeID(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CurrentCommitteeIDKey)
	if bz == nil {
		return 0, errorsmod.Wrap(types.ErrInvalidGenesis, "current committee ID not set at genesis")
	}
	return types.Uint64FromBytes(bz), nil
}

func (k Keeper) IncrementCurrentCommitteeID(ctx sdk.Context) error {
	id, err := k.GetCurrentCommitteeID(ctx)
	if err != nil {
		return err
	}
	k.SetCurrentCommitteeID(ctx, id+1)
	return nil
}

func (k Keeper) SetVotingStartHeight(ctx sdk.Context, votingStartHeight uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.VotingStartHeightKey, types.GetKeyFromID(votingStartHeight))
}

func (k Keeper) GetVotingStartHeight(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.VotingStartHeightKey)
	if bz == nil {
		return 0, errorsmod.Wrap(types.ErrInvalidGenesis, "voting start height not set at genesis")
	}
	return types.Uint64FromBytes(bz), nil
}

func (k Keeper) SetVotingPeriod(ctx sdk.Context, votingPeriod uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.VotingPeriodKey, types.GetKeyFromID(votingPeriod))
}

func (k Keeper) GetVotingPeriod(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.VotingPeriodKey)
	if bz == nil {
		return 0, errorsmod.Wrap(types.ErrInvalidGenesis, "voting period not set at genesis")
	}
	return types.Uint64FromBytes(bz), nil
}

// StoreNewCommittee stores a committee, adding a new ID
func (k Keeper) StoreNewCommittee(ctx sdk.Context, votingStartHeight uint64) error {
	currentCommitteeID, err := k.GetCurrentCommitteeID(ctx)
	if err != nil {
		return err
	}

	votingPeriod, err := k.GetVotingPeriod(ctx)
	if err != nil {
		return err
	}
	com := types.Committee{
		ID:                currentCommitteeID + 1,
		VotingStartHeight: votingStartHeight,
		StartHeight:       votingStartHeight + votingPeriod,
		EndHeight:         votingStartHeight + votingPeriod*2,
		Votes:             []types.Vote{},
		Members:           []sdk.ValAddress{},
	}
	k.SetCommittee(ctx, com)

	return nil
}

func (k Keeper) GetCommittee(ctx sdk.Context, committeeID uint64) (types.Committee, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	bz := store.Get(types.GetKeyFromID(committeeID))
	if bz == nil {
		return types.Committee{}, false
	}
	var com types.Committee
	k.cdc.MustUnmarshal(bz, &com)
	return com, true
}

// SetCommittee puts a committee into the store.
func (k Keeper) SetCommittee(ctx sdk.Context, committee types.Committee) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	bz := k.cdc.MustMarshal(&committee)
	store.Set(types.GetKeyFromID(committee.ID), bz)
}

// // DeleteProposal removes a proposal from the store.
// func (k Keeper) DeleteProposal(ctx sdk.Context, proposalID uint64) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposalKeyPrefix)
// 	store.Delete(types.GetKeyFromID(proposalID))
// }

// IterateProposals provides an iterator over all stored proposals.
// For each proposal, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateCommittee(ctx sdk.Context, cb func(proposal types.Committee) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var committee types.Committee
		k.cdc.MustUnmarshal(iterator.Value(), &committee)
		if cb(committee) {
			break
		}
	}
}

func (k Keeper) GetCommittees(ctx sdk.Context) types.Committees {
	results := types.Committees{}
	k.IterateCommittee(ctx, func(prop types.Committee) bool {
		results = append(results, prop)
		return false
	})
	return results
}

// // DeleteProposalAndVotes removes a proposal and its associated votes.
// func (k Keeper) DeleteProposalAndVotes(ctx sdk.Context, proposalID uint64) {
// 	votes := k.GetVotesByProposal(ctx, proposalID)
// 	k.DeleteProposal(ctx, proposalID)
// 	for _, v := range votes {
// 		k.DeleteVote(ctx, v.ProposalID, v.Voter)
// 	}
// }

// ------------------------------------------
//				Votes
// ------------------------------------------

// GetVote gets a vote from the store.
func (k Keeper) GetVote(ctx sdk.Context, epochID uint64, voter sdk.ValAddress) (types.Vote, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	bz := store.Get(types.GetVoteKey(epochID, voter))
	if bz == nil {
		return types.Vote{}, false
	}
	var vote types.Vote
	k.cdc.MustUnmarshal(bz, &vote)
	return vote, true
}

// SetVote puts a vote into the store.
func (k Keeper) SetVote(ctx sdk.Context, vote types.Vote) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	bz := k.cdc.MustMarshal(&vote)
	store.Set(types.GetVoteKey(vote.CommitteeID, vote.Voter), bz)
}

// DeleteVote removes a Vote from the store.
func (k Keeper) DeleteVote(ctx sdk.Context, committeeID uint64, voter sdk.ValAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	store.Delete(types.GetVoteKey(committeeID, voter))
}

// IterateVotes provides an iterator over all stored votes.
// For each vote, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateVotes(ctx sdk.Context, cb func(vote types.Vote) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		k.cdc.MustUnmarshal(iterator.Value(), &vote)

		if cb(vote) {
			break
		}
	}
}

// GetVotes returns all stored votes.
func (k Keeper) GetVotes(ctx sdk.Context) []types.Vote {
	results := []types.Vote{}
	k.IterateVotes(ctx, func(vote types.Vote) bool {
		results = append(results, vote)
		return false
	})
	return results
}

// GetVotesByProposal returns all votes for one proposal.
func (k Keeper) GetVotesByCommittee(ctx sdk.Context, committeeID uint64) []types.Vote {
	results := []types.Vote{}
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), append(types.VoteKeyPrefix, types.GetKeyFromID(committeeID)...))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		k.cdc.MustUnmarshal(iterator.Value(), &vote)
		results = append(results, vote)
	}

	return results
}

// ------------------------------------------
//				Voters
// ------------------------------------------

func (k Keeper) SetVoter(ctx sdk.Context, voter sdk.ValAddress, pk vrf.PublicKey) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoterKeyPrefix)
	store.Set(types.GetVoterKey(voter), pk)
	fmt.Printf("voterStoreKey: %v, publicKey: %v\n", types.GetVoterKey(voter), pk)
}

func (k Keeper) IterateVoters(ctx sdk.Context, cb func(voter sdk.ValAddress, pk vrf.PublicKey) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.VoterKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		if cb(sdk.ValAddress(iterator.Key()[1:]), vrf.PublicKey(iterator.Value())) {
			break
		}
	}
}

// GetVotes returns all stored voters
func (k Keeper) GetVoters(ctx sdk.Context) []sdk.ValAddress {
	results := []sdk.ValAddress{}
	k.IterateVoters(ctx, func(voter sdk.ValAddress, _ vrf.PublicKey) bool {
		results = append(results, voter)
		return false
	})
	return results
}

func (k Keeper) AddVoter(ctx sdk.Context, voter sdk.ValAddress, key []byte) error {
	if len(key) != vrf.PublicKeySize {
		return types.ErrInvalidPublicKey
	}

	k.SetVoter(ctx, voter, vrf.PublicKey(key))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegister,
			sdk.NewAttribute(types.AttributeKeyVoter, voter.String()),
			// TODO: types.AttributeKeyPublicKey
		),
	)

	return nil
}

func (k Keeper) AddVote(ctx sdk.Context, committeeID uint64, voter sdk.ValAddress, ballots []*types.Ballot) error {
	// Validate
	com, found := k.GetCommittee(ctx, committeeID)
	if !found {
		return errorsmod.Wrapf(types.ErrUnknownCommittee, "%d", committeeID)
	}
	if com.HasVotingEndedBy(ctx.BlockHeight()) {
		return errorsmod.Wrapf(types.ErrProposalExpired, "%d ≥ %d", ctx.BlockHeight(), com.StartHeight)
	}

	// TODO: verify if the voter is registered
	// TODO: verify whether ballots are valid or not

	// Store vote, overwriting any prior vote
	k.SetVote(ctx, types.NewVote(committeeID, voter, ballots))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVote,
			sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", com.ID)),
			sdk.NewAttribute(types.AttributeKeyVoter, voter.String()),
			// TODO: types.AttributeKeyBallots
		),
	)

	return nil
}
