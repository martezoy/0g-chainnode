package types

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "da-signers"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// QuerierRoute Top level query string
	QuerierRoute = "dasigners"
)

var (
	// prefix
	SignerKeyPrefix       = []byte{0x00}
	EpochQuorumsKeyPrefix = []byte{0x01}
	RegistrationKeyPrefix = []byte{0x02}
	QuorumCountKeyPrefix  = []byte{0x03}

	// keys
	ParamsKey      = []byte{0x05}
	EpochNumberKey = []byte{0x06}
)

func GetSignerKeyFromAccount(account string) ([]byte, error) {
	return hex.DecodeString(account)
}

func GetEpochQuorumKey(epoch uint64, quorumId uint64) []byte {
	b := sdk.Uint64ToBigEndian(epoch)
	b = append(b, sdk.Uint64ToBigEndian(quorumId)...)
	return b
}

func GetQuorumCountKey(epoch uint64) []byte {
	return sdk.Uint64ToBigEndian(epoch)
}

func GetEpochRegistrationKeyPrefix(epoch uint64) []byte {
	return append(RegistrationKeyPrefix, sdk.Uint64ToBigEndian(epoch)...)
}

func GetRegistrationKey(account string) ([]byte, error) {
	return hex.DecodeString(account)
}
