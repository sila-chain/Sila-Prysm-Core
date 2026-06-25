package grpc_api

import (
	"context"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/iface"
	validatorHelpers "github.com/sila-chain/Sila-Consensus-Core/v7/validator/helpers"
	"github.com/golang/protobuf/ptypes/empty"
)

var (
	_ = iface.NodeClient(&grpcNodeClient{})
)

type grpcNodeClient struct {
	*grpcClientManager[silapb.NodeClient]
}

func (c *grpcNodeClient) SyncStatus(ctx context.Context, in *empty.Empty) (*silapb.SyncStatus, error) {
	return c.getClient().GetSyncStatus(ctx, in)
}

func (c *grpcNodeClient) Genesis(ctx context.Context, in *empty.Empty) (*silapb.Genesis, error) {
	return c.getClient().GetGenesis(ctx, in)
}

func (c *grpcNodeClient) Version(ctx context.Context, in *empty.Empty) (*silapb.Version, error) {
	return c.getClient().GetVersion(ctx, in)
}

func (c *grpcNodeClient) Peers(ctx context.Context, in *empty.Empty) (*silapb.Peers, error) {
	return c.getClient().ListPeers(ctx, in)
}

func (c *grpcNodeClient) IsReady(ctx context.Context) bool {
	// GetHealth returns 200 OK only if node is synced and not optimistic.
	// otherwise it will throw an error
	_, err := c.getClient().GetHealth(ctx, &silapb.HealthRequest{})
	if err != nil {
		log.WithError(err).WithField("url", c.conn.GetGrpcConnectionProvider().CurrentHost()).Debug("Node is not ready")
		return false
	}
	return true
}

// NewNodeClient creates a new gRPC node client that supports
// dynamic connection switching via the NodeConnection's GrpcConnectionProvider.
func NewNodeClient(conn validatorHelpers.NodeConnection) iface.NodeClient {
	return &grpcNodeClient{
		grpcClientManager: newGrpcClientManager(conn, silapb.NewNodeClient),
	}
}
