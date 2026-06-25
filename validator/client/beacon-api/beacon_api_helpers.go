package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

var beaconAPITogRPCValidatorStatus = map[string]silapb.ValidatorStatus{
	"pending_initialized": silapb.ValidatorStatus_DEPOSITED,
	"pending_queued":      silapb.ValidatorStatus_PENDING,
	"active_ongoing":      silapb.ValidatorStatus_ACTIVE,
	"active_exiting":      silapb.ValidatorStatus_EXITING,
	"active_slashed":      silapb.ValidatorStatus_SLASHING,
	"exited_unslashed":    silapb.ValidatorStatus_EXITED,
	"exited_slashed":      silapb.ValidatorStatus_EXITED,
	"withdrawal_possible": silapb.ValidatorStatus_EXITED,
	"withdrawal_done":     silapb.ValidatorStatus_EXITED,
}

func (c *beaconApiValidatorClient) fork(ctx context.Context) (*structs.GetStateForkResponse, error) {
	const endpoint = "/sila/v1/beacon/states/head/fork"

	stateForkResponseJson := &structs.GetStateForkResponse{}

	if err := c.handler.Get(ctx, endpoint, stateForkResponseJson); err != nil {
		return nil, err
	}

	return stateForkResponseJson, nil
}

func (c *beaconApiValidatorClient) headers(ctx context.Context) (*structs.GetBlockHeadersResponse, error) {
	const endpoint = "/sila/v1/beacon/headers"

	blockHeadersResponseJson := &structs.GetBlockHeadersResponse{}

	if err := c.handler.Get(ctx, endpoint, blockHeadersResponseJson); err != nil {
		return nil, err
	}

	return blockHeadersResponseJson, nil
}

func (c *beaconApiValidatorClient) liveness(ctx context.Context, epoch primitives.Epoch, validatorIndexes []string) (*structs.GetLivenessResponse, error) {
	const endpoint = "/sila/v1/validator/liveness/"
	url := endpoint + strconv.FormatUint(uint64(epoch), 10)

	livenessResponseJson := &structs.GetLivenessResponse{}

	marshalledJsonValidatorIndexes, err := json.Marshal(validatorIndexes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal validator indexes")
	}

	if err = c.handler.Post(ctx, url, nil, bytes.NewBuffer(marshalledJsonValidatorIndexes), livenessResponseJson); err != nil {
		return nil, err
	}

	return livenessResponseJson, nil
}

func (c *beaconApiValidatorClient) syncing(ctx context.Context) (*structs.SyncStatusResponse, error) {
	const endpoint = "/sila/v1/node/syncing"

	syncingResponseJson := &structs.SyncStatusResponse{}

	if err := c.handler.Get(ctx, endpoint, syncingResponseJson); err != nil {
		return nil, err
	}

	return syncingResponseJson, nil
}

func (c *beaconApiValidatorClient) isSyncing(ctx context.Context) (bool, error) {
	response, err := c.syncing(ctx)
	if err != nil || response == nil || response.Data == nil {
		return true, errors.Wrapf(err, "failed to get syncing status")
	}

	return response.Data.IsSyncing, err
}

func (c *beaconApiValidatorClient) isOptimistic(ctx context.Context) (bool, error) {
	response, err := c.syncing(ctx)
	if err != nil || response == nil || response.Data == nil {
		return true, errors.Wrapf(err, "failed to get syncing status")
	}

	return response.Data.IsOptimistic, err
}
