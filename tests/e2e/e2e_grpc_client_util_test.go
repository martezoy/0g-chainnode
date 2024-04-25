package e2e_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *IntegrationTestSuite) TestGrpcClientUtil_Account() {
	// ARRANGE
	// setup 0g account
	zgChainAcc := suite.ZgChain.NewFundedAccount("account-test", sdk.NewCoins(a0gi(big.NewInt(1e5))))

	// ACT
	rsp, err := suite.ZgChain.Grpc.BaseAccount(zgChainAcc.SdkAddress.String())

	// ASSERT
	suite.Require().NoError(err)
	suite.Equal(zgChainAcc.SdkAddress.String(), rsp.Address)
	suite.Greater(rsp.AccountNumber, uint64(1))
	suite.Equal(uint64(0), rsp.Sequence)
}
