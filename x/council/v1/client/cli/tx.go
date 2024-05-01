package cli

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/0glabs/0g-chain/crypto/vrf"
	"github.com/0glabs/0g-chain/x/council/v1/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkkr "github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	vrfalgo "github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/cosmos/cosmos-sdk/client/input"
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
		NewRegisterCmd(),
		NewVoteCmd(),
	)
	return cmd
}

func NewRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a voter",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// bypass the restriction of set keyring options
			ctx := client.GetClientContextFromCmd(cmd).WithKeyringOptions(vrf.VrfOption())
			client.SetCmdClientContext(cmd, ctx)
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			kr := clientCtx.Keyring
			// get account name by address
			accAddr := clientCtx.GetFromAddress()
			accRecord, err := kr.KeyByAddress(accAddr)
			if err != nil {
				// not found record by address in keyring
				return nil
			}

			// check voter account record exists
			voterAccName := accRecord.Name + "-voter"
			_, err = kr.Key(voterAccName)
			if err == nil {
				// account exists, ask for user confirmation
				response, err2 := input.GetConfirmation(fmt.Sprintf("override the existing name %s", voterAccName), bufio.NewReader(clientCtx.Input), cmd.ErrOrStderr())
				if err2 != nil {
					return err2
				}

				if !response {
					return errors.New("aborted")
				}

				err2 = kr.Delete(voterAccName)
				if err2 != nil {
					return err2
				}
			}

			keyringAlgos, _ := kr.SupportedAlgorithms()
			algo, err := sdkkr.NewSigningAlgoFromString("vrf", keyringAlgos)
			if err != nil {
				return err
			}

			newRecord, err := kr.NewAccount(voterAccName, "", "", "", algo)
			if err != nil {
				return err
			}

			pubKey, err := newRecord.GetPubKey()
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromHex(hex.EncodeToString(clientCtx.GetFromAddress().Bytes()))
			if err != nil {
				return err
			}

			msg := &types.MsgRegister{
				Voter: valAddr.String(),
				Key:   pubKey.Bytes(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote council-id",
		Short: "Vote on a proposal",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			kr := clientCtx.Keyring

			// get account name by address
			inAddr := clientCtx.GetFromAddress()

			valAddr, err := sdk.ValAddressFromHex(hex.EncodeToString(inAddr.Bytes()))
			if err != nil {
				return err
			}

			inRecord, err := kr.KeyByAddress(inAddr)
			if err != nil {
				// not found record by address in keyring
				return nil
			}

			// check voter account record exists
			voterAccName := inRecord.Name + "-voter"
			voterRecord, err := kr.Key(voterAccName)
			if err != nil {
				// not found voter account
				return err
			}
			sk := vrfalgo.PrivateKey(voterRecord.GetLocal().PrivKey.Value)

			councilID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			votingStartHeight := types.DefaultVotingStartHeight + (councilID-1)*types.DefaultVotingPeriod

			rsp, err := stakingtypes.NewQueryClient(clientCtx).HistoricalInfo(cmd.Context(), &stakingtypes.QueryHistoricalInfoRequest{Height: int64(votingStartHeight)})
			if err != nil {
				return err
			}

			var tokens sdkmath.Int
			for _, val := range rsp.Hist.Valset {
				thisValAddr := val.GetOperator()

				if thisValAddr.Equals(valAddr) {
					tokens = val.GetTokens()
				}
			}
			// the denom of token is neuron, need to convert to A0GI
			a0gi := tokens.Quo(sdk.NewInt(1_000_000_000_000_000_000))
			// 1_000 0AGI token / vote
			numBallots := a0gi.Quo(sdk.NewInt(1_000)).Uint64()
			ballots := make([]*types.Ballot, numBallots)
			for i := range ballots {
				ballotID := uint64(i)
				ballots[i] = &types.Ballot{
					ID:      ballotID,
					Content: sk.Compute(bytes.Join([][]byte{rsp.Hist.Header.LastCommitHash, types.Uint64ToBytes(ballotID)}, nil)),
				}
			}

			msg := &types.MsgVote{
				CouncilID: councilID,
				Voter:     valAddr.String(),
				Ballots:   ballots,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
