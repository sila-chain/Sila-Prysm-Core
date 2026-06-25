package cache

import (
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls/blst"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestAdd(t *testing.T) {
	k, err := blst.RandKey()
	require.NoError(t, err)
	sig := k.Sign([]byte{'X'})

	t.Run("new ID", func(t *testing.T) {
		t.Run("first ID ever", func(t *testing.T) {
			c := NewAttestationCache()
			ab := bitfield.NewBitlist(8)
			ab.SetBitAt(0, true)
			att := &silapb.Attestation{
				Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
				AggregationBits: ab,
				Signature:       sig.Marshal(),
			}
			id, err := attestation.NewId(att, attestation.Data)
			require.NoError(t, err)
			require.NoError(t, c.Add(att))

			require.Equal(t, 1, len(c.atts))
			group, ok := c.atts[id]
			require.Equal(t, true, ok)
			assert.Equal(t, primitives.Slot(123), group.slot)
			require.Equal(t, 1, len(group.atts))
			assert.DeepEqual(t, group.atts[0], att)
		})
		t.Run("other ID exists", func(t *testing.T) {
			c := NewAttestationCache()
			ab := bitfield.NewBitlist(8)
			ab.SetBitAt(0, true)
			existingAtt := &silapb.Attestation{
				Data:            &silapb.AttestationData{BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
				AggregationBits: ab,
				Signature:       sig.Marshal(),
			}
			existingId, err := attestation.NewId(existingAtt, attestation.Data)
			require.NoError(t, err)
			c.atts[existingId] = &attGroup{slot: existingAtt.Data.Slot, atts: []silapb.Att{existingAtt}}

			att := &silapb.Attestation{
				Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
				AggregationBits: ab,
				Signature:       sig.Marshal(),
			}
			id, err := attestation.NewId(att, attestation.Data)
			require.NoError(t, err)
			require.NoError(t, c.Add(att))

			require.Equal(t, 2, len(c.atts))
			group, ok := c.atts[id]
			require.Equal(t, true, ok)
			assert.Equal(t, primitives.Slot(123), group.slot)
			require.Equal(t, 1, len(group.atts))
			assert.DeepEqual(t, group.atts[0], att)
		})
	})
	t.Run("aggregated", func(t *testing.T) {
		c := NewAttestationCache()
		existingAtt := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		id, err := attestation.NewId(existingAtt, attestation.Data)
		require.NoError(t, err)
		c.atts[id] = &attGroup{slot: existingAtt.Data.Slot, atts: []silapb.Att{existingAtt}}

		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(0, true)
		att.AggregationBits.SetBitAt(1, true)
		require.NoError(t, c.Add(att))

		require.Equal(t, 1, len(c.atts))
		group, ok := c.atts[id]
		require.Equal(t, true, ok)
		assert.Equal(t, primitives.Slot(123), group.slot)
		require.Equal(t, 2, len(group.atts))
		assert.DeepEqual(t, group.atts[0], existingAtt)
		assert.DeepEqual(t, group.atts[1], att)
	})
	t.Run("unaggregated - existing bit", func(t *testing.T) {
		c := NewAttestationCache()
		existingAtt := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		existingAtt.AggregationBits.SetBitAt(0, true)
		id, err := attestation.NewId(existingAtt, attestation.Data)
		require.NoError(t, err)
		c.atts[id] = &attGroup{slot: existingAtt.Data.Slot, atts: []silapb.Att{existingAtt}}

		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(0, true)
		require.NoError(t, c.Add(att))

		require.Equal(t, 1, len(c.atts))
		group, ok := c.atts[id]
		require.Equal(t, true, ok)
		assert.Equal(t, primitives.Slot(123), group.slot)
		require.Equal(t, 1, len(group.atts))
		assert.DeepEqual(t, []int{0}, group.atts[0].GetAggregationBits().BitIndices())
	})
	t.Run("unaggregated - new bit", func(t *testing.T) {
		c := NewAttestationCache()
		existingAtt := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		existingAtt.AggregationBits.SetBitAt(0, true)
		id, err := attestation.NewId(existingAtt, attestation.Data)
		require.NoError(t, err)
		c.atts[id] = &attGroup{slot: existingAtt.Data.Slot, atts: []silapb.Att{existingAtt}}

		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(1, true)
		require.NoError(t, c.Add(att))

		require.Equal(t, 1, len(c.atts))
		group, ok := c.atts[id]
		require.Equal(t, true, ok)
		assert.Equal(t, primitives.Slot(123), group.slot)
		require.Equal(t, 1, len(group.atts))
		assert.DeepEqual(t, []int{0, 1}, group.atts[0].GetAggregationBits().BitIndices())
	})
}

func TestGetAll(t *testing.T) {
	c := NewAttestationCache()
	c.atts[bytesutil.ToBytes32([]byte("id1"))] = &attGroup{atts: []silapb.Att{&silapb.Attestation{}, &silapb.Attestation{}}}
	c.atts[bytesutil.ToBytes32([]byte("id2"))] = &attGroup{atts: []silapb.Att{&silapb.Attestation{}}}

	assert.Equal(t, 3, len(c.GetAll()))
}

func TestCount(t *testing.T) {
	c := NewAttestationCache()
	c.atts[bytesutil.ToBytes32([]byte("id1"))] = &attGroup{atts: []silapb.Att{&silapb.Attestation{}, &silapb.Attestation{}}}
	c.atts[bytesutil.ToBytes32([]byte("id2"))] = &attGroup{atts: []silapb.Att{&silapb.Attestation{}}}

	assert.Equal(t, 3, c.Count())
}

func TestDeleteCovered(t *testing.T) {
	k, err := blst.RandKey()
	require.NoError(t, err)
	sig := k.Sign([]byte{'X'})

	att1 := &silapb.Attestation{
		Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
		AggregationBits: bitfield.NewBitlist(8),
		Signature:       sig.Marshal(),
	}
	att1.AggregationBits.SetBitAt(0, true)

	att2 := &silapb.Attestation{
		Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
		AggregationBits: bitfield.NewBitlist(8),
		Signature:       sig.Marshal(),
	}
	att2.AggregationBits.SetBitAt(1, true)
	att2.AggregationBits.SetBitAt(2, true)

	att3 := &silapb.Attestation{
		Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
		AggregationBits: bitfield.NewBitlist(8),
		Signature:       sig.Marshal(),
	}
	att3.AggregationBits.SetBitAt(1, true)
	att3.AggregationBits.SetBitAt(3, true)
	att3.AggregationBits.SetBitAt(4, true)

	c := NewAttestationCache()
	id, err := attestation.NewId(att1, attestation.Data)
	require.NoError(t, err)
	c.atts[id] = &attGroup{slot: att1.Data.Slot, atts: []silapb.Att{att1, att2, att3}}

	t.Run("no matching group", func(t *testing.T) {
		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 456, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(0, true)
		att.AggregationBits.SetBitAt(1, true)
		att.AggregationBits.SetBitAt(2, true)
		att.AggregationBits.SetBitAt(3, true)
		att.AggregationBits.SetBitAt(4, true)
		require.NoError(t, c.DeleteCovered(att))

		assert.Equal(t, 3, len(c.atts[id].atts))
	})
	t.Run("covered atts deleted", func(t *testing.T) {
		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(0, true)
		att.AggregationBits.SetBitAt(1, true)
		att.AggregationBits.SetBitAt(3, true)
		att.AggregationBits.SetBitAt(4, true)
		require.NoError(t, c.DeleteCovered(att))

		atts := c.atts[id].atts
		require.Equal(t, 1, len(atts))
		assert.DeepEqual(t, att2, atts[0])
	})
	t.Run("last att in group deleted", func(t *testing.T) {
		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(0, true)
		att.AggregationBits.SetBitAt(1, true)
		att.AggregationBits.SetBitAt(2, true)
		att.AggregationBits.SetBitAt(3, true)
		att.AggregationBits.SetBitAt(4, true)
		require.NoError(t, c.DeleteCovered(att))

		assert.Equal(t, 0, len(c.atts))
	})
}

func TestPruneBefore(t *testing.T) {
	c := NewAttestationCache()
	c.atts[bytesutil.ToBytes32([]byte("id1"))] = &attGroup{slot: 1, atts: []silapb.Att{&silapb.Attestation{}, &silapb.Attestation{}}}
	c.atts[bytesutil.ToBytes32([]byte("id2"))] = &attGroup{slot: 3, atts: []silapb.Att{&silapb.Attestation{}}}
	c.atts[bytesutil.ToBytes32([]byte("id3"))] = &attGroup{slot: 2, atts: []silapb.Att{&silapb.Attestation{}}}

	count := c.PruneBefore(3)

	require.Equal(t, 1, len(c.atts))
	_, ok := c.atts[bytesutil.ToBytes32([]byte("id2"))]
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(3), count)
}

func TestAggregateIsRedundant(t *testing.T) {
	k, err := blst.RandKey()
	require.NoError(t, err)
	sig := k.Sign([]byte{'X'})

	c := NewAttestationCache()
	existingAtt := &silapb.Attestation{
		Data:            &silapb.AttestationData{Slot: 123, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
		AggregationBits: bitfield.NewBitlist(8),
		Signature:       sig.Marshal(),
	}
	existingAtt.AggregationBits.SetBitAt(0, true)
	existingAtt.AggregationBits.SetBitAt(1, true)
	id, err := attestation.NewId(existingAtt, attestation.Data)
	require.NoError(t, err)
	c.atts[id] = &attGroup{slot: existingAtt.Data.Slot, atts: []silapb.Att{existingAtt}}

	t.Run("no matching group", func(t *testing.T) {
		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: 456, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(0, true)

		redundant, err := c.AggregateIsRedundant(att)
		require.NoError(t, err)
		assert.Equal(t, false, redundant)
	})
	t.Run("redundant", func(t *testing.T) {
		att := &silapb.Attestation{
			Data:            &silapb.AttestationData{Slot: existingAtt.Data.Slot, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
			AggregationBits: bitfield.NewBitlist(8),
			Signature:       sig.Marshal(),
		}
		att.AggregationBits.SetBitAt(0, true)

		redundant, err := c.AggregateIsRedundant(att)
		require.NoError(t, err)
		assert.Equal(t, true, redundant)
	})
	t.Run("not redundant", func(t *testing.T) {
		t.Run("strictly better", func(t *testing.T) {
			att := &silapb.Attestation{
				Data:            &silapb.AttestationData{Slot: existingAtt.Data.Slot, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
				AggregationBits: bitfield.NewBitlist(8),
				Signature:       sig.Marshal(),
			}
			att.AggregationBits.SetBitAt(0, true)
			att.AggregationBits.SetBitAt(1, true)
			att.AggregationBits.SetBitAt(2, true)

			redundant, err := c.AggregateIsRedundant(att)
			require.NoError(t, err)
			assert.Equal(t, false, redundant)
		})
		t.Run("overlapping and new bits", func(t *testing.T) {
			att := &silapb.Attestation{
				Data:            &silapb.AttestationData{Slot: existingAtt.Data.Slot, BeaconBlockRoot: make([]byte, 32), Source: &silapb.Checkpoint{Root: make([]byte, 32)}, Target: &silapb.Checkpoint{Root: make([]byte, 32)}},
				AggregationBits: bitfield.NewBitlist(8),
				Signature:       sig.Marshal(),
			}
			att.AggregationBits.SetBitAt(0, true)
			att.AggregationBits.SetBitAt(2, true)

			redundant, err := c.AggregateIsRedundant(att)
			require.NoError(t, err)
			assert.Equal(t, false, redundant)
		})
	})
}

func TestGetBySlotAndCommitteeIndex(t *testing.T) {
	c := NewAttestationCache()
	c.atts[bytesutil.ToBytes32([]byte("id1"))] = &attGroup{slot: 1, atts: []silapb.Att{&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1, CommitteeIndex: 1}}, &silapb.Attestation{Data: &silapb.AttestationData{Slot: 1, CommitteeIndex: 1}}}}
	c.atts[bytesutil.ToBytes32([]byte("id2"))] = &attGroup{slot: 2, atts: []silapb.Att{&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2, CommitteeIndex: 2}}}}
	c.atts[bytesutil.ToBytes32([]byte("id3"))] = &attGroup{slot: 1, atts: []silapb.Att{&silapb.Attestation{Data: &silapb.AttestationData{Slot: 2, CommitteeIndex: 2}}}}

	// committeeIndex has to be small enough to fit in the bitvector
	atts := GetBySlotAndCommitteeIndex[*silapb.Attestation](c, 1, 1)
	require.Equal(t, 2, len(atts))
	assert.Equal(t, primitives.Slot(1), atts[0].Data.Slot)
	assert.Equal(t, primitives.Slot(1), atts[1].Data.Slot)
	assert.Equal(t, primitives.CommitteeIndex(1), atts[0].Data.CommitteeIndex)
	assert.Equal(t, primitives.CommitteeIndex(1), atts[1].Data.CommitteeIndex)
}
