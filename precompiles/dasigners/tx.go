package dasigners

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/x/evm/statedb"
)

func (d *DASignersPrecompile) RegisterSigner(ctx sdk.Context, evm *vm.EVM, stateDB *statedb.StateDB, method *abi.Method, args []interface{}) ([]byte, error) {
	msg, err := NewMsgRegisterSigner(args)
	if err != nil {
		return nil, err
	}
	// validation
	sender := ToLowerHexWithoutPrefix(evm.Origin)
	if sender != msg.Signer.Account {
		return nil, fmt.Errorf(ErrInvalidSender, sender, msg.Signer.Account)
	}
	// execute
	_, err = d.dasignersKeeper.RegisterSigner(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}
	// emit events
	err = d.EmitNewSignerEvent(ctx, stateDB, args[0].(IDASignersSignerDetail))
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}

func (d *DASignersPrecompile) RegisterNextEpoch(ctx sdk.Context, evm *vm.EVM, stateDB *statedb.StateDB, method *abi.Method, args []interface{}) ([]byte, error) {
	msg, err := NewMsgRegisterNextEpoch(args, ToLowerHexWithoutPrefix(evm.Origin))
	if err != nil {
		return nil, err
	}
	// execute
	_, err = d.dasignersKeeper.RegisterNextEpoch(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}

func (d *DASignersPrecompile) UpdateSocket(ctx sdk.Context, evm *vm.EVM, stateDB *statedb.StateDB, method *abi.Method, args []interface{}) ([]byte, error) {
	msg, err := NewMsgUpdateSocket(args, ToLowerHexWithoutPrefix(evm.Origin))
	if err != nil {
		return nil, err
	}
	// execute
	_, err = d.dasignersKeeper.UpdateSocket(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}
	// emit events
	err = d.EmitSocketUpdatedEvent(ctx, stateDB, evm.Origin, args[0].(string))
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack()
}
