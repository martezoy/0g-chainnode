package dasigners

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/x/evm/statedb"
)

const (
	NewSignerEvent     = "NewSigner"
	SocketUpdatedEvent = "SocketUpdated"
)

func (d *DASignersPrecompile) EmitNewSignerEvent(ctx sdk.Context, stateDB *statedb.StateDB, signer IDASignersSignerDetail) error {
	event := d.abi.Events[NewSignerEvent]
	quries := make([]interface{}, 2)
	quries[0] = event.ID
	quries[1] = signer.Signer
	topics, err := abi.MakeTopics(quries)
	if err != nil {
		return err
	}
	b, err := event.Inputs.Pack(signer.Signer, signer.PkG1, signer.PkG2)
	if err != nil {
		return err
	}
	stateDB.AddLog(&types.Log{
		Address:     d.Address(),
		Topics:      topics[0],
		Data:        b,
		BlockNumber: uint64(ctx.BlockHeight()),
	})
	return d.EmitSocketUpdatedEvent(ctx, stateDB, signer.Signer, signer.Socket)
}

func (d *DASignersPrecompile) EmitSocketUpdatedEvent(ctx sdk.Context, stateDB *statedb.StateDB, signer common.Address, socket string) error {
	event := d.abi.Events[SocketUpdatedEvent]
	quries := make([]interface{}, 2)
	quries[0] = event.ID
	quries[1] = signer
	topics, err := abi.MakeTopics(quries)
	if err != nil {
		return err
	}
	b, err := event.Inputs.Pack(signer, socket)
	if err != nil {
		return err
	}
	stateDB.AddLog(&types.Log{
		Address:     d.Address(),
		Topics:      topics[0],
		Data:        b,
		BlockNumber: uint64(ctx.BlockHeight()),
	})
	return nil
}
