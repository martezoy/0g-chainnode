package types

import (
	"encoding/hex"
	fmt "fmt"

	"github.com/0glabs/0g-chain/crypto/bn254util"
	"github.com/consensys/gnark-crypto/ecc/bn254"
)

func ValidateHexAddress(account string) error {
	addr, err := hex.DecodeString(account)
	if err != nil {
		return err
	}
	if len(addr) != 20 {
		return fmt.Errorf("invalid address length")
	}
	return nil
}

func (s *Signer) Validate() error {
	if len(s.PubkeyG1) != bn254util.G1PointSize {
		return fmt.Errorf("invalid G1 pubkey length")
	}
	if len(s.PubkeyG2) != bn254util.G2PointSize {
		return fmt.Errorf("invalid G2 pubkey length")
	}
	if err := ValidateHexAddress(s.Account); err != nil {
		return err
	}
	return nil
}

func (s *Signer) ValidateSignature(hash *bn254.G1Affine, signature *bn254.G1Affine) bool {
	pubkeyG1 := bn254util.DeserializeG1(s.PubkeyG1)
	pubkeyG2 := bn254util.DeserializeG2(s.PubkeyG2)
	gamma := bn254util.Gamma(hash, signature, pubkeyG1, pubkeyG2)

	// pairing
	P := [2]bn254.G1Affine{
		*new(bn254.G1Affine).Add(signature, new(bn254.G1Affine).ScalarMultiplication(pubkeyG1, gamma)),
		*new(bn254.G1Affine).Add(hash, new(bn254.G1Affine).ScalarMultiplication(bn254util.GetG1Generator(), gamma)),
	}
	Q := [2]bn254.G2Affine{
		*new(bn254.G2Affine).Neg(bn254util.GetG2Generator()),
		*pubkeyG2,
	}

	ok, err := bn254.PairingCheck(P[:], Q[:])
	if err != nil {
		return false
	}
	return ok
}
