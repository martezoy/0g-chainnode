package query

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	committeetypes "github.com/0glabs/0g-chain/x/committee/v1/types"
	dastypes "github.com/0glabs/0g-chain/x/das/v1/types"
	evmutiltypes "github.com/0glabs/0g-chain/x/evmutil/types"
)

// QueryClient is a wrapper with all Cosmos and 0gChain grpc query clients
type QueryClient struct {
	// cosmos-sdk query clients

	Tm           tmservice.ServiceClient
	Tx           txtypes.ServiceClient
	Auth         authtypes.QueryClient
	Authz        authz.QueryClient
	Bank         banktypes.QueryClient
	Distribution disttypes.QueryClient
	Evidence     evidencetypes.QueryClient
	Gov          govv1types.QueryClient
	GovBeta      govv1beta1types.QueryClient
	Mint         minttypes.QueryClient
	Params       paramstypes.QueryClient
	Slashing     slashingtypes.QueryClient
	Staking      stakingtypes.QueryClient
	Upgrade      upgradetypes.QueryClient
	Consensus    consensustypes.QueryClient

	// 3rd party query clients

	Evm         evmtypes.QueryClient
	Feemarket   feemarkettypes.QueryClient
	IbcClient   ibcclienttypes.QueryClient
	IbcTransfer ibctransfertypes.QueryClient

	// 0g-chain module query clients

	Committee committeetypes.QueryClient
	Das       dastypes.QueryClient
	Evmutil   evmutiltypes.QueryClient
}

// NewQueryClient creates a new QueryClient and initializes all the module query clients
func NewQueryClient(grpcEndpoint string) (*QueryClient, error) {
	conn, err := newGrpcConnection(context.Background(), grpcEndpoint)
	if err != nil {
		return &QueryClient{}, err
	}
	client := &QueryClient{
		Tm:           tmservice.NewServiceClient(conn),
		Tx:           txtypes.NewServiceClient(conn),
		Auth:         authtypes.NewQueryClient(conn),
		Authz:        authz.NewQueryClient(conn),
		Bank:         banktypes.NewQueryClient(conn),
		Distribution: disttypes.NewQueryClient(conn),
		Evidence:     evidencetypes.NewQueryClient(conn),
		Gov:          govv1types.NewQueryClient(conn),
		GovBeta:      govv1beta1types.NewQueryClient(conn),
		Mint:         minttypes.NewQueryClient(conn),
		Params:       paramstypes.NewQueryClient(conn),
		Slashing:     slashingtypes.NewQueryClient(conn),
		Staking:      stakingtypes.NewQueryClient(conn),
		Upgrade:      upgradetypes.NewQueryClient(conn),
		Consensus:    consensustypes.NewQueryClient(conn),

		Evm:         evmtypes.NewQueryClient(conn),
		Feemarket:   feemarkettypes.NewQueryClient(conn),
		IbcClient:   ibcclienttypes.NewQueryClient(conn),
		IbcTransfer: ibctransfertypes.NewQueryClient(conn),

		Committee: committeetypes.NewQueryClient(conn),
		Das:       dastypes.NewQueryClient(conn),
		Evmutil:   evmutiltypes.NewQueryClient(conn),
	}
	return client, nil
}
