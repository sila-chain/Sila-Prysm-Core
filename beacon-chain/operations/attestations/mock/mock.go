// lint:nopanic -- Mock / test code, panic is allowed.
package mock

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/attestations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

var _ attestations.Pool = &PoolMock{}

// PoolMock --
type PoolMock struct {
	AggregatedAtts   []silapb.Att
	UnaggregatedAtts []silapb.Att
}

// AggregateUnaggregatedAttestations --
func (*PoolMock) AggregateUnaggregatedAttestations(_ context.Context) error {
	panic("implement me")
}

// AggregateUnaggregatedAttestationsBySlotIndex --
func (*PoolMock) AggregateUnaggregatedAttestationsBySlotIndex(_ context.Context, _ primitives.Slot, _ primitives.CommitteeIndex) error {
	panic("implement me")
}

// SaveAggregatedAttestation --
func (*PoolMock) SaveAggregatedAttestation(_ silapb.Att) error {
	panic("implement me")
}

// SaveAggregatedAttestations --
func (m *PoolMock) SaveAggregatedAttestations(atts []silapb.Att) error {
	m.AggregatedAtts = append(m.AggregatedAtts, atts...)
	return nil
}

// AggregatedAttestations --
func (m *PoolMock) AggregatedAttestations() []silapb.Att {
	return m.AggregatedAtts
}

// AggregatedAttestationsBySlotIndex --
func (*PoolMock) AggregatedAttestationsBySlotIndex(_ context.Context, _ primitives.Slot, _ primitives.CommitteeIndex) []*silapb.Attestation {
	panic("implement me")
}

// AggregatedAttestationsBySlotIndexElectra --
func (*PoolMock) AggregatedAttestationsBySlotIndexElectra(_ context.Context, _ primitives.Slot, _ primitives.CommitteeIndex) []*silapb.AttestationElectra {
	panic("implement me")
}

// DeleteAggregatedAttestation --
func (*PoolMock) DeleteAggregatedAttestation(_ silapb.Att) error {
	panic("implement me")
}

// HasAggregatedAttestation --
func (*PoolMock) HasAggregatedAttestation(_ silapb.Att) (bool, error) {
	panic("implement me")
}

// AggregatedAttestationCount --
func (*PoolMock) AggregatedAttestationCount() int {
	panic("implement me")
}

// DeleteSeenAggregatedAttestationsBefore --
func (*PoolMock) DeleteSeenAggregatedAttestationsBefore(_ primitives.Slot) {
	panic("implement me")
}

// SeenAggregatedAttestationCount --
func (*PoolMock) SeenAggregatedAttestationCount() int {
	panic("implement me")
}

// SaveUnaggregatedAttestation --
func (*PoolMock) SaveUnaggregatedAttestation(_ silapb.Att) error {
	panic("implement me")
}

// SaveUnaggregatedAttestations --
func (m *PoolMock) SaveUnaggregatedAttestations(atts []silapb.Att) error {
	m.UnaggregatedAtts = append(m.UnaggregatedAtts, atts...)
	return nil
}

// UnaggregatedAttestations --
func (m *PoolMock) UnaggregatedAttestations() []silapb.Att {
	return m.UnaggregatedAtts
}

// UnaggregatedAttestationsBySlotIndex --
func (*PoolMock) UnaggregatedAttestationsBySlotIndex(_ context.Context, _ primitives.Slot, _ primitives.CommitteeIndex) []*silapb.Attestation {
	panic("implement me")
}

// UnaggregatedAttestationsBySlotIndexElectra --
func (*PoolMock) UnaggregatedAttestationsBySlotIndexElectra(_ context.Context, _ primitives.Slot, _ primitives.CommitteeIndex) []*silapb.AttestationElectra {
	panic("implement me")
}

// DeleteUnaggregatedAttestation --
func (*PoolMock) DeleteUnaggregatedAttestation(_ silapb.Att) error {
	panic("implement me")
}

// DeleteSeenUnaggregatedAttestations --
func (*PoolMock) DeleteSeenUnaggregatedAttestations() (int, error) {
	panic("implement me")
}

// UnaggregatedAttestationCount --
func (*PoolMock) UnaggregatedAttestationCount() int {
	panic("implement me")
}

// SaveBlockAttestation --
func (*PoolMock) SaveBlockAttestation(_ silapb.Att) error {
	panic("implement me")
}

// SaveBlockAttestations --
func (*PoolMock) SaveBlockAttestations(_ []silapb.Att) error {
	panic("implement me")
}

// BlockAttestations --
func (*PoolMock) BlockAttestations() []silapb.Att {
	panic("implement me")
}

// DeleteBlockAttestation --
func (*PoolMock) DeleteBlockAttestation(_ silapb.Att) error {
	panic("implement me")
}

// SaveForkchoiceAttestation --
func (*PoolMock) SaveForkchoiceAttestation(_ silapb.Att) error {
	panic("implement me")
}

// SaveForkchoiceAttestations --
func (*PoolMock) SaveForkchoiceAttestations(_ []silapb.Att) error {
	panic("implement me")
}

// ForkchoiceAttestations --
func (*PoolMock) ForkchoiceAttestations() []silapb.Att {
	panic("implement me")
}

// DeleteForkchoiceAttestation --
func (*PoolMock) DeleteForkchoiceAttestation(_ silapb.Att) error {
	panic("implement me")
}

// ForkchoiceAttestationCount --
func (*PoolMock) ForkchoiceAttestationCount() int {
	panic("implement me")
}
