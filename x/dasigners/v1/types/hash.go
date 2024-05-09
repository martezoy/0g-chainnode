package types

import (
	"math/big"

	"github.com/0glabs/0g-chain/crypto/bn254util"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func PubkeyRegistrationHash(operatorAddress common.Address, chainId *big.Int) *bn254.G1Affine {
	toHash := make([]byte, 0)
	toHash = append(toHash, operatorAddress.Bytes()...)
	// make sure chainId is 32 bytes
	toHash = append(toHash, common.LeftPadBytes(chainId.Bytes(), 32)...)
	toHash = append(toHash, []byte("0G_BN254_Pubkey_Registration")...)

	msgHash := crypto.Keccak256(toHash)
	// convert to [32]byte
	var msgHash32 [32]byte
	copy(msgHash32[:], msgHash)

	// hash to G1
	return bn254util.MapToCurve(msgHash32)
}

func EpochRegistrationHash(operatorAddress common.Address, epoch uint64, chainId *big.Int) *bn254.G1Affine {
	toHash := make([]byte, 0)
	toHash = append(toHash, operatorAddress.Bytes()...)
	toHash = append(toHash, sdk.Uint64ToBigEndian(epoch)...)
	toHash = append(toHash, common.LeftPadBytes(chainId.Bytes(), 32)...)

	msgHash := crypto.Keccak256(toHash)
	// convert to [32]byte
	var msgHash32 [32]byte
	copy(msgHash32[:], msgHash)

	// hash to G1
	return bn254util.MapToCurve(msgHash32)
}
