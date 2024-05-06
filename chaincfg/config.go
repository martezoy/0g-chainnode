package chaincfg

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	AppName   = "0gchaind"
	EnvPrefix = "0GCHAIN"
)

func SetSDKConfig() *sdk.Config {
	config := sdk.GetConfig()
	setBech32Prefixes(config)
	setBip44CoinType(config)
	return config
}
