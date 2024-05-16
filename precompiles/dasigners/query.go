package dasigners

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

func (d *DASignersPrecompile) EpochNumber(ctx sdk.Context, _ *vm.EVM, method *abi.Method, _ []interface{}) ([]byte, error) {
	epochNumber, err := d.dasignersKeeper.GetEpochNumber(ctx)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack(big.NewInt(int64(epochNumber)))
}

func (d *DASignersPrecompile) QuorumCount(ctx sdk.Context, _ *vm.EVM, method *abi.Method, args []interface{}) ([]byte, error) {
	req, err := NewQueryQuorumCountRequest(args)
	response, err := d.dasignersKeeper.QuorumCount(ctx, req)
	if err != nil {
		return nil, err
	}
	return method.Outputs.Pack(big.NewInt(int64(response.QuorumCount)))
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
	signers := make([]IDASignersSignerDetail, len(response.Signer))
	for i, signer := range response.Signer {
		signers[i] = NewIDASignersSignerDetail(signer)
	}
	return method.Outputs.Pack(signers)
}

func (d *DASignersPrecompile) GetQuorum(ctx sdk.Context, _ *vm.EVM, method *abi.Method, args []interface{}) ([]byte, error) {
	req, err := NewQueryEpochQuorumRequest(args)
	if err != nil {
		return nil, err
	}
	response, err := d.dasignersKeeper.EpochQuorum(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, err
	}
	signers := make([]common.Address, len(response.Quorum.Signers))
	for i, signer := range response.Quorum.Signers {
		signers[i] = common.HexToAddress(signer)
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
