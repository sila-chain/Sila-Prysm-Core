package testing

import (
	"math/rand"
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time"
)

// BitlistWithAllBitsSet creates list of bitlists with all bits set.
func BitlistWithAllBitsSet(length uint64) bitfield.Bitlist {
	b := bitfield.NewBitlist(length)
	for i := range length {
		b.SetBitAt(i, true)
	}
	return b
}

// BitlistsWithSingleBitSet creates list of bitlists with a single bit set in each.
func BitlistsWithSingleBitSet(n, length uint64) []bitfield.Bitlist {
	lists := make([]bitfield.Bitlist, n)
	for i := range n {
		b := bitfield.NewBitlist(length)
		b.SetBitAt(i%length, true)
		lists[i] = b
	}
	return lists
}

// Bitlists64WithSingleBitSet creates list of bitlists with a single bit set in each.
func Bitlists64WithSingleBitSet(n, length uint64) []*bitfield.Bitlist64 {
	lists := make([]*bitfield.Bitlist64, n)
	for i := range n {
		b := bitfield.NewBitlist64(length)
		b.SetBitAt(i%length, true)
		lists[i] = b
	}
	return lists
}

// BitlistsWithMultipleBitSet creates list of bitlists with random n bits set.
func BitlistsWithMultipleBitSet(t testing.TB, n, length, count uint64) []bitfield.Bitlist {
	seed := time.Now().UnixNano()
	t.Logf("bitlistsWithMultipleBitSet random seed: %v", seed)
	r := rand.New(rand.NewSource(seed)) // #nosec G404
	lists := make([]bitfield.Bitlist, n)
	for i := range n {
		b := bitfield.NewBitlist(length)
		keys := r.Perm(int(length)) // lint:ignore uintcast -- This is safe in test code.
		for _, key := range keys[:count] {
			b.SetBitAt(uint64(key), true)
		}
		lists[i] = b
	}
	return lists
}

// Bitlists64WithMultipleBitSet creates list of bitlists with random n bits set.
func Bitlists64WithMultipleBitSet(t testing.TB, n, length, count uint64) []*bitfield.Bitlist64 {
	seed := time.Now().UnixNano()
	t.Logf("Bitlists64WithMultipleBitSet random seed: %v", seed)
	r := rand.New(rand.NewSource(seed)) // #nosec G404
	lists := make([]*bitfield.Bitlist64, n)
	for i := range n {
		b := bitfield.NewBitlist64(length)
		keys := r.Perm(int(length)) // lint:ignore uintcast -- This is safe in test code.
		for _, key := range keys[:count] {
			b.SetBitAt(uint64(key), true)
		}
		lists[i] = b
	}
	return lists
}

// MakeAttestationsFromBitlists creates list of attestations from list of bitlist.
func MakeAttestationsFromBitlists(bl []bitfield.Bitlist) []silapb.Att {
	atts := make([]silapb.Att, len(bl))
	for i, b := range bl {
		atts[i] = &silapb.Attestation{
			AggregationBits: b,
			Data: &silapb.AttestationData{
				Slot:           42,
				CommitteeIndex: 1,
			},
			Signature: bls.NewAggregateSignature().Marshal(),
		}
	}
	return atts
}

// MakeSyncContributionsFromBitVector creates list of sync contributions from list of bitvector.
func MakeSyncContributionsFromBitVector(bl []bitfield.Bitvector128) []*silapb.SyncCommitteeContribution {
	c := make([]*silapb.SyncCommitteeContribution, len(bl))
	for i, b := range bl {
		c[i] = &silapb.SyncCommitteeContribution{
			Slot:              primitives.Slot(1),
			SubcommitteeIndex: 2,
			AggregationBits:   b,
			Signature:         bls.NewAggregateSignature().Marshal(),
		}
	}
	return c
}
