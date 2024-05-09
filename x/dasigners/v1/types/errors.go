package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrSignerExists           = errorsmod.Register(ModuleName, 1, "signer exists")
	ErrEpochNumberNotSet      = errorsmod.Register(ModuleName, 2, "epoch number not set")
	ErrSignerNotFound         = errorsmod.Register(ModuleName, 3, "signer not found")
	ErrInvalidSignature       = errorsmod.Register(ModuleName, 4, "invalid signature")
	ErrEpochSignerSetNotFound = errorsmod.Register(ModuleName, 5, "signer set for epoch not found")
	ErrSignerLengthNotMatch   = errorsmod.Register(ModuleName, 6, "signer set length not match")
)
