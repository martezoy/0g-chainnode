package types

import (
	"encoding/hex"
	fmt "fmt"

	"github.com/0glabs/0g-chain/crypto/bn254util"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _, _, _ sdk.Msg = &MsgRegisterSigner{}, &MsgUpdateSocket{}, &MsgRegisterNextEpoch{}

// GetSigners returns the expected signers for a MsgRegister message.
func (msg *MsgRegisterSigner) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromHex(msg.Signer.Account)
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
func (msg *MsgRegisterSigner) ValidateBasic() error {
	if err := msg.Signer.Validate(); err != nil {
		return err
	}
	if len(msg.Signature) != bn254util.G1PointSize {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRegisterSigner) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgVote message.
func (msg *MsgUpdateSocket) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromHex(msg.Account)
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
func (msg *MsgUpdateSocket) ValidateBasic() error {
	if err := ValidateHexAddress(msg.Account); err != nil {
		return err
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateSocket) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgVote message.
func (msg *MsgRegisterNextEpoch) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromHex(msg.Account)
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
func (msg *MsgRegisterNextEpoch) ValidateBasic() error {
	if err := ValidateHexAddress(msg.Account); err != nil {
		return err
	}
	if len(msg.Signature) != bn254util.G1PointSize {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRegisterNextEpoch) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}
