package keeper

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Ballot struct {
	account string
	content []byte
}

func (k Keeper) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		k.Logger(ctx).Error("[BeginBlock] cannot get epoch number")
		panic(err)
	}
	params := k.GetParams(ctx)
	expectedEpoch := uint64(ctx.BlockHeight()) / params.EpochBlocks
	if expectedEpoch == epochNumber {
		return
	}
	if expectedEpoch > epochNumber+1 || expectedEpoch < epochNumber {
		panic("block height is not continuous")
	}
	// new epoch
	registrations := []Ballot{}
	k.IterateRegistrations(ctx, expectedEpoch, func(account string, signature []byte) (stop bool) {
		registrations = append(registrations, Ballot{
			account: account,
			content: signature,
		})
		return false
	})
	ballots := []Ballot{}
	tokensPerVote, ok := sdk.NewIntFromString(params.TokensPerVote)
	if !ok {
		panic("failed to load params tokens_per_vote")
	}
	for _, registration := range registrations {
		// get validator
		valAddr, err := sdk.ValAddressFromHex(registration.account)
		if err != nil {
			k.Logger(ctx).Error("[BeginBlock] invalid account")
			continue
		}
		validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
		if !found {
			continue
		}
		num := validator.Tokens.Quo(sdk.NewInt(1_000_000_000_000_000_000)).Quo(tokensPerVote).Abs().BigInt()
		if num.Cmp(big.NewInt(int64(params.MaxVotes))) > 0 {
			num = big.NewInt(int64(params.MaxVotes))
		}
		content := registration.content
		ballotNum := num.Int64()
		for j := 0; j < int(ballotNum); j += 1 {
			ballots = append(ballots, Ballot{
				account: registration.account,
				content: content,
			})
			content = crypto.Keccak256(content)
		}
	}
	sort.Slice(ballots, func(i, j int) bool {
		return bytes.Compare(ballots[i].content, ballots[j].content) < 0
	})

	quorums := types.Quorums{
		Quorums: make([]*types.Quorum, 0),
	}
	if len(ballots) >= int(params.EncodedSlices) {
		for i := 0; i+int(params.EncodedSlices) < len(ballots); i += 1 {
			quorum := types.Quorum{
				Signers: make([]string, params.EncodedSlices),
			}
			for j := 0; j < int(params.EncodedSlices); j += 1 {
				quorum.Signers[j] = ballots[i+j].account
			}
			quorums.Quorums = append(quorums.Quorums, &quorum)
		}
	} else {
		quorum := types.Quorum{
			Signers: make([]string, params.EncodedSlices),
		}
		n := len(ballots)
		for i := 0; i < int(params.EncodedSlices); i += 1 {
			quorum.Signers[i] = ballots[i%n].account
		}
		quorums.Quorums = append(quorums.Quorums, &quorum)
	}

	// save to store
	k.SetEpochQuorums(ctx, expectedEpoch, quorums)
	k.SetEpochNumber(ctx, expectedEpoch)
}
