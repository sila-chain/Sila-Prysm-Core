package grpc_api

import (
	"context"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/iface"
	validatorHelpers "github.com/sila-chain/Sila-Consensus-Core/v7/validator/helpers"
	"github.com/golang/protobuf/ptypes/empty"
)

type grpcChainClient struct {
	*grpcClientManager[silapb.BeaconChainClient]
}

func (c *grpcChainClient) ChainHead(ctx context.Context, in *empty.Empty) (*silapb.ChainHead, error) {
	return c.getClient().GetChainHead(ctx, in)
}

func (c *grpcChainClient) ValidatorBalances(ctx context.Context, in *silapb.ListValidatorBalancesRequest) (*silapb.ValidatorBalances, error) {
	return c.getClient().ListValidatorBalances(ctx, in)
}

func (c *grpcChainClient) Validators(ctx context.Context, in *silapb.ListValidatorsRequest) (*silapb.Validators, error) {
	return c.getClient().ListValidators(ctx, in)
}

func (c *grpcChainClient) ValidatorQueue(ctx context.Context, in *empty.Empty) (*silapb.ValidatorQueue, error) {
	return c.getClient().GetValidatorQueue(ctx, in)
}

func (c *grpcChainClient) ValidatorPerformance(ctx context.Context, in *silapb.ValidatorPerformanceRequest) (*silapb.ValidatorPerformanceResponse, error) {
	return c.getClient().GetValidatorPerformance(ctx, in)
}

func (c *grpcChainClient) ValidatorParticipation(ctx context.Context, in *silapb.GetValidatorParticipationRequest) (*silapb.ValidatorParticipationResponse, error) {
	return c.getClient().GetValidatorParticipation(ctx, in)
}

// NewGrpcChainClient creates a new gRPC chain client that supports
// dynamic connection switching via the NodeConnection's GrpcConnectionProvider.
func NewGrpcChainClient(conn validatorHelpers.NodeConnection) iface.ChainClient {
	return &grpcChainClient{
		grpcClientManager: newGrpcClientManager(conn, silapb.NewBeaconChainClient),
	}
}
