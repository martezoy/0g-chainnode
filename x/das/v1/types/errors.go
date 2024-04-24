package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrUnknownRequest = errorsmod.Register(ModuleName, 0, "request not found")
	ErrInvalidGenesis = errorsmod.Register(ModuleName, 1, "invalid genesis")
)
