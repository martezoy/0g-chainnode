package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/0glabs/0g-chain/chaincfg"
)

func main() {
	chaincfg.SetSDKConfig().Seal()

	rootCmd := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, chaincfg.EnvPrefix, chaincfg.DefaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}
