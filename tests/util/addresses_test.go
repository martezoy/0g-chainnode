package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0glabs/0g-chain/chaincfg"
	"github.com/0glabs/0g-chain/tests/util"
)

func TestAddressConversion(t *testing.T) {
	chaincfg.SetSDKConfig()
	bech32Addr := sdk.MustAccAddressFromBech32("0g17d2wax0zhjrrecvaszuyxdf5wcu5a0p4qlx3t5")
	hexAddr := common.HexToAddress("0xf354ee99e2bc863cE19d80b843353476394EbC35")
	require.Equal(t, bech32Addr, util.EvmToSdkAddress(hexAddr))
	require.Equal(t, hexAddr, util.SdkToEvmAddress(bech32Addr))
}
