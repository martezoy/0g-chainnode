package community_test

import (
	"testing"

	"github.com/0glabs/0g-chain/x/community"
	"github.com/0glabs/0g-chain/x/community/keeper"
	"github.com/0glabs/0g-chain/x/community/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

func TestABCIPayoutAccumulatedStakingRewards(t *testing.T) {
	testFunc := func(ctx sdk.Context, k keeper.Keeper) {
		community.BeginBlocker(ctx, k)
	}
	suite.Run(t, testutil.NewStakingRewardsTestSuite(testFunc))
}
