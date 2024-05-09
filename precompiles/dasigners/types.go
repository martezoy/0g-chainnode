package dasigners

import (
	"fmt"
	"math/big"
	"strings"

	precopmiles_common "github.com/0glabs/0g-chain/precompiles/common"
	dasignerstypes "github.com/0glabs/0g-chain/x/dasigners/v1/types"
	"github.com/ethereum/go-ethereum/common"
)

func NewBN254G1Point(b []byte) BN254G1Point {
	return BN254G1Point{
		X: new(big.Int).SetBytes(b[:32]),
		Y: new(big.Int).SetBytes(b[32:64]),
	}
}

func (p BN254G1Point) Serialize() []byte {
	b := make([]byte, 0)
	b = append(b, common.LeftPadBytes(p.X.Bytes(), 32)...)
	b = append(b, common.LeftPadBytes(p.Y.Bytes(), 32)...)
	return b
}

func NewBN254G2Point(b []byte) BN254G2Point {
	return BN254G2Point{
		X: [2]*big.Int{
			new(big.Int).SetBytes(b[:32]),
			new(big.Int).SetBytes(b[32:64]),
		},
		Y: [2]*big.Int{
			new(big.Int).SetBytes(b[64:96]),
			new(big.Int).SetBytes(b[96:128]),
		},
	}
}

func (p BN254G2Point) Serialize() []byte {
	b := make([]byte, 0)
	b = append(b, common.LeftPadBytes(p.X[0].Bytes(), 32)...)
	b = append(b, common.LeftPadBytes(p.X[1].Bytes(), 32)...)
	b = append(b, common.LeftPadBytes(p.Y[0].Bytes(), 32)...)
	b = append(b, common.LeftPadBytes(p.Y[1].Bytes(), 32)...)
	return b
}

func NewQuerySignerRequest(args []interface{}) (*dasignerstypes.QuerySignerRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(precopmiles_common.ErrInvalidNumberOfArgs, 1, len(args))
	}

	return &dasignerstypes.QuerySignerRequest{
		Account: args[0].(string),
	}, nil
}

func NewQueryEpochSignerSetRequest(args []interface{}) (*dasignerstypes.QueryEpochSignerSetRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(precopmiles_common.ErrInvalidNumberOfArgs, 1, len(args))
	}

	return &dasignerstypes.QueryEpochSignerSetRequest{
		EpochNumber: args[0].(*big.Int).Uint64(),
	}, nil
}

func NewQueryAggregatePubkeyG1Request(args []interface{}) (*dasignerstypes.QueryAggregatePubkeyG1Request, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(precopmiles_common.ErrInvalidNumberOfArgs, 2, len(args))
	}

	return &dasignerstypes.QueryAggregatePubkeyG1Request{
		EpochNumber:   args[0].(*big.Int).Uint64(),
		SignersBitmap: args[1].([]byte),
	}, nil
}

func NewIDASignersSignerDetail(signer *dasignerstypes.Signer) IDASignersSignerDetail {
	return IDASignersSignerDetail{
		Signer: common.HexToAddress(signer.Account),
		Socket: signer.Socket,
		PkG1:   NewBN254G1Point(signer.PubkeyG1),
		PkG2:   NewBN254G2Point(signer.PubkeyG2),
	}
}

func ToLowerHexWithoutPrefix(addr common.Address) string {
	return strings.ToLower(addr.Hex()[2:])
}

func NewMsgRegisterSigner(args []interface{}) (*dasignerstypes.MsgRegisterSigner, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(precopmiles_common.ErrInvalidNumberOfArgs, 2, len(args))
	}

	signer := args[0].(IDASignersSignerDetail)
	return &dasignerstypes.MsgRegisterSigner{
		Signer: &dasignerstypes.Signer{
			Account:  ToLowerHexWithoutPrefix(signer.Signer),
			Socket:   signer.Socket,
			PubkeyG1: signer.PkG1.Serialize(),
			PubkeyG2: signer.PkG2.Serialize(),
		},
		Signature: args[1].(BN254G1Point).Serialize(),
	}, nil
}

func NewMsgRegisterNextEpoch(args []interface{}, account string) (*dasignerstypes.MsgRegisterNextEpoch, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(precopmiles_common.ErrInvalidNumberOfArgs, 1, len(args))
	}

	return &dasignerstypes.MsgRegisterNextEpoch{
		Account:   account,
		Signature: args[0].(BN254G1Point).Serialize(),
	}, nil
}

func NewMsgUpdateSocket(args []interface{}, account string) (*dasignerstypes.MsgUpdateSocket, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(precopmiles_common.ErrInvalidNumberOfArgs, 1, len(args))
	}

	return &dasignerstypes.MsgUpdateSocket{
		Account: account,
		Socket:  args[0].(string),
	}, nil
}
