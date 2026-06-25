package kv

import (
	"sort"
	"testing"

	"github.com/sila-chain/go-bitfield"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestKV_Forkchoice_CanSaveRetrieve(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b1101}})
	att2 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b1101}})
	att3 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b1101}})
	atts := []silapb.Att{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.saveForkchoiceAttestation(att))
	}

	returned := cache.ForkchoiceAttestations()

	sort.Slice(returned, func(i, j int) bool {
		return returned[i].GetData().Slot < returned[j].GetData().Slot
	})

	assert.DeepEqual(t, atts, returned)
}

func TestKV_Forkchoice_CanDelete(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b1101}})
	att2 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b1101}})
	att3 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b1101}})
	atts := []silapb.Att{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.saveForkchoiceAttestation(att))
	}

	require.NoError(t, cache.DeleteForkchoiceAttestation(att1))
	require.NoError(t, cache.DeleteForkchoiceAttestation(att3))

	returned := cache.ForkchoiceAttestations()
	wanted := []silapb.Att{att2}
	assert.DeepEqual(t, wanted, returned)
}

func TestKV_Forkchoice_CanCount(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b1101}})
	att2 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b1101}})
	att3 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b1101}})
	atts := []*silapb.Attestation{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.saveForkchoiceAttestation(att))
	}

	require.Equal(t, 3, cache.ForkchoiceAttestationCount())
}
