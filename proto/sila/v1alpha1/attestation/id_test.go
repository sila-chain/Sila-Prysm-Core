package attestation_test

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestNewId(t *testing.T) {
	t.Run("full source", func(t *testing.T) {
		att := util.HydrateAttestation(&silapb.Attestation{})
		_, err := attestation.NewId(att, attestation.Full)
		assert.NoError(t, err)
	})
	t.Run("data source Phase 0", func(t *testing.T) {
		att := util.HydrateAttestation(&silapb.Attestation{})
		_, err := attestation.NewId(att, attestation.Data)
		assert.NoError(t, err)
	})
	t.Run("data source Electra", func(t *testing.T) {
		cb := primitives.NewAttestationCommitteeBits()
		cb.SetBitAt(0, true)
		att := util.HydrateAttestationElectra(&silapb.AttestationElectra{CommitteeBits: cb})
		_, err := attestation.NewId(att, attestation.Data)
		assert.NoError(t, err)
	})
	t.Run("ID is different between versions", func(t *testing.T) {
		phase0Att := util.HydrateAttestation(&silapb.Attestation{})
		phase0Id, err := attestation.NewId(phase0Att, attestation.Data)
		require.NoError(t, err)
		cb := primitives.NewAttestationCommitteeBits()
		cb.SetBitAt(0, true) // setting committee bit 0 for Electra corresponds to attestation data's committee index 0 for Phase 0
		electraAtt := util.HydrateAttestationElectra(&silapb.AttestationElectra{CommitteeBits: cb})
		electraId, err := attestation.NewId(electraAtt, attestation.Data)
		require.NoError(t, err)

		assert.NotEqual(t, phase0Id, electraId)
	})
	t.Run("ID is different for different committee bits", func(t *testing.T) {
		cb := primitives.NewAttestationCommitteeBits()
		cb.SetBitAt(0, true)
		cb.SetBitAt(1, true)
		att := util.HydrateAttestationElectra(&silapb.AttestationElectra{CommitteeBits: cb})
		id1, err := attestation.NewId(att, attestation.Data)
		assert.NoError(t, err)
		cb = primitives.NewAttestationCommitteeBits()
		cb.SetBitAt(0, true)
		cb.SetBitAt(2, true)
		att = util.HydrateAttestationElectra(&silapb.AttestationElectra{CommitteeBits: cb})
		id2, err := attestation.NewId(att, attestation.Data)
		assert.NoError(t, err)

		assert.NotEqual(t, id1, id2)
	})
	t.Run("invalid source", func(t *testing.T) {
		att := util.HydrateAttestation(&silapb.Attestation{})
		_, err := attestation.NewId(att, 123)
		assert.ErrorContains(t, "invalid source requested", err)
	})
	t.Run("data source Electra - 0 bits set", func(t *testing.T) {
		cb := primitives.NewAttestationCommitteeBits()
		att := util.HydrateAttestationElectra(&silapb.AttestationElectra{CommitteeBits: cb})
		_, err := attestation.NewId(att, attestation.Data)
		assert.ErrorContains(t, "no committee bits are set", err)
	})
}
