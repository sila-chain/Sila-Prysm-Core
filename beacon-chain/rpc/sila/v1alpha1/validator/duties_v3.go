package validator

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/rpc/core"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAttesterDuties returns attester duties for the requested validators at the given epoch.
func (vs *Server) GetAttesterDuties(ctx context.Context, req *silapb.AttesterDutiesRequest) (*silapb.AttesterDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "validator.GetAttesterDuties")
	defer span.End()

	if vs.SyncChecker.Syncing() {
		return nil, status.Error(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}

	currentEpoch := slots.ToEpoch(vs.TimeFetcher.CurrentSlot())
	if req.Epoch > currentEpoch+1 {
		return nil, status.Errorf(codes.InvalidArgument, "Request epoch %d can not be greater than next epoch %d", req.Epoch, currentEpoch+1)
	}

	s, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	s, err = vs.stateForEpoch(ctx, s, req.Epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get state for epoch: %v", err)
	}

	duties, rpcErr := vs.CoreService.AttesterDuties(ctx, s, req.Epoch, req.ValidatorIndices)
	if rpcErr != nil {
		return nil, status.Errorf(core.ErrorReasonToGRPC(rpcErr.Reason), "%v", rpcErr.Err)
	}

	dependentRoot, err := vs.attestationDependentRoot(ctx, s, req.Epoch)
	if err != nil {
		return nil, err
	}

	optimistic, err := vs.OptimisticModeFetcher.IsOptimistic(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine optimistic status: %v", err)
	}

	dutiesResponses := make([]*silapb.AttesterDuty, len(duties))
	for i, d := range duties {
		dutiesResponses[i] = &silapb.AttesterDuty{
			Pubkey:                  d.Pubkey[:],
			ValidatorIndex:          d.ValidatorIndex,
			CommitteeIndex:          d.CommitteeIndex,
			CommitteeLength:         d.CommitteeLength,
			CommitteesAtSlot:        d.CommitteesAtSlot,
			ValidatorCommitteeIndex: d.ValidatorCommitteeIndex,
			Slot:                    d.Slot,
		}
	}

	return &silapb.AttesterDutiesResponse{
		DependentRoot:       dependentRoot,
		ExecutionOptimistic: optimistic,
		Duties:              dutiesResponses,
	}, nil
}

// GetProposerDutiesV2 returns proposer duties for the given epoch.
func (vs *Server) GetProposerDutiesV2(ctx context.Context, req *silapb.ProposerDutiesRequest) (*silapb.ProposerDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "validator.GetProposerDutiesV2")
	defer span.End()

	if vs.SyncChecker.Syncing() {
		return nil, status.Error(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}

	currentEpoch := slots.ToEpoch(vs.TimeFetcher.CurrentSlot())
	if req.Epoch > currentEpoch+1 {
		return nil, status.Errorf(codes.InvalidArgument, "Request epoch %d can not be greater than next epoch %d", req.Epoch, currentEpoch+1)
	}

	s, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	s, err = vs.stateForEpoch(ctx, s, req.Epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get state for epoch: %v", err)
	}

	duties, rpcErr := vs.CoreService.ProposerDuties(ctx, s, req.Epoch)
	if rpcErr != nil {
		return nil, status.Errorf(core.ErrorReasonToGRPC(rpcErr.Reason), "%v", rpcErr.Err)
	}

	dependentRoot, err := vs.proposalDependentRoot(ctx, s, req.Epoch)
	if err != nil {
		return nil, err
	}

	optimistic, err := vs.OptimisticModeFetcher.IsOptimistic(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine optimistic status: %v", err)
	}

	dutiesResponses := make([]*silapb.ProposerDutyV2, len(duties))
	for i, d := range duties {
		dutiesResponses[i] = &silapb.ProposerDutyV2{
			Pubkey:         d.Pubkey[:],
			ValidatorIndex: d.ValidatorIndex,
			Slot:           d.Slot,
		}
	}

	return &silapb.ProposerDutiesResponse{
		DependentRoot:       dependentRoot,
		ExecutionOptimistic: optimistic,
		Duties:              dutiesResponses,
	}, nil
}

// GetSyncCommitteeDuties returns sync committee duties for the requested validators at the given epoch.
func (vs *Server) GetSyncCommitteeDuties(ctx context.Context, req *silapb.SyncCommitteeDutiesRequest) (*silapb.SyncCommitteeDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "validator.GetSyncCommitteeDuties")
	defer span.End()

	if vs.SyncChecker.Syncing() {
		return nil, status.Error(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}

	currentEpoch := slots.ToEpoch(vs.TimeFetcher.CurrentSlot())
	lastValidEpoch := core.SyncCommitteeDutiesLastValidEpoch(currentEpoch)
	if req.Epoch > lastValidEpoch {
		return nil, status.Errorf(codes.InvalidArgument, "Request epoch %d can not be greater than last valid epoch %d for sync committee duties", req.Epoch, lastValidEpoch)
	}

	s, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	s, err = vs.stateForEpoch(ctx, s, req.Epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get state for epoch: %v", err)
	}

	duties, rpcErr := vs.CoreService.SyncCommitteeDuties(ctx, s, req.Epoch, currentEpoch, req.ValidatorIndices)
	if rpcErr != nil {
		return nil, status.Errorf(core.ErrorReasonToGRPC(rpcErr.Reason), "%v", rpcErr.Err)
	}

	optimistic, err := vs.OptimisticModeFetcher.IsOptimistic(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine optimistic status: %v", err)
	}

	dutiesResponses := make([]*silapb.SyncCommitteeDuty, len(duties))
	for i, d := range duties {
		dutiesResponses[i] = &silapb.SyncCommitteeDuty{
			Pubkey:                        d.Pubkey[:],
			ValidatorIndex:                d.ValidatorIndex,
			ValidatorSyncCommitteeIndices: d.ValidatorSyncCommitteeIndices,
		}
	}

	return &silapb.SyncCommitteeDutiesResponse{
		ExecutionOptimistic: optimistic,
		Duties:              dutiesResponses,
	}, nil
}

// GetPTCDuties returns payload timeliness committee duties for the requested validators at the given epoch.
func (vs *Server) GetPTCDuties(ctx context.Context, req *silapb.PTCDutiesRequest) (*silapb.PTCDutiesResponse, error) {
	ctx, span := trace.StartSpan(ctx, "validator.GetPTCDuties")
	defer span.End()

	if vs.SyncChecker.Syncing() {
		return nil, status.Error(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}

	if req.Epoch < params.BeaconConfig().GloasForkEpoch {
		return nil, status.Errorf(codes.InvalidArgument, "Request epoch %d is before Gloas fork epoch %d", req.Epoch, params.BeaconConfig().GloasForkEpoch)
	}

	currentEpoch := slots.ToEpoch(vs.TimeFetcher.CurrentSlot())
	nextEpoch := currentEpoch.Add(1)
	if req.Epoch > nextEpoch {
		return nil, status.Errorf(codes.InvalidArgument, "Request epoch %d can not be greater than next epoch %d", req.Epoch, nextEpoch)
	}

	s, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	s, err = vs.stateForEpoch(ctx, s, req.Epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get state for epoch: %v", err)
	}

	duties, rpcErr := vs.CoreService.PTCDuties(ctx, s, req.Epoch, req.ValidatorIndices)
	if rpcErr != nil {
		return nil, status.Errorf(core.ErrorReasonToGRPC(rpcErr.Reason), "%v", rpcErr.Err)
	}

	dependentRoot, err := vs.attestationDependentRoot(ctx, s, req.Epoch)
	if err != nil {
		return nil, err
	}

	optimistic, err := vs.OptimisticModeFetcher.IsOptimistic(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine optimistic status: %v", err)
	}

	dutiesResponses := make([]*silapb.PTCDuty, len(duties))
	for i, d := range duties {
		dutiesResponses[i] = &silapb.PTCDuty{
			Pubkey:         d.Pubkey[:],
			ValidatorIndex: d.ValidatorIndex,
			Slot:           d.Slot,
		}
	}

	return &silapb.PTCDutiesResponse{
		DependentRoot:       dependentRoot,
		ExecutionOptimistic: optimistic,
		Duties:              dutiesResponses,
	}, nil
}

// attestationDependentRoot returns the dependent root for attestation-style duties.
// For epochs <= 1 it returns the genesis block root; otherwise it computes the root from state.
func (vs *Server) attestationDependentRoot(ctx context.Context, s state.BeaconState, epoch primitives.Epoch) ([]byte, error) {
	if epoch <= 1 {
		r, err := vs.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not get genesis block root: %v", err)
		}
		return r[:], nil
	}
	root, err := core.AttestationDependentRoot(s, epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get dependent root: %v", err)
	}
	return root, nil
}

// proposalDependentRoot returns the dependent root for proposer duties.
// Epoch 0 always needs genesis root. Epoch 1 also needs it post-Fulu because
// V2 uses AttestationDependentRoot which requires epoch > 1.
func (vs *Server) proposalDependentRoot(ctx context.Context, s state.BeaconState, epoch primitives.Epoch) ([]byte, error) {
	if epoch == 0 || (epoch == 1 && s.Version() >= version.Fulu) {
		r, err := vs.BeaconDB.GenesisBlockRoot(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not get genesis block root: %v", err)
		}
		return r[:], nil
	}
	root, err := core.ProposalDependentRootV2(s, epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get dependent root: %v", err)
	}
	return root, nil
}
