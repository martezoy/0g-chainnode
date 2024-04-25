package types

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _, _ sdk.Msg = &MsgRequestDAS{}, &MsgReportDASResult{}

func NewMsgRequestDAS(fromAddr sdk.AccAddress, streamID, hash string, numBlobs uint32) *MsgRequestDAS {
	return &MsgRequestDAS{
		Requester:       fromAddr.String(),
		StreamID:        streamID,
		BatchHeaderHash: hash,
		NumBlobs:        numBlobs,
	}
}

func (msg MsgRequestDAS) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgRequestDAS) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid requester account address (%s)", err)
	}

	return nil
}

func (msg *MsgReportDASResult) GetSigners() []sdk.AccAddress {
	samplerValAddr, err := sdk.ValAddressFromBech32(msg.Sampler)
	if err != nil {
		panic(err)
	}
	accAddr, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(samplerValAddr.Bytes()))
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

func (msg *MsgReportDASResult) ValidateBasic() error {
	_, err := sdk.ValAddressFromBech32(msg.Sampler)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sampler validator address (%s)", err)
	}
	return nil
}
