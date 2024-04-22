package grpc_test

import (
	"testing"

	"github.com/0glabs/0g-chain/client/grpc"
	"github.com/stretchr/testify/require"
)

func TestNewClient_InvalidEndpoint(t *testing.T) {
	_, err := grpc.NewClient("invalid-url")
	require.ErrorContains(t, err, "unknown grpc url scheme")
	_, err = grpc.NewClient("")
	require.ErrorContains(t, err, "grpc url cannot be empty")
}
