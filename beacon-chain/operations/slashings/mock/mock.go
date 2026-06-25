package mock

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// PoolMock is a fake implementation of PoolManager.
type PoolMock struct {
	PendingAttSlashings  []silapb.AttSlashing
	PendingPropSlashings []*silapb.ProposerSlashing
}

// PendingAttesterSlashings --
func (m *PoolMock) PendingAttesterSlashings(_ context.Context, _ state.ReadOnlyBeaconState, _ bool) []silapb.AttSlashing {
	return m.PendingAttSlashings
}

// PendingProposerSlashings --
func (m *PoolMock) PendingProposerSlashings(_ context.Context, _ state.ReadOnlyBeaconState, _ bool) []*silapb.ProposerSlashing {
	return m.PendingPropSlashings
}

// InsertAttesterSlashing --
func (m *PoolMock) InsertAttesterSlashing(_ context.Context, _ state.ReadOnlyBeaconState, slashing silapb.AttSlashing) error {
	m.PendingAttSlashings = append(m.PendingAttSlashings, slashing)
	return nil
}

// InsertProposerSlashing --
func (m *PoolMock) InsertProposerSlashing(_ context.Context, _ state.ReadOnlyBeaconState, slashing *silapb.ProposerSlashing) error {
	m.PendingPropSlashings = append(m.PendingPropSlashings, slashing)
	return nil
}

// ConvertToElectra --
func (*PoolMock) ConvertToElectra() {}

// MarkIncludedAttesterSlashing --
func (*PoolMock) MarkIncludedAttesterSlashing(_ silapb.AttSlashing) {
	panic("implement me") // lint:nopanic -- Test / mock code.
}

// MarkIncludedProposerSlashing --
func (*PoolMock) MarkIncludedProposerSlashing(_ *silapb.ProposerSlashing) {
	panic("implement me") // lint:nopanic -- Test / mock code.
}
