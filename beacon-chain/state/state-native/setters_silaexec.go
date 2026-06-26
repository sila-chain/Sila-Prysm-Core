package state_native

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stateutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// SetSilaData for the beacon state.
func (b *BeaconState) SetSilaData(val *silapb.SilaData) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.silaexecData = val
	b.markFieldAsDirty(types.SilaData)
	return nil
}

// SetSilaDataVotes for the beacon state. Updates the entire
// list to a new value by overwriting the previous one.
func (b *BeaconState) SetSilaDataVotes(val []*silapb.SilaData) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.sharedFieldReferences[types.SilaDataVotes].MinusRef()
	b.sharedFieldReferences[types.SilaDataVotes] = stateutil.NewRef(1)

	b.silaDataVotes = val
	b.markFieldAsDirty(types.SilaDataVotes)
	b.rebuildTrie[types.SilaDataVotes] = true
	return nil
}

// SetSilaExecutionDepositIndex for the beacon state.
func (b *BeaconState) SetSilaExecutionDepositIndex(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.silaExecutionDepositIndex = val
	b.markFieldAsDirty(types.SilaExecutionDepositIndex)
	return nil
}

// AppendSilaDataVotes for the beacon state. Appends the new value
// to the end of list.
func (b *BeaconState) AppendSilaDataVotes(val *silapb.SilaData) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	votes := b.silaDataVotes
	if b.sharedFieldReferences[types.SilaDataVotes].Refs() > 1 {
		// Copy elements in underlying array by reference.
		votes = make([]*silapb.SilaData, 0, len(b.silaDataVotes)+1)
		votes = append(votes, b.silaDataVotes...)
		b.sharedFieldReferences[types.SilaDataVotes].MinusRef()
		b.sharedFieldReferences[types.SilaDataVotes] = stateutil.NewRef(1)
	}

	b.silaDataVotes = append(votes, val)
	b.markFieldAsDirty(types.SilaDataVotes)
	b.addDirtyIndices(types.SilaDataVotes, []uint64{uint64(len(b.silaDataVotes) - 1)})
	return nil
}
