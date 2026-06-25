package validator

import (
	"fmt"
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	aggtesting "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation/aggregation/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func BenchmarkProposerAtts_sortByProfitability(b *testing.B) {
	bitlistLen := params.BeaconConfig().MaxValidatorsPerCommittee

	tests := []struct {
		name   string
		inputs []bitfield.Bitlist
	}{
		{
			name:   "256 attestations with single bit set",
			inputs: aggtesting.BitlistsWithSingleBitSet(256, bitlistLen),
		},
		{
			name:   "256 attestations with 64 random bits set",
			inputs: aggtesting.BitlistsWithSingleBitSet(256, bitlistLen),
		},
		{
			name:   "512 attestations with single bit set",
			inputs: aggtesting.BitlistsWithSingleBitSet(512, bitlistLen),
		},
		{
			name:   "1024 attestations with 64 random bits set",
			inputs: aggtesting.BitlistsWithMultipleBitSet(b, 1024, bitlistLen, 64),
		},
		{
			name:   "1024 attestations with 512 random bits set",
			inputs: aggtesting.BitlistsWithMultipleBitSet(b, 1024, bitlistLen, 512),
		},
		{
			name:   "1024 attestations with 1000 random bits set",
			inputs: aggtesting.BitlistsWithMultipleBitSet(b, 1024, bitlistLen, 1000),
		},
	}

	runner := func(atts []silapb.Att) {
		attsCopy := make(proposerAtts, len(atts))
		for i, att := range atts {
			attsCopy[i] = att.(*silapb.Attestation).Copy()
		}
		_, err := attsCopy.sort()
		require.NoError(b, err, "Could not sort attestations by profitability")
	}

	for _, tt := range tests {
		b.Run(fmt.Sprintf("max-cover_%s", tt.name), func(b *testing.B) {
			b.StopTimer()
			atts := aggtesting.MakeAttestationsFromBitlists(tt.inputs)
			b.StartTimer()
			for b.Loop() {
				runner(atts)
			}
		})
	}
}
