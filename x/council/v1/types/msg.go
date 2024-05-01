package types

import (
	"encoding/hex"

	"github.com/coniks-sys/coniks-go/crypto/vrf"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _, _ sdk.Msg = &MsgRegister{}, &MsgVote{}

// GetSigners returns the expected signers for a MsgRegister message.
func (msg *MsgRegister) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromBech32(msg.Voter)
	if err != nil {
		panic(err)
	}
	accAddr, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(valAddr.Bytes()))
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// ValidateBasic does a sanity check of the provided data
func (msg *MsgRegister) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.Voter); err != nil {
		return ErrInvalidValidatorAddress
	}
	if len(msg.Key) != vrf.PublicKeySize {
		return ErrInvalidPublicKey
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRegister) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgVote message.
func (msg *MsgVote) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromBech32(msg.Voter)
	if err != nil {
		panic(err)
	}
	accAddr, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(valAddr.Bytes()))
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// ValidateBasic does a sanity check of the provided data
func (msg *MsgVote) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.Voter); err != nil {
		return ErrInvalidValidatorAddress
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgVote) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}
