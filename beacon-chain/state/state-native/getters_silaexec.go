package state_native

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// SilaData corresponding to the proof-of-work chain information stored in the beacon state.
func (b *BeaconState) SilaData() *silapb.SilaData {
	if b.silaexecData == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.silaexecDataVal()
}

// silaexecDataVal corresponding to the proof-of-work chain information stored in the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) silaexecDataVal() *silapb.SilaData {
	if b.silaexecData == nil {
		return nil
	}

	return b.silaexecData.Copy()
}

// SilaDataVotes corresponds to votes from Sila on the canonical proof-of-work chain
// data retrieved from silaexec.
func (b *BeaconState) SilaDataVotes() []*silapb.SilaData {
	if b.silaDataVotes == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.silaDataVotesVal()
}

// silaDataVotesVal corresponds to votes from Sila on the canonical proof-of-work chain
// data retrieved from silaexec.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) silaDataVotesVal() []*silapb.SilaData {
	if b.silaDataVotes == nil {
		return nil
	}

	res := make([]*silapb.SilaData, len(b.silaDataVotes))
	for i := range res {
		res[i] = b.silaDataVotes[i].Copy()
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
