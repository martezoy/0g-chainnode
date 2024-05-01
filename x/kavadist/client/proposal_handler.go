package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/0glabs/0g-chain/x/kavadist/client/cli"
)

// community-pool multi-spend proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal)
)
