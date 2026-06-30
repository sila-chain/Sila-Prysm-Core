package grpc_api

import (
	"context"
	"fmt"
	"sort"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/rpc/silaapi/helpers"
	statenative "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/validator"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	silaapi "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaapi/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/iface"
	validatorHelpers "github.com/sila-chain/Sila-Consensus-Core/v7/validator/helpers"
)

type grpcSilaChainClient struct {
	chainClient iface.ChainClient
}

func (g grpcSilaChainClient) ValidatorCount(ctx context.Context, _ string, statuses []validator.Status) ([]iface.ValidatorCount, error) {
	resp, err := g.chainClient.Validators(ctx, &silapb.ListValidatorsRequest{PageSize: 0})
	if err != nil {
		return nil, errors.Wrap(err, "list validators failed")
	}

	var vals []*silapb.Validator
	for _, val := range resp.ValidatorList {
		vals = append(vals, val.Validator)
	}

	head, err := g.chainClient.ChainHead(ctx, &empty.Empty{})
	if err != nil {
		return nil, errors.Wrap(err, "get chain head")
	}

	if len(statuses) == 0 {
		for _, val := range silaapi.ValidatorStatus_value {
			statuses = append(statuses, validator.Status(val))
		}
	}

	valCount, err := validatorCountByStatus(vals, statuses, head.HeadEpoch)
	if err != nil {
		return nil, errors.Wrap(err, "validator count by status")
	}

	return valCount, nil
}

// validatorCountByStatus returns a slice of validator count for each status in the given epoch.
func validatorCountByStatus(validators []*silapb.Validator, statuses []validator.Status, epoch primitives.Epoch) ([]iface.ValidatorCount, error) {
	countByStatus := make(map[validator.Status]uint64)
	for _, val := range validators {
		readOnlyVal, err := statenative.NewValidator(val)
		if err != nil {
			return nil, fmt.Errorf("could not convert validator: %w", err)
		}
		valStatus, err := helpers.ValidatorStatus(readOnlyVal, epoch)
		if err != nil {
			return nil, fmt.Errorf("could not get validator status: %w", err)
		}
		valSubStatus, err := helpers.ValidatorSubStatus(readOnlyVal, epoch)
		if err != nil {
			return nil, fmt.Errorf("could not get validator sub status: %w", err)
		}

		for _, status := range statuses {
			if valStatus == status || valSubStatus == status {
				countByStatus[status]++
			}
		}
	}

	var resp []iface.ValidatorCount
	for status, count := range countByStatus {
		resp = append(resp, iface.ValidatorCount{
			Status: status.String(),
			Count:  count,
		})
	}

	// Sort the response slice according to status strings for deterministic ordering of validator count response.
	sort.Slice(resp, func(i, j int) bool {
		return resp[i].Status < resp[j].Status
	})

	return resp, nil
}

func (c *grpcSilaChainClient) ValidatorPerformance(ctx context.Context, in *silapb.ValidatorPerformanceRequest) (*silapb.ValidatorPerformanceResponse, error) {
	return c.chainClient.ValidatorPerformance(ctx, in)
}

// NewGrpcSilaChainClient creates a new gRPC Sila chain client that supports
// dynamic connection switching via the NodeConnection's GrpcConnectionProvider.
func NewGrpcSilaChainClient(conn validatorHelpers.NodeConnection) iface.SilaChainClient {
	return &grpcSilaChainClient{chainClient: NewGrpcChainClient(conn)}
}
