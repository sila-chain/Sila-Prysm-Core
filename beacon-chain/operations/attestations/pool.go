package attestations

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/attestations/kv"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// Pool defines the necessary methods for Sila attestations pool to serve
// fork choice and validators. In the current design, aggregated attestations
// are used by proposer actor. Unaggregated attestations are used by
// aggregator actor.
type Pool interface {
	// For Aggregated attestations
	AggregateUnaggregatedAttestations(ctx context.Context) error
	SaveAggregatedAttestation(att silapb.Att) error
	SaveAggregatedAttestations(atts []silapb.Att) error
	AggregatedAttestations() []silapb.Att
	AggregatedAttestationsBySlotIndex(ctx context.Context, slot primitives.Slot, committeeIndex primitives.CommitteeIndex) []*silapb.Attestation
	AggregatedAttestationsBySlotIndexElectra(ctx context.Context, slot primitives.Slot, committeeIndex primitives.CommitteeIndex) []*silapb.AttestationElectra
	DeleteAggregatedAttestation(att silapb.Att) error
	HasAggregatedAttestation(att silapb.Att) (bool, error)
	AggregatedAttestationCount() int
	// Seen aggregated attestations.
	DeleteSeenAggregatedAttestationsBefore(expirySlot primitives.Slot)
	SeenAggregatedAttestationCount() int
	// For unaggregated attestations.
	SaveUnaggregatedAttestation(att silapb.Att) error
	SaveUnaggregatedAttestations(atts []silapb.Att) error
	UnaggregatedAttestations() []silapb.Att
	UnaggregatedAttestationsBySlotIndex(ctx context.Context, slot primitives.Slot, committeeIndex primitives.CommitteeIndex) []*silapb.Attestation
	UnaggregatedAttestationsBySlotIndexElectra(ctx context.Context, slot primitives.Slot, committeeIndex primitives.CommitteeIndex) []*silapb.AttestationElectra
	DeleteUnaggregatedAttestation(att silapb.Att) error
	DeleteSeenUnaggregatedAttestations() (int, error)
	UnaggregatedAttestationCount() int
	// For attestations that were included in the block.
	SaveBlockAttestation(att silapb.Att) error
	BlockAttestations() []silapb.Att
	DeleteBlockAttestation(att silapb.Att) error
	// For attestations to be passed to fork choice.
	SaveForkchoiceAttestations(atts []silapb.Att) error
	ForkchoiceAttestations() []silapb.Att
	DeleteForkchoiceAttestation(att silapb.Att) error
	ForkchoiceAttestationCount() int
}

// NewPool initializes a new attestation pool.
func NewPool() *kv.AttCaches {
	return kv.NewAttCaches()
}
