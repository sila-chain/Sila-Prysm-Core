package state_native

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// PreviousEpochAttestations corresponding to blocks on the beacon chain.
func (b *BeaconState) PreviousEpochAttestations() ([]*silapb.PendingAttestation, error) {
	if b.version != version.Phase0 {
		return nil, errNotSupported("PreviousEpochAttestations", b.version)
	}

	if b.previousEpochAttestations == nil {
		return nil, nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.previousEpochAttestationsVal(), nil
}

// previousEpochAttestationsVal corresponding to blocks on the beacon chain.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) previousEpochAttestationsVal() []*silapb.PendingAttestation {
	if b.previousEpochAttestations == nil {
		return nil
	}

	res := make([]*silapb.PendingAttestation, len(b.previousEpochAttestations))
	for i := range res {
		res[i] = b.previousEpochAttestations[i].Copy()
	}
	return res
}

// CurrentEpochAttestations corresponding to blocks on the beacon chain.
func (b *BeaconState) CurrentEpochAttestations() ([]*silapb.PendingAttestation, error) {
	if b.version != version.Phase0 {
		return nil, errNotSupported("CurrentEpochAttestations", b.version)
	}

	if b.currentEpochAttestations == nil {
		return nil, nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.currentEpochAttestationsVal(), nil
}

// currentEpochAttestations corresponding to blocks on the beacon chain.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) currentEpochAttestationsVal() []*silapb.PendingAttestation {
	if b.currentEpochAttestations == nil {
		return nil
	}

	res := make([]*silapb.PendingAttestation, len(b.currentEpochAttestations))
	for i := range res {
		res[i] = b.currentEpochAttestations[i].Copy()
	}
	return res
}
