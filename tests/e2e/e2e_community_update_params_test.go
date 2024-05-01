package e2e_test

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *IntegrationTestSuite) decodeTxMsgResponse(txRes *sdk.TxResponse, ptr codec.ProtoMarshaler) {
	// convert txRes.Data hex string to bytes
	txResBytes, err := hex.DecodeString(txRes.Data)
	suite.Require().NoError(err)

	// Unmarshal data to TxMsgData
	var txMsgData sdk.TxMsgData
	suite.Kava.EncodingConfig.Marshaler.MustUnmarshal(txResBytes, &txMsgData)
	suite.T().Logf("txData.MsgResponses: %v", txMsgData.MsgResponses)

	// Parse MsgResponse
	suite.Kava.EncodingConfig.Marshaler.MustUnmarshal(txMsgData.MsgResponses[0].Value, ptr)
	suite.Require().NoError(err)
}
