package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrSignerExists               = errorsmod.Register(ModuleName, 1, "signer exists")
	ErrEpochNumberNotSet          = errorsmod.Register(ModuleName, 2, "epoch number not set")
	ErrSignerNotFound             = errorsmod.Register(ModuleName, 3, "signer not found")
	ErrInvalidSignature           = errorsmod.Register(ModuleName, 4, "invalid signature")
	ErrQuorumNotFound             = errorsmod.Register(ModuleName, 5, "quorum for epoch not found")
	ErrQuorumIdOutOfBound         = errorsmod.Register(ModuleName, 6, "quorum id out of bound")
	ErrQuorumBitmapLengthMismatch = errorsmod.Register(ModuleName, 7, "quorum bitmap length mismatch")
	ErrInsufficientBonded         = errorsmod.Register(ModuleName, 8, "insufficient bonded amount")
)
