package v2

import (
	"github.com/0glabs/0g-chain/x/cdp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// MigrateStore performs in-place store migrations for consensus version 2
// V2 adds the begin_blocker_execution_block_interval param to parameters.
func MigrateStore(ctx sdk.Context, paramstore paramtypes.Subspace) error {
	migrateParamsStore(ctx, paramstore)
	return nil
}

// migrateParamsStore ensures the param key table exists and has the begin_blocker_execution_block_interval property
func migrateParamsStore(ctx sdk.Context, paramstore paramtypes.Subspace) {
	if !paramstore.HasKeyTable() {
		paramstore.WithKeyTable(types.ParamKeyTable())
	}
	paramstore.Set(ctx, types.KeyBeginBlockerExecutionBlockInterval, types.DefaultBeginBlockerExecutionBlockInterval)
}
