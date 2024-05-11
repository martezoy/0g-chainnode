package dasigners

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
)

func (d *DASignersPrecompile) EpochNumber(ctx sdk.Context, _ *vm.EVM, method *abi.Method, _ []interface{}) ([]byte, error) {
	epochNumber, err := d.dasignersKeeper.GetEpochNumber(ctx)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack(big.NewInt(int64(epochNumber)))
}

func (d *DASignersPrecompile) GetSigner(ctx sdk.Context, _ *vm.EVM, method *abi.Method, args []interface{}) ([]byte, error) {
	req, err := NewQuerySignerRequest(args)
	if err != nil {
		return nil, err
	}
	response, err := d.dasignersKeeper.Signer(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack(NewIDASignersSignerDetail(response.Signer))
}

func (d *DASignersPrecompile) GetSigners(ctx sdk.Context, _ *vm.EVM, method *abi.Method, args []interface{}) ([]byte, error) {
	req, err := NewQueryEpochSignerSetRequest(args)
	if err != nil {
		return nil, err
	}
	response, err := d.dasignersKeeper.EpochSignerSet(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, err
	}
	signers := make([]IDASignersSignerDetail, 0)
	for _, signer := range response.Signers {
		signers = append(signers, NewIDASignersSignerDetail(signer))
	}
	return method.Outputs.Pack(signers)
}

func (d *DASignersPrecompile) GetAggPkG1(ctx sdk.Context, _ *vm.EVM, method *abi.Method, args []interface{}) ([]byte, error) {
	req, err := NewQueryAggregatePubkeyG1Request(args)
	if err != nil {
		return nil, err
	}
	response, err := d.dasignersKeeper.AggregatePubkeyG1(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack(NewBN254G1Point(response.AggregatePubkeyG1), big.NewInt(int64(response.Total)), big.NewInt(int64(response.Hit)))
}
