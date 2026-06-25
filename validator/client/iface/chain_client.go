package iface

import (
	"context"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/golang/protobuf/ptypes/empty"
)

type ChainClient interface {
	ChainHead(ctx context.Context, in *empty.Empty) (*silapb.ChainHead, error)
	ValidatorBalances(ctx context.Context, in *silapb.ListValidatorBalancesRequest) (*silapb.ValidatorBalances, error)
	Validators(ctx context.Context, in *silapb.ListValidatorsRequest) (*silapb.Validators, error)
	ValidatorQueue(ctx context.Context, in *empty.Empty) (*silapb.ValidatorQueue, error)
	ValidatorParticipation(ctx context.Context, in *silapb.GetValidatorParticipationRequest) (*silapb.ValidatorParticipationResponse, error)
	ValidatorPerformance(context.Context, *silapb.ValidatorPerformanceRequest) (*silapb.ValidatorPerformanceResponse, error)
}
