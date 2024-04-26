package service

import (
	"context"
	"encoding/hex"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/0glabs/0g-chain/helper/da/client"
	"github.com/0glabs/0g-chain/helper/da/types"
	"github.com/0glabs/0g-chain/helper/da/utils/sizedw8grp"

	jsoniter "github.com/json-iterator/go"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	defaultClientInstance = 10
)

var rpcClient client.DaLightRpcClient

func OnMessage(ctx context.Context, c *websocket.Conn, messageType websocket.MessageType, data []byte) {
	if messageType == websocket.TextMessage {
		rawMsg := unwrapJsonRpc(data)
		if verifyQuery(rawMsg) {
			eventStr := jsoniter.Get(rawMsg, "events").ToString()
			events := map[string][]string{}
			if err := jsoniter.UnmarshalFromString(eventStr, &events); err == nil {
				dasRequestMap := make(map[string]string, 4)
				for key, val := range events {
					if strings.HasPrefix(key, "das_request.") {
						dasRequestMap[strings.ReplaceAll(key, "das_request.", "")] = val[0]
					}
				}
				if len(dasRequestMap) == 4 {
					rid, _ := strconv.ParseUint(dasRequestMap["request_id"], 10, 64)
					numBlobs, _ := strconv.ParseUint(dasRequestMap["num_blobs"], 10, 64)
					req := types.DASRequest{
						RequestId:       rid,
						StreamId:        dasRequestMap["stream_id"],
						BatchHeaderHash: dasRequestMap["batch_header_hash"],
						NumBlobs:        numBlobs,
					}
					err := handleDasRequest(ctx, req)

					if err != nil {
						log.Err(err).Msgf("failed to handle das request: %v, %v", req, err)
					} else {
						log.Info().Msgf("successfully handled das request: %v", req)
					}
				}
			}
		}
	} else {
		// TODO: handle other message
	}
}

func OnClose() {
	if rpcClient != nil {
		rpcClient.Destroy()
		rpcClient = nil
	}
}

func unwrapJsonRpc(data []byte) []byte {
	result := jsoniter.Get(data, "result")
	if 0 < len(result.Keys()) {
		return []byte(result.ToString())
	}
	return []byte{}
}

func verifyQuery(data []byte) bool {
	if len(data) > 0 {
		return jsoniter.Get(data, "query").ToString() == "tm.event='Tx'"
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func handleDasRequest(ctx context.Context, request types.DASRequest) error {
	if rpcClient == nil {
		addrVal := ctx.Value(types.DA_RPC_ADDRESS)
		if addrVal == nil {
			return errors.New("da light service address not found in context")
		}

		limit := ctx.Value(types.INSTANCE_LIMIT)
		if limit == nil {
			limit = defaultClientInstance
		}

		rpcClient = client.NewDaLightClient(addrVal.(string), limit.(int))
	}

	streamID, err := hex.DecodeString(request.StreamId)
	if err != nil {
		return err
	}

	batchHeaderHash, err := hex.DecodeString(request.BatchHeaderHash)
	if err != nil {
		return err
	}

	result := make(chan bool, request.NumBlobs)
	taskCnt := min(rpcClient.GetInstanceCount(), int(request.NumBlobs))
	wg := sizedw8grp.New(taskCnt)

	for i := uint64(0); i < request.NumBlobs; i++ {
		wg.Add()
		go func(idx uint64) {
			defer wg.Done()
			ret, err := rpcClient.Sample(ctx, streamID, batchHeaderHash, uint32(idx), 1)
			if err != nil {
				log.Err(err).Msgf("failed to sample data availability with blob index %d", idx)
				result <- false
			} else {
				log.Info().Msgf("sample result for blob index %d: %v", idx, ret)
				result <- ret
			}
		}(i)
	}
	wg.Wait()
	close(result)

	finalResult := true
	for val := range result {
		if !val {
			finalResult = false
			break
		}
	}

	return runEvmosdCliReportDasResult(ctx, request.RequestId, finalResult)
}

func runEvmosdCliReportDasResult(ctx context.Context, requestId uint64, result bool) error {
	relativePath := ctx.Value(types.NODE_CLI_RELATIVE_PATH)
	if relativePath == nil {
		return errors.New("relativePath not found in context")
	}

	account := ctx.Value(types.NODE_CLI_EXEC_ACCOUNT)
	if account == nil {
		return errors.New("account not found in context")
	}

	args := []string{
		"tx",
		"das",
		"report-das-result",
		strconv.FormatUint(requestId, 10),
		strconv.FormatBool(result),
		"--from", account.(string),
		"--gas-prices", "7678500neuron", // TODO:  use args to set gas prices
	}

	homePath := ctx.Value(types.NODE_HOME_PATH)
	if len(homePath.(string)) > 0 {
		args = append(args, "--home", homePath.(string))
	}

	keyring := ctx.Value(types.NODE_CLI_EXEC_KEYRING)
	if len(keyring.(string)) > 0 {
		args = append(args, "--keyring-backend", keyring.(string))
	}

	cmdStr := relativePath.(string) + "0gchaind"
	cmd := exec.Command(cmdStr, append(args, "-y")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
