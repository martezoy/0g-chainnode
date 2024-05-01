package community_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/0glabs/0g-chain/app"
	"github.com/0glabs/0g-chain/x/community/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1})
	tApp.InitializeFromGenesisStates()

	accKeeper := tApp.GetAccountKeeper()
	acc := accKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
