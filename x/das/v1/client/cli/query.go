package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0glabs/0g-chain/x/das/v1/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for the inflation module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the das module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetNextRequestID(),
	)

	return cmd
}

func GetNextRequestID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-request-id",
		Short: "Query the next request ID",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryNextRequestIDRequest{}
			res, err := queryClient.NextRequestID(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%v\n", res.NextRequestID))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
