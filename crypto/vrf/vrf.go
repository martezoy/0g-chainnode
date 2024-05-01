package vrf

import (
	"bytes"
	"crypto/subtle"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	vrfalgo "github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

const (
	// PrivKeySize defines the size of the PrivKey bytes
	PrivKeySize = 64
	// PubKeySize defines the size of the PubKey bytes
	PubKeySize = 32
	// KeyType is the string constant for the vrf algorithm
	KeyType = "vrf"
)

// Amino encoding names
const (
	// PrivKeyName defines the amino encoding name for the vrf private key
	PrivKeyName = "vrf/PrivKey"
	// PubKeyName defines the amino encoding name for the vrf public key
	PubKeyName = "vrf/PubKey"
)

// ----------------------------------------------------------------------------
// vrf Private Key

var (
	_ cryptotypes.PrivKey  = &PrivKey{}
	_ codec.AminoMarshaler = &PrivKey{}
)

// GenerateKey generates a new random private key. It returns an error upon
// failure.
func GenerateKey() (*PrivKey, error) {
	priv, err := vrfalgo.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	return &PrivKey{
		Key: priv,
	}, nil
}

func (privKey PrivKey) getVrfPrivateKey() vrfalgo.PrivateKey {
	return vrfalgo.PrivateKey(privKey.Key)
}

// Bytes returns the byte representation of the Private Key.
func (privKey PrivKey) Bytes() []byte {
	bz := make([]byte, len(privKey.Key))
	copy(bz, privKey.Key)

	return bz
}

// PubKey returns the private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey PrivKey) PubKey() cryptotypes.PubKey {
	pk, _ := vrfalgo.PrivateKey(privKey.Key).Public()

	return &PubKey{
		Key: pk,
	}
}

// Equals returns true if two private keys are equal and false otherwise.
func (privKey PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && subtle.ConstantTimeCompare(privKey.Bytes(), other.Bytes()) == 1
}

// Type returns vrf
func (privKey PrivKey) Type() string {
	return KeyType
}

// Compute generates the vrf value for the byte slice m using the
// underlying private key sk.
func (privKey PrivKey) Sign(digestBz []byte) ([]byte, error) {
	sk := privKey.getVrfPrivateKey()

	return sk.Compute(digestBz), nil
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PrivKeySize {
		return fmt.Errorf("invalid privkey size, expected %d got %d", PrivKeySize, len(bz))
	}
	privKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey PrivKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

// ----------------------------------------------------------------------------
// vrf Public Key

var (
	_ cryptotypes.PubKey   = &PubKey{}
	_ codec.AminoMarshaler = &PubKey{}
)

// func (pubKey PubKey) getVrfPublicKey() vrfalgo.PublicKey {
// 	return vrfalgo.PublicKey(pubKey.Key)
// }

// Address returns the address of the ECDSA public key.
// The function will return an empty address if the public key is invalid.
func (pubKey PubKey) Address() tmcrypto.Address {
	return tmcrypto.Address(common.BytesToAddress(pubKey.Key).Bytes())
}

// Bytes returns the raw bytes of the ECDSA public key.
func (pubKey PubKey) Bytes() []byte {
	bz := make([]byte, len(pubKey.Key))
	copy(bz, pubKey.Key)

	return bz
}

// String implements the fmt.Stringer interface.
func (pubKey PubKey) String() string {
	return fmt.Sprintf("vrf{%X}", pubKey.Key)
}

// Type returns vrf
func (pubKey PubKey) Type() string {
	return KeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// Verify returns true iff vrf=Compute(m) for the sk that
// corresponds to pk.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	panic("not implement")
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errorsmod.Wrapf(errortypes.ErrInvalidPubKey, "invalid pubkey size, expected %d, got %d", PubKeySize, len(bz))
	}
	pubKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey PubKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}
