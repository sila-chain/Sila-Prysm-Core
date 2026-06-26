package state_native

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// SilaExecutionData corresponding to the proof-of-work chain information stored in the beacon state.
func (b *BeaconState) SilaExecutionData() *silapb.SilaExecutionData {
	if b.silaexecData == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.silaexecDataVal()
}

// silaexecDataVal corresponding to the proof-of-work chain information stored in the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) silaexecDataVal() *silapb.SilaExecutionData {
	if b.silaexecData == nil {
		return nil
	}

	return b.silaexecData.Copy()
}

// SilaExecutionDataVotes corresponds to votes from Sila on the canonical proof-of-work chain
// data retrieved from silaexec.
func (b *BeaconState) SilaExecutionDataVotes() []*silapb.SilaExecutionData {
	if b.silaExecutionDataVotes == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.silaExecutionDataVotesVal()
}

// silaExecutionDataVotesVal corresponds to votes from Sila on the canonical proof-of-work chain
// data retrieved from silaexec.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) silaExecutionDataVotesVal() []*silapb.SilaExecutionData {
	if b.silaExecutionDataVotes == nil {
		return nil
	}

	res := make([]*silapb.SilaExecutionData, len(b.silaExecutionDataVotes))
	for i := range res {
		res[i] = b.silaExecutionDataVotes[i].Copy()
	}
	return res
}

// SilaExecutionDepositIndex corresponds to the index of the deposit made to the
// validator sila deposit at the time of this state's silaexec data.
func (b *BeaconState) SilaExecutionDepositIndex() uint64 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.silaExecutionDepositIndex
}
