package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "council"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName
)

// Key prefixes
var (
	CouncilKeyPrefix = []byte{0x00} // prefix for keys that store councils
	VoteKeyPrefix    = []byte{0x01} // prefix for keys that store votes
	VoterKeyPrefix   = []byte{0x02} // prefix for keys that store voters

	ParamsKey            = []byte{0x03}
	VotingStartHeightKey = []byte{0x04}
	VotingPeriodKey      = []byte{0x05}
	CurrentCouncilIDKey  = []byte{0x06}
)

// GetKeyFromID returns the bytes to use as a key for a uint64 id
func GetKeyFromID(id uint64) []byte {
	return Uint64ToBytes(id)
}

func GetVoteKey(councilID uint64, voter sdk.ValAddress) []byte {
	return append(GetKeyFromID(councilID), voter.Bytes()...)
}

func GetVoterKey(voter sdk.ValAddress) []byte {
	return voter.Bytes()
}

// Uint64ToBytes converts a uint64 into fixed length bytes for use in store keys.
func Uint64ToBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(id))
	return bz
}

// Uint64FromBytes converts some fixed length bytes back into a uint64.
func Uint64FromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
