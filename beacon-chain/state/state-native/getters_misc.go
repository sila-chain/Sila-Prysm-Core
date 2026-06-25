package state_native

import (
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// Id is the identifier of the beacon state.
func (b *BeaconState) Id() uint64 {
	return b.id
}

// GenesisTime of the beacon state as a time.Time.
func (b *BeaconState) GenesisTime() time.Time {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return time.Unix(int64(b.genesisTime), 0)
}

// GenesisValidatorsRoot of the beacon state.
func (b *BeaconState) GenesisValidatorsRoot() []byte {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.genesisValidatorsRoot[:]
}

// Version of the beacon state. This method
// is strictly meant to be used without a lock
// internally.
func (b *BeaconState) Version() int {
	return b.version
}

// Slot of the current beacon chain state.
func (b *BeaconState) Slot() primitives.Slot {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.slot
}

// Fork version of the beacon chain.
func (b *BeaconState) Fork() *silapb.Fork {
	if b.fork == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.forkVal()
}

// forkVal version of the beacon chain.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) forkVal() *silapb.Fork {
	if b.fork == nil {
		return nil
	}

	prevVersion := make([]byte, len(b.fork.PreviousVersion))
	copy(prevVersion, b.fork.PreviousVersion)
	currVersion := make([]byte, len(b.fork.CurrentVersion))
	copy(currVersion, b.fork.CurrentVersion)
	return &silapb.Fork{
		PreviousVersion: prevVersion,
		CurrentVersion:  currVersion,
		Epoch:           b.fork.Epoch,
	}
}

// HistoricalRoots based on epochs stored in the beacon state.
func (b *BeaconState) HistoricalRoots() [][]byte {
	if b.historicalRoots == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.historicalRoots.Slice()
}

// HistoricalSummaries of the beacon state.
func (b *BeaconState) HistoricalSummaries() ([]*silapb.HistoricalSummary, error) {
	if b.version < version.Capella {
		return nil, errNotSupported("HistoricalSummaries", b.version)
	}

	if b.historicalSummaries == nil {
		return nil, nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.historicalSummariesVal(), nil
}

// historicalSummariesVal of the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) historicalSummariesVal() []*silapb.HistoricalSummary {
	return silapb.CopySlice(b.historicalSummaries)
}
