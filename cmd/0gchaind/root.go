package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	ethermintclient "github.com/evmos/ethermint/client"
	"github.com/evmos/ethermint/crypto/hd"
	ethermintserver "github.com/evmos/ethermint/server"
	servercfg "github.com/evmos/ethermint/server/config"
	"github.com/spf13/cobra"
	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/0glabs/0g-chain/app"
	"github.com/0glabs/0g-chain/app/params"
	"github.com/0glabs/0g-chain/chaincfg"
	kavaclient "github.com/0glabs/0g-chain/client"
	"github.com/0glabs/0g-chain/cmd/opendb"
	"github.com/0glabs/0g-chain/crypto/vrf"
)

func customKeyringOptions() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = append(hd.SupportedAlgorithms, vrf.VrfAlgo)
		options.SupportedAlgosLedger = append(hd.SupportedAlgorithmsLedger, vrf.VrfAlgo)
	}
}

// NewRootCmd creates a new root command for the 0g-chain blockchain.
func NewRootCmd() *cobra.Command {
	encodingConfig := app.MakeEncodingConfig()
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(chaincfg.DefaultNodeHome).
		WithKeyringOptions(customKeyringOptions()).
		WithViper(chaincfg.EnvPrefix)

	rootCmd := &cobra.Command{
		Use:   chaincfg.AppName,
		Short: "Daemon and CLI for the 0g-chain blockchain.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err = client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := servercfg.AppConfig(chaincfg.BaseDenom)

			return server.InterceptConfigsPreRunHandler(
				cmd,
				customAppTemplate,
				customAppConfig,
				tmcfg.DefaultConfig(),
			)
		},
	}

	addSubCmds(rootCmd, encodingConfig, chaincfg.DefaultNodeHome)

	return rootCmd
}

// addSubCmds registers all the sub commands used by 0g-chain.
func addSubCmds(rootCmd *cobra.Command, encodingConfig params.EncodingConfig, defaultNodeHome string) {
	rootCmd.AddCommand(
		StatusCommand(),
		ethermintclient.ValidateChainID(
			genutilcli.InitCmd(app.ModuleBasics, defaultNodeHome),
		),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, defaultNodeHome),
		AssertInvariantsCmd(encodingConfig),
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, defaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(defaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true), // TODO add other shells, drop tmcli dependency, unhide?
		// testnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}), // TODO add
		debug.Cmd(),
		config.Cmd(),
	)

	ac := appCreator{
		encodingConfig: encodingConfig,
	}

	opts := ethermintserver.StartOptions{
		AppCreator:      ac.newApp,
		DefaultNodeHome: chaincfg.DefaultNodeHome,
		DBOpener:        opendb.OpenDB,
	}
	// ethermintserver adds additional flags to start the JSON-RPC server for evm support
	ethermintserver.AddCommands(
		rootCmd,
		opts,
		ac.appExport,
		ac.addStartCmdFlags,
	)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		newQueryCmd(),
		newTxCmd(),
		kavaclient.KeyCommands(chaincfg.DefaultNodeHome),
	)
}
