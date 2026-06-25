package kv

import (
	"bytes"
	"sort"
	"testing"

	"github.com/sila-chain/go-bitfield"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	c "github.com/patrickmn/go-cache"
)

func TestKV_Unaggregated_UnaggregatedAttestations(t *testing.T) {
	t.Run("not returned when hasSeenBit fails", func(t *testing.T) {
		att := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b101}})
		id, err := attestation.NewId(att, attestation.Data)
		require.NoError(t, err)

		cache := NewAttCaches()
		require.NoError(t, cache.SaveUnaggregatedAttestation(att))
		cache.seenAtt.Delete(id.String())
		// cache a bitlist whose length is different from the attestation bitlist's length
		cache.seenAtt.Set(id.String(), []bitfield.Bitlist{{0b1001}}, c.DefaultExpiration)

		atts := cache.UnaggregatedAttestations()
		assert.Equal(t, 0, len(atts))
	})
}

func TestKV_Unaggregated_SaveUnaggregatedAttestation(t *testing.T) {
	tests := []struct {
		name          string
		att           silapb.Att
		count         int
		wantErrString string
	}{
		{
			name: "nil attestation",
			att:  nil,
		},
		{
			name:          "already aggregated",
			att:           &silapb.Attestation{AggregationBits: bitfield.Bitlist{0b10101}, Data: &silapb.AttestationData{Slot: 2}},
			wantErrString: "attestation is aggregated",
		},
		{
			name: "invalid hash",
			att: &silapb.Attestation{
				Data: &silapb.AttestationData{
					BeaconBlockRoot: []byte{0b0},
				},
			},
			wantErrString: "could not create attestation ID",
		},
		{
			name:  "normal save",
			att:   util.HydrateAttestation(&silapb.Attestation{AggregationBits: bitfield.Bitlist{0b0001}}),
			count: 1,
		},
		{
			name: "already seen",
			att: util.HydrateAttestation(&silapb.Attestation{
				Data: &silapb.AttestationData{
					Slot: 100,
				},
				AggregationBits: bitfield.Bitlist{0b10000001},
			}),
			count: 0,
		},
	}
	id, err := attestation.NewId(util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 100}}), attestation.Data)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewAttCaches()
			cache.seenAtt.Set(id.String(), []bitfield.Bitlist{{0xff}}, c.DefaultExpiration)
			assert.Equal(t, 0, len(cache.unAggregatedAtt), "Invalid start pool, atts: %d", len(cache.unAggregatedAtt))

			if tt.att != nil && tt.att.GetSignature() == nil {
				tt.att.(*silapb.Attestation).Signature = make([]byte, fieldparams.BLSSignatureLength)
			}

			err := cache.SaveUnaggregatedAttestation(tt.att)
			if tt.wantErrString != "" {
				assert.ErrorContains(t, tt.wantErrString, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.count, len(cache.unAggregatedAtt), "Wrong attestation count")
			assert.Equal(t, tt.count, cache.UnaggregatedAttestationCount(), "Wrong attestation count")
		})
	}
}

func TestKV_Unaggregated_SaveUnaggregatedAttestations(t *testing.T) {
	tests := []struct {
		name          string
		atts          []silapb.Att
		count         int
		wantErrString string
	}{
		{
			name: "unaggregated only",
			atts: []silapb.Att{
				util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}}),
				util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2}}),
				util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}}),
			},
			count: 3,
		},
		{
			name: "has aggregated",
			atts: []silapb.Att{
				util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}}),
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0b1111}, Data: &silapb.AttestationData{Slot: 2}},
				util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}}),
			},
			wantErrString: "attestation is aggregated",
			count:         1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewAttCaches()
			assert.Equal(t, 0, len(cache.unAggregatedAtt), "Invalid start pool, atts: %d", len(cache.unAggregatedAtt))

			err := cache.SaveUnaggregatedAttestations(tt.atts)
			if tt.wantErrString != "" {
				assert.ErrorContains(t, tt.wantErrString, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.count, len(cache.unAggregatedAtt), "Wrong attestation count")
			assert.Equal(t, tt.count, cache.UnaggregatedAttestationCount(), "Wrong attestation count")
		})
	}
}

func TestKV_Unaggregated_DeleteUnaggregatedAttestation(t *testing.T) {
	t.Run("nil attestation", func(t *testing.T) {
		cache := NewAttCaches()
		assert.NoError(t, cache.DeleteUnaggregatedAttestation(nil))
	})

	t.Run("aggregated attestation", func(t *testing.T) {
		cache := NewAttCaches()
		att := &silapb.Attestation{AggregationBits: bitfield.Bitlist{0b1111}, Data: &silapb.AttestationData{Slot: 2}}
		err := cache.DeleteUnaggregatedAttestation(att)
		assert.ErrorContains(t, "attestation is aggregated", err)
	})

	t.Run("successful deletion", func(t *testing.T) {
		cache := NewAttCaches()
		att1 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b101}})
		att2 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b110}})
		att3 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b110}})
		atts := []silapb.Att{att1, att2, att3}
		require.NoError(t, cache.SaveUnaggregatedAttestations(atts))
		for _, att := range atts {
			assert.NoError(t, cache.DeleteUnaggregatedAttestation(att))
		}
		returned := cache.UnaggregatedAttestations()
		assert.DeepEqual(t, []silapb.Att{}, returned)
	})

	t.Run("deleted when insertSeenBit fails", func(t *testing.T) {
		att := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b101}})
		id, err := attestation.NewId(att, attestation.Data)
		require.NoError(t, err)

		cache := NewAttCaches()
		require.NoError(t, cache.SaveUnaggregatedAttestation(att))
		cache.seenAtt.Delete(id.String())
		// cache a bitlist whose length is different from the attestation bitlist's length
		cache.seenAtt.Set(id.String(), []bitfield.Bitlist{{0b1001}}, c.DefaultExpiration)

		require.NoError(t, cache.DeleteUnaggregatedAttestation(att))
		assert.Equal(t, 0, len(cache.unAggregatedAtt), "Attestation was not deleted")
	})
}

func TestKV_Unaggregated_DeleteSeenUnaggregatedAttestations(t *testing.T) {
	d := util.HydrateAttestationData(&silapb.AttestationData{})

	t.Run("no attestations", func(t *testing.T) {
		cache := NewAttCaches()
		count, err := cache.DeleteSeenUnaggregatedAttestations()
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("none seen", func(t *testing.T) {
		cache := NewAttCaches()
		atts := []silapb.Att{
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1001}}),
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1010}}),
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1100}}),
		}
		require.NoError(t, cache.SaveUnaggregatedAttestations(atts))
		assert.Equal(t, 3, cache.UnaggregatedAttestationCount())

		// As none of attestations have been marked seen, nothing should be deleted.
		count, err := cache.DeleteSeenUnaggregatedAttestations()
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 3, cache.UnaggregatedAttestationCount())
	})

	t.Run("some seen", func(t *testing.T) {
		cache := NewAttCaches()
		atts := []silapb.Att{
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1001}}),
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1010}}),
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1100}}),
		}
		require.NoError(t, cache.SaveUnaggregatedAttestations(atts))
		assert.Equal(t, 3, cache.UnaggregatedAttestationCount())

		require.NoError(t, cache.insertSeenBit(atts[1]))

		// Only seen attestations must be deleted.
		count, err := cache.DeleteSeenUnaggregatedAttestations()
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 2, cache.UnaggregatedAttestationCount())
		returned := cache.UnaggregatedAttestations()
		sort.Slice(returned, func(i, j int) bool {
			return bytes.Compare(returned[i].GetAggregationBits(), returned[j].GetAggregationBits()) < 0
		})
		assert.DeepEqual(t, []silapb.Att{atts[0], atts[2]}, returned)
	})

	t.Run("all seen", func(t *testing.T) {
		cache := NewAttCaches()
		atts := []silapb.Att{
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1001}}),
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1010}}),
			util.HydrateAttestation(&silapb.Attestation{Data: d, AggregationBits: bitfield.Bitlist{0b1100}}),
		}
		require.NoError(t, cache.SaveUnaggregatedAttestations(atts))
		assert.Equal(t, 3, cache.UnaggregatedAttestationCount())

		require.NoError(t, cache.insertSeenBit(atts[0]))
		require.NoError(t, cache.insertSeenBit(atts[1]))
		require.NoError(t, cache.insertSeenBit(atts[2]))

		// All attestations have been processed -- all should be removed.
		count, err := cache.DeleteSeenUnaggregatedAttestations()
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 0, cache.UnaggregatedAttestationCount())
		returned := cache.UnaggregatedAttestations()
		assert.DeepEqual(t, []silapb.Att{}, returned)
	})

	t.Run("deleted when hasSeenBit fails", func(t *testing.T) {
		att := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b101}})
		id, err := attestation.NewId(att, attestation.Data)
		require.NoError(t, err)

		cache := NewAttCaches()
		require.NoError(t, cache.SaveUnaggregatedAttestation(att))
		cache.seenAtt.Delete(id.String())
		// cache a bitlist whose length is different from the attestation bitlist's length
		cache.seenAtt.Set(id.String(), []bitfield.Bitlist{{0b1001}}, c.DefaultExpiration)

		count, err := cache.DeleteSeenUnaggregatedAttestations()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 0, len(cache.unAggregatedAtt), "Attestation was not deleted")
	})
}

func TestKV_Unaggregated_UnaggregatedAttestationsBySlotIndex(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1, CommitteeIndex: 1}, AggregationBits: bitfield.Bitlist{0b101}})
	att2 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1, CommitteeIndex: 2}, AggregationBits: bitfield.Bitlist{0b110}})
	att3 := util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2, CommitteeIndex: 1}, AggregationBits: bitfield.Bitlist{0b110}})
	atts := []*silapb.Attestation{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.SaveUnaggregatedAttestation(att))
	}
	ctx := t.Context()
	returned := cache.UnaggregatedAttestationsBySlotIndex(ctx, 1, 1)
	assert.DeepEqual(t, []*silapb.Attestation{att1}, returned)
	returned = cache.UnaggregatedAttestationsBySlotIndex(ctx, 1, 2)
	assert.DeepEqual(t, []*silapb.Attestation{att2}, returned)
	returned = cache.UnaggregatedAttestationsBySlotIndex(ctx, 2, 1)
	assert.DeepEqual(t, []*silapb.Attestation{att3}, returned)
}

func TestKV_Unaggregated_UnaggregatedAttestationsBySlotIndexElectra(t *testing.T) {
	cache := NewAttCaches()

	committeeBits := primitives.NewAttestationCommitteeBits()
	committeeBits.SetBitAt(1, true)
	att1 := util.HydrateAttestationElectra(&silapb.AttestationElectra{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b101}, CommitteeBits: committeeBits})
	committeeBits = primitives.NewAttestationCommitteeBits()
	committeeBits.SetBitAt(2, true)
	att2 := util.HydrateAttestationElectra(&silapb.AttestationElectra{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b110}, CommitteeBits: committeeBits})
	committeeBits = primitives.NewAttestationCommitteeBits()
	committeeBits.SetBitAt(1, true)
	att3 := util.HydrateAttestationElectra(&silapb.AttestationElectra{Data: &silapb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b110}, CommitteeBits: committeeBits})
	atts := []*silapb.AttestationElectra{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.SaveUnaggregatedAttestation(att))
	}
	ctx := t.Context()
	returned := cache.UnaggregatedAttestationsBySlotIndexElectra(ctx, 1, 1)
	assert.DeepEqual(t, []*silapb.AttestationElectra{att1}, returned)
	returned = cache.UnaggregatedAttestationsBySlotIndexElectra(ctx, 1, 2)
	assert.DeepEqual(t, []*silapb.AttestationElectra{att2}, returned)
	returned = cache.UnaggregatedAttestationsBySlotIndexElectra(ctx, 2, 1)
	assert.DeepEqual(t, []*silapb.AttestationElectra{att3}, returned)
}
