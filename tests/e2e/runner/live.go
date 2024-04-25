package runner

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/0glabs/0g-chain/client/grpc"
)

// LiveNodeRunnerConfig implements NodeRunner.
// It connects to a running network via the RPC, GRPC, and EVM urls.
type LiveNodeRunnerConfig struct {
	ZgChainRpcUrl    string
	ZgChainGrpcUrl   string
	ZgChainEvmRpcUrl string

	UpgradeHeight int64
}

// LiveNodeRunner implements NodeRunner for an already-running chain.
// If a LiveNodeRunner is used, end-to-end tests are run against a live chain.
type LiveNodeRunner struct {
	config LiveNodeRunnerConfig
}

var _ NodeRunner = LiveNodeRunner{}

// NewLiveNodeRunner creates a new LiveNodeRunner.
func NewLiveNodeRunner(config LiveNodeRunnerConfig) *LiveNodeRunner {
	return &LiveNodeRunner{config}
}

// StartChains implements NodeRunner.
// It initializes connections to the chain based on parameters.
// It attempts to ping the necessary endpoints and panics if they cannot be reached.
func (r LiveNodeRunner) StartChains() Chains {
	fmt.Println("establishing connection to live 0g-chain network")
	chains := NewChains()

	zgChain := ChainDetails{
		RpcUrl:    r.config.ZgChainRpcUrl,
		GrpcUrl:   r.config.ZgChainGrpcUrl,
		EvmRpcUrl: r.config.ZgChainEvmRpcUrl,
	}

	if err := waitForChainStart(zgChain); err != nil {
		panic(fmt.Sprintf("failed to ping chain: %s", err))
	}

	// determine chain id
	client, err := grpc.NewClient(zgChain.GrpcUrl)
	if err != nil {
		panic(fmt.Sprintf("failed to create 0gchain grpc client: %s", err))
	}

	nodeInfo, err := client.Query.Tm.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	if err != nil {
		panic(fmt.Sprintf("failed to fetch 0gchain node info: %s", err))
	}
	zgChain.ChainId = nodeInfo.DefaultNodeInfo.Network

	// determine staking denom
	stakingParams, err := client.Query.Staking.Params(context.Background(), &stakingtypes.QueryParamsRequest{})
	if err != nil {
		panic(fmt.Sprintf("failed to fetch 0gchain staking params: %s", err))
	}
	zgChain.StakingDenom = stakingParams.Params.BondDenom

	chains.Register("0gchain", &zgChain)

	fmt.Printf("successfully connected to live network %+v\n", zgChain)

	return chains
}

// Shutdown implements NodeRunner.
// As the chains are externally operated, this is a no-op.
func (LiveNodeRunner) Shutdown() {
	fmt.Println("shutting down e2e test connections.")
}
