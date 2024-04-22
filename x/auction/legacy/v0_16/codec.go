package types

import (
	v017auction "github.com/0glabs/0g-chain/x/auction/types"
	types "github.com/cosmos/cosmos-sdk/codec/types"
)

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"kava.auction.v1beta1.GenesisAuction",
		(*v017auction.GenesisAuction)(nil),
		&v017auction.SurplusAuction{},
		&v017auction.DebtAuction{},
		&v017auction.CollateralAuction{},
	)
}
