package client

import (
	"context"
	"time"

	"github.com/0glabs/0g-chain/helper/da/light"

	"github.com/pkg/errors"
)

type DaLightRpcClient interface {
	Sample(ctx context.Context, streamId, headerHash []byte, blobIdx, times uint32) (bool, error)
	Destroy()
	GetInstanceCount() int
}

type daLightClient struct {
	maxInstance int
	pool        ConnectionPool
}

func NewDaLightClient(address string, instanceLimit int) DaLightRpcClient {
	return &daLightClient{
		maxInstance: instanceLimit,
		pool:        NewConnectionPool(address, instanceLimit, 10*time.Minute),
	}
}

func (c *daLightClient) Sample(ctx context.Context, streamId, headerHash []byte, blobIdx, times uint32) (bool, error) {
	connection, err := c.pool.GetConnection()
	if err != nil {
		return false, errors.Wrap(err, "failed to connect to da light server")
	}
	defer c.pool.ReleaseConnection(connection)

	req := &light.SampleRequest{
		StreamId:        streamId,
		BatchHeaderHash: headerHash,
		BlobIndex:       blobIdx,
		Times:           times,
	}
	client := light.NewLightClient(connection)
	reply, err := client.Sample(ctx, req)
	if err != nil {
		return false, errors.Wrap(err, "failed to sample from da light server")
	}

	return reply.Success, nil
}

func (c *daLightClient) Destroy() {
	if c.pool != nil {
		c.pool.Close()
		c.pool = nil
	}
}

func (c *daLightClient) GetInstanceCount() int {
	return c.maxInstance
}
