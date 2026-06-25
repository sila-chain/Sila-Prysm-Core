package client

import (
	"testing"

	field_params "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func TestLogSubmittedAtts(t *testing.T) {
	t.Run("phase0 attestations", func(t *testing.T) {
		logHook := logTest.NewGlobal()
		v := validator{
			submittedAtts: make(map[submittedAttKey]*submittedAtt),
		}
		att := util.HydrateAttestation(&silapb.Attestation{})
		att.Data.CommitteeIndex = 12
		require.NoError(t, v.saveSubmittedAtt(att, make([]byte, field_params.BLSPubkeyLength), false))
		v.LogSubmittedAtts(0)
		assert.LogsContain(t, logHook, "committeeIndices=\"[12]\"")
	})
	t.Run("electra attestations", func(t *testing.T) {
		logHook := logTest.NewGlobal()
		v := validator{
			submittedAtts: make(map[submittedAttKey]*submittedAtt),
		}
		att := util.HydrateAttestationElectra(&silapb.AttestationElectra{})
		att.Data.CommitteeIndex = 0
		att.CommitteeBits = primitives.NewAttestationCommitteeBits()
		att.CommitteeBits.SetBitAt(44, true)
		require.NoError(t, v.saveSubmittedAtt(att, make([]byte, field_params.BLSPubkeyLength), false))
		v.LogSubmittedAtts(0)
		assert.LogsContain(t, logHook, "committeeIndices=\"[44]\"")
	})
	t.Run("electra attestations multiple saved", func(t *testing.T) {
		logHook := logTest.NewGlobal()
		v := validator{
			submittedAtts: make(map[submittedAttKey]*submittedAtt),
		}
		att := util.HydrateAttestationElectra(&silapb.AttestationElectra{})
		att.Data.CommitteeIndex = 0
		att.CommitteeBits = primitives.NewAttestationCommitteeBits()
		att.CommitteeBits.SetBitAt(23, true)
		require.NoError(t, v.saveSubmittedAtt(att, make([]byte, field_params.BLSPubkeyLength), false))
		att2 := util.HydrateAttestationElectra(&silapb.AttestationElectra{})
		att2.Data.CommitteeIndex = 0
		att2.CommitteeBits = primitives.NewAttestationCommitteeBits()
		att2.CommitteeBits.SetBitAt(2, true)
		require.NoError(t, v.saveSubmittedAtt(att2, make([]byte, field_params.BLSPubkeyLength), false))
		v.LogSubmittedAtts(0)
		assert.LogsContain(t, logHook, "committeeIndices=\"[23 2]\"")
	})
	t.Run("phase0 aggregates", func(t *testing.T) {
		logHook := logTest.NewGlobal()
		v := validator{
			submittedAggregates: make(map[submittedAttKey]*submittedAtt),
		}
		agg := &silapb.AggregateAttestationAndProof{}
		agg.Aggregate = util.HydrateAttestation(&silapb.Attestation{})
		agg.Aggregate.Data.CommitteeIndex = 12
		require.NoError(t, v.saveSubmittedAtt(agg.AggregateVal(), make([]byte, field_params.BLSPubkeyLength), true))
		v.LogSubmittedAtts(0)
		assert.LogsContain(t, logHook, "committeeIndices=\"[12]\"")
	})
	t.Run("electra aggregates", func(t *testing.T) {
		logHook := logTest.NewGlobal()
		v := validator{
			submittedAggregates: make(map[submittedAttKey]*submittedAtt),
		}
		agg := &silapb.AggregateAttestationAndProofElectra{}
		agg.Aggregate = util.HydrateAttestationElectra(&silapb.AttestationElectra{})
		agg.Aggregate.Data.CommitteeIndex = 0
		agg.Aggregate.CommitteeBits = primitives.NewAttestationCommitteeBits()
		agg.Aggregate.CommitteeBits.SetBitAt(63, true)
		require.NoError(t, v.saveSubmittedAtt(agg.AggregateVal(), make([]byte, field_params.BLSPubkeyLength), true))
		v.LogSubmittedAtts(0)
		assert.LogsContain(t, logHook, "committeeIndices=\"[63]\"")
	})
}
