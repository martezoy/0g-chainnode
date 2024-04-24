package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0glabs/0g-chain/x/committee/v1/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for the inflation module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the committee module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCurrentCommitteeID(),
		GetRegisteredVoters(),
	)

	return cmd
}

func GetCurrentCommitteeID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-committee-id",
		Short: "Query the current committee ID",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryCurrentCommitteeIDRequest{}
			res, err := queryClient.CurrentCommitteeID(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%v\n", res.CurrentCommitteeID))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetRegisteredVoters() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-voters",
		Short: "Query registered voters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryRegisteredVotersRequest{}
			res, err := queryClient.RegisteredVoters(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%v\n", strings.Join(res.Voters, ",")))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
