package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/0glabs/0g-chain/x/das/v1/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		NewRequestDASCmd(),
		NewReportDASResultCmd(),
	)
	return cmd
}

func NewRequestDASCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-das steram-id batch-header-hash num-blobs",
		Short: "Request data-availability-sampling",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			numBlobs, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgRequestDAS(clientCtx.GetFromAddress(), args[0], args[1], uint32(numBlobs))
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd

}

func NewReportDASResultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report-das-result request-id results",
		Short: "Report data-availability-sampling result",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			requestID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			n := len(args) - 1
			results := make([]bool, n)
			for i := 0; i < n; i++ {
				var err error
				results[i], err = strconv.ParseBool(args[i+1])
				if err != nil {
					return err
				}
			}

			// get account name by address
			accAddr := clientCtx.GetFromAddress()

			samplerAddr, err := sdk.ValAddressFromHex(hex.EncodeToString(accAddr.Bytes()))
			if err != nil {
				return err
			}

			msg := &types.MsgReportDASResult{
				RequestID: requestID,
				Sampler:   samplerAddr.String(),
				Results:   results,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
