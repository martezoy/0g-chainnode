package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0glabs/0g-chain/x/dasigners/v1/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for the inflation module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the dasigners module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetEpochNumber(),
	)

	return cmd
}

func GetEpochNumber() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-number",
		Short: "Query current epoch number",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryEpochNumberRequest{}
			res, err := queryClient.EpochNumber(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%v\n", res.EpochNumber))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
