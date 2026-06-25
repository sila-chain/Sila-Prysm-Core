package mock

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// PoolMock is a fake implementation of PoolManager.
type PoolMock struct {
	Changes []*eth.SignedBLSToSilaChange
}

// PendingBLSToExecChanges --
func (m *PoolMock) PendingBLSToExecChanges() ([]*eth.SignedBLSToSilaChange, error) {
	return m.Changes, nil
}

// BLSToExecChangesForInclusion --
func (m *PoolMock) BLSToExecChangesForInclusion(_ state.ReadOnlyBeaconState) ([]*eth.SignedBLSToSilaChange, error) {
	return m.Changes, nil
}

// InsertBLSToExecChange --
func (m *PoolMock) InsertBLSToExecChange(change *eth.SignedBLSToSilaChange) {
	m.Changes = append(m.Changes, change)
}

// MarkIncluded --
func (*PoolMock) MarkIncluded(_ *eth.SignedBLSToSilaChange) {
	panic("implement me") // lint:nopanic -- mock / test code.
}

// ValidatorExists --
func (*PoolMock) ValidatorExists(_ primitives.ValidatorIndex) bool {
	panic("implement me") // lint:nopanic -- mock / test code.
}
