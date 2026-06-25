package mock

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// PoolMock is a fake implementation of PoolManager.
type PoolMock struct {
	Attestations []*silapb.PayloadAttestation
}

// PendingPayloadAttestations --
func (m *PoolMock) PendingPayloadAttestations(_ primitives.Slot) []*silapb.PayloadAttestation {
	return m.Attestations
}

// InsertPayloadAttestation --
func (m *PoolMock) InsertPayloadAttestation(msg *silapb.PayloadAttestationMessage, _ uint64) error {
	m.Attestations = append(m.Attestations, &silapb.PayloadAttestation{
		Data:      msg.Data,
		Signature: msg.Signature,
	})
	return nil
}

// Seen --
func (*PoolMock) Seen(_ *silapb.PayloadAttestationData, _ uint64) bool {
	return false
}

// MarkIncluded --
func (*PoolMock) MarkIncluded(_ *silapb.PayloadAttestation) {
}
