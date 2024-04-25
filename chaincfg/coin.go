package chaincfg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// Bip44CoinType satisfies EIP84. See https://github.com/ethereum/EIPs/issues/84 for more info.
	Bip44CoinType uint32 = 459 // TODO: need new coin type for 0g-chain (a0gi)
	// eth = 60
	// kava = 459 // see https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	// BIP44HDPath is the default BIP44 HD path used on Ethereum.
	//BIP44HDPath = ethaccounts.DefaultBaseDerivationPath.String()
)

// TODO: Implement BIP44CoinType and BIP44HDPath
// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func setBip44CoinType(config *sdk.Config) {
	config.SetCoinType(Bip44CoinType)
	//config.SetPurpose(sdk.Purpose)            // Shared
	//config.SetFullFundraiserPath(BIP44HDPath) //nolint: staticcheck
}
