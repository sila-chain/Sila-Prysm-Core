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

func TestKV_BlockAttestation_CanSaveRetrieve(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b1101}})
	att2 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b1101}})
	att3 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b1101}})
	atts := []silapb.Att{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.SaveBlockAttestation(att))
	}
	// Diff bit length should not panic.
	att4 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b11011}})
	if err := cache.SaveBlockAttestation(att4); err != bitfield.ErrBitlistDifferentLength {
		t.Errorf("Unexpected error: wanted %v, got %v", bitfield.ErrBitlistDifferentLength, err)
	}

	returned := cache.BlockAttestations()

	sort.Slice(returned, func(i, j int) bool {
		return returned[i].GetData().Slot < returned[j].GetData().Slot
	})

	assert.DeepEqual(t, atts, returned)
}

func TestKV_BlockAttestation_CanDelete(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b1101}})
	att2 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b1101}})
	att3 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b1101}})
	atts := []silapb.Att{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.SaveBlockAttestation(att))
	}

	require.NoError(t, cache.DeleteBlockAttestation(att1))
	require.NoError(t, cache.DeleteBlockAttestation(att3))

	returned := cache.BlockAttestations()
	wanted := []silapb.Att{att2}
	assert.DeepEqual(t, wanted, returned)
}
