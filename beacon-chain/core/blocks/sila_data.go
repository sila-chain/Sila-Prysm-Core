package blocks

import (
	"bytes"
	"context"
	"errors"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// ProcessSilaDataInBlock is an operation performed on each
// beacon block to ensure the SILAEXEC data votes are processed
// into the beacon state.
//
// Official spec definition:
//
//	def process_sila_data(state: BeaconState, body: BeaconBlockBody) -> None:
//	 state.sila_data_votes.append(body.sila_data)
//	 if state.sila_data_votes.count(body.sila_data) * 2 > EPOCHS_PER_SilaExecution_VOTING_PERIOD * SLOTS_PER_EPOCH:
//	     state.sila_data = body.sila_data
func ProcessSilaDataInBlock(ctx context.Context, beaconState state.BeaconState, silaexecData *silapb.SilaData) (state.BeaconState, error) {
	_, span := trace.StartSpan(ctx, "blocks.ProcessSilaDataInBlock")
	defer span.End()

	if beaconState == nil || beaconState.IsNil() {
		return nil, errors.New("nil state")
	}
	if err := beaconState.AppendSilaDataVotes(silaexecData); err != nil {
		return nil, err
	}
	hasSupport, err := SilaDataHasEnoughSupport(beaconState, silaexecData)
	if err != nil {
		return nil, err
	}
	if hasSupport {
		if err := beaconState.SetSilaData(silaexecData); err != nil {
			return nil, err
		}
	}
	return beaconState, nil
}

// AreSilaDataEqual checks equality between two silaexec data objects.
func AreSilaDataEqual(a, b *silapb.SilaData) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.DepositCount == b.DepositCount &&
		bytes.Equal(a.BlockHash, b.BlockHash) &&
		bytes.Equal(a.DepositRoot, b.DepositRoot)
}

// SilaDataHasEnoughSupport returns true when the given silaData has more than 50% votes in the
// silaexec voting period. A vote is cast by including silaData in a block and part of state processing
// appends silaData to the state in the SilaDataVotes list. Iterating through this list checks the
// votes to see if they match the silaData.
func SilaDataHasEnoughSupport(beaconState state.ReadOnlyBeaconState, data *silapb.SilaData) (bool, error) {
	voteCount := uint64(0)

	for _, vote := range beaconState.SilaDataVotes() {
		if AreSilaDataEqual(vote, data) {
			voteCount++
		}
	}

	// If 50+% majority converged on the same silaData, then it has enough support to update the
	// state.
	support := params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSilaExecutionVotingPeriod))
	return voteCount*2 > uint64(support), nil
}
