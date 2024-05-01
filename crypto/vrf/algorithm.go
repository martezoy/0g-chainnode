package vrf

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var (
	// SupportedAlgorithms defines the list of signing algorithms used on Evmos:
	//  - eth_secp256k1 (Ethereum)
	//  - secp256k1 (Tendermint)
	SupportedAlgorithms = keyring.SigningAlgoList{VrfAlgo}
	// SupportedAlgorithmsLedger defines the list of signing algorithms used on Evmos for the Ledger device:
	//  - eth_secp256k1 (Ethereum)
	//  - secp256k1 (Tendermint)
	SupportedAlgorithmsLedger = keyring.SigningAlgoList{VrfAlgo}
)

func VrfOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = SupportedAlgorithms
		options.SupportedAlgosLedger = SupportedAlgorithmsLedger
	}
}

const (
	VrfType = hd.PubKeyType(KeyType)
)

var (
	_       keyring.SignatureAlgo = VrfAlgo
	VrfAlgo                       = vrfAlgo{}
)

type vrfAlgo struct{}

func (s vrfAlgo) Name() hd.PubKeyType {
	return VrfType
}

func (s vrfAlgo) Derive() hd.DeriveFn {
	return func(mnemonic, bip39Passphrase, path string) ([]byte, error) {
		key, err := GenerateKey()
		if err != nil {
			return nil, err
		}

		return key.Bytes(), nil
	}
}

func (s vrfAlgo) Generate() hd.GenerateFn {
	return func(bz []byte) cryptotypes.PrivKey {
		key, _ := GenerateKey()
		return key
	}
}
