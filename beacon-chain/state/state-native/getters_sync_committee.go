package state_native

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// CurrentSyncCommittee of the current sync committee in beacon chain state.
func (b *BeaconState) CurrentSyncCommittee() (*silapb.SyncCommittee, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.version == version.Phase0 {
		return nil, errNotSupported("CurrentSyncCommittee", b.version)
	}

	if b.currentSyncCommittee == nil {
		return nil, nil
	}

	return b.currentSyncCommitteeVal(), nil
}

// currentSyncCommitteeVal of the current sync committee in beacon chain state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) currentSyncCommitteeVal() *silapb.SyncCommittee {
	return copySyncCommittee(b.currentSyncCommittee)
}

// NextSyncCommittee of the next sync committee in beacon chain state.
func (b *BeaconState) NextSyncCommittee() (*silapb.SyncCommittee, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.version == version.Phase0 {
		return nil, errNotSupported("NextSyncCommittee", b.version)
	}

	if b.nextSyncCommittee == nil {
		return nil, nil
	}

	return b.nextSyncCommitteeVal(), nil
}

// nextSyncCommitteeVal of the next sync committee in beacon chain state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) nextSyncCommitteeVal() *silapb.SyncCommittee {
	return copySyncCommittee(b.nextSyncCommittee)
}

// copySyncCommittee copies the provided sync committee object.
func copySyncCommittee(data *silapb.SyncCommittee) *silapb.SyncCommittee {
	if data == nil {
		return nil
	}
	return &silapb.SyncCommittee{
		Pubkeys:         bytesutil.SafeCopy2dBytes(data.Pubkeys),
		AggregatePubkey: bytesutil.SafeCopyBytes(data.AggregatePubkey),
	}
}
