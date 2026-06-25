package attestations_test

import (
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation/aggregation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation/aggregation/attestations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
)

func TestAggregateAttestations_MaxCover_NewMaxCover(t *testing.T) {
	type args struct {
		atts []*silapb.Attestation
	}
	tests := []struct {
		name string
		args args
		want *aggregation.MaxCoverProblem
	}{
		{
			name: "nil attestations",
			args: args{
				atts: nil,
			},
			want: &aggregation.MaxCoverProblem{Candidates: []*aggregation.MaxCoverCandidate{}},
		},
		{
			name: "no attestations",
			args: args{
				atts: []*silapb.Attestation{},
			},
			want: &aggregation.MaxCoverProblem{Candidates: []*aggregation.MaxCoverCandidate{}},
		},
		{
			name: "single attestation",
			args: args{
				atts: []*silapb.Attestation{
					{AggregationBits: bitfield.Bitlist{0b00001010, 0b1}},
				},
			},
			want: &aggregation.MaxCoverProblem{
				Candidates: aggregation.MaxCoverCandidates{
					aggregation.NewMaxCoverCandidate(0, &bitfield.Bitlist{0b00001010, 0b1}),
				},
			},
		},
		{
			name: "multiple attestations",
			args: args{
				atts: []*silapb.Attestation{
					{AggregationBits: bitfield.Bitlist{0b00001010, 0b1}},
					{AggregationBits: bitfield.Bitlist{0b00101010, 0b1}},
					{AggregationBits: bitfield.Bitlist{0b11111010, 0b1}},
					{AggregationBits: bitfield.Bitlist{0b00000010, 0b1}},
					{AggregationBits: bitfield.Bitlist{0b00000001, 0b1}},
				},
			},
			want: &aggregation.MaxCoverProblem{
				Candidates: aggregation.MaxCoverCandidates{
					aggregation.NewMaxCoverCandidate(0, &bitfield.Bitlist{0b00001010, 0b1}),
					aggregation.NewMaxCoverCandidate(1, &bitfield.Bitlist{0b00101010, 0b1}),
					aggregation.NewMaxCoverCandidate(2, &bitfield.Bitlist{0b11111010, 0b1}),
					aggregation.NewMaxCoverCandidate(3, &bitfield.Bitlist{0b00000010, 0b1}),
					aggregation.NewMaxCoverCandidate(4, &bitfield.Bitlist{0b00000001, 0b1}),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, tt.want, attestations.NewMaxCover(tt.args.atts))
		})
	}
}

func TestAggregateAttestations_MaxCover_AttList_validate(t *testing.T) {
	tests := []struct {
		name      string
		atts      attestations.AttList
		wantedErr string
	}{
		{
			name:      "nil list",
			atts:      nil,
			wantedErr: "nil list",
		},
		{
			name:      "empty list",
			atts:      attestations.AttList{},
			wantedErr: "empty list",
		},
		{
			name:      "first bitlist is nil",
			atts:      attestations.AttList{&silapb.Attestation{}},
			wantedErr: "bitlist cannot be nil or empty",
		},
		{
			name: "non first bitlist is nil",
			atts: attestations.AttList{
				&silapb.Attestation{AggregationBits: bitfield.NewBitlist(64)},
				&silapb.Attestation{},
			},
			wantedErr: "bitlist cannot be nil or empty",
		},
		{
			name: "first bitlist is empty",
			atts: attestations.AttList{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{}},
			},
			wantedErr: "bitlist cannot be nil or empty",
		},
		{
			name: "non first bitlist is empty",
			atts: attestations.AttList{
				&silapb.Attestation{AggregationBits: bitfield.NewBitlist(64)},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{}},
			},
			wantedErr: "bitlist cannot be nil or empty",
		},
		{
			name: "valid bitlists",
			atts: attestations.AttList{
				&silapb.Attestation{AggregationBits: bitfield.NewBitlist(64)},
				&silapb.Attestation{AggregationBits: bitfield.NewBitlist(64)},
				&silapb.Attestation{AggregationBits: bitfield.NewBitlist(64)},
				&silapb.Attestation{AggregationBits: bitfield.NewBitlist(64)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.atts.ValidateForTesting()
			if tt.wantedErr != "" {
				assert.ErrorContains(t, tt.wantedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAggregateAttestations_rearrangeProcessedAttestations(t *testing.T) {
	tests := []struct {
		name     string
		atts     []silapb.Att
		keys     []int
		wantAtts []silapb.Att
	}{
		{
			name: "nil attestations",
		},
		{
			name: "single attestation no processed keys",
			atts: []silapb.Att{
				&silapb.Attestation{},
			},
			wantAtts: []silapb.Att{
				&silapb.Attestation{},
			},
		},
		{
			name: "single attestation processed",
			atts: []silapb.Att{
				&silapb.Attestation{},
			},
			keys: []int{0},
			wantAtts: []silapb.Att{
				nil,
			},
		},
		{
			name: "multiple processed, last attestation marked",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x02}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
			},
			keys: []int{1, 4}, // Only attestation at index 1, should be moved, att at 4 is already at the end.
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x02}},
				nil, nil,
			},
		},
		{
			name: "all processed",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x02}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
			},
			keys: []int{0, 1, 2, 3, 4},
			wantAtts: []silapb.Att{
				nil, nil, nil, nil, nil,
			},
		},
		{
			name: "operate on slice, single attestation marked",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x02}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
				// Assuming some attestations have been already marked as nil, during previous rounds:
				nil, nil, nil,
			},
			keys: []int{2},
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				nil, nil, nil, nil,
			},
		},
		{
			name: "operate on slice, non-last attestation marked",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x02}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x05}},
				// Assuming some attestations have been already marked as nil, during previous rounds:
				nil, nil, nil,
			},
			keys: []int{2, 3},
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x05}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
				nil, nil, nil, nil, nil,
			},
		},
		{
			name: "operate on slice, last attestation marked",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x02}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
				// Assuming some attestations have been already marked as nil, during previous rounds:
				nil, nil, nil,
			},
			keys: []int{2, 4},
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				nil, nil, nil, nil, nil,
			},
		},
		{
			name: "many items, many selected, keys unsorted",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x02}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x04}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x05}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x06}},
			},
			keys: []int{4, 1, 2, 5, 6},
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x03}},
				nil, nil, nil, nil, nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := make([]*bitfield.Bitlist64, len(tt.atts))
			for i := 0; i < len(tt.atts); i++ {
				if tt.atts[i] != nil {
					var err error
					candidates[i], err = tt.atts[i].GetAggregationBits().ToBitlist64()
					if err != nil {
						t.Error(err)
					}
				}
			}
			attestations.RearrangeProcessedAttestations(tt.atts, candidates, tt.keys)
			assert.DeepEqual(t, tt.atts, tt.wantAtts)
		})
	}
}

func TestAggregateAttestations_aggregateAttestations(t *testing.T) {
	sign := bls.NewAggregateSignature().Marshal()
	tests := []struct {
		name          string
		atts          []silapb.Att
		wantAtts      []silapb.Att
		keys          []int
		coverage      *bitfield.Bitlist64
		wantTargetIdx int
		wantErr       string
	}{
		{
			name:          "nil attestation",
			wantTargetIdx: 0,
			wantErr:       attestations.ErrInvalidAttestationCount.Error(),
			keys:          []int{0, 1, 2},
		},
		{
			name: "single attestation",
			atts: []silapb.Att{
				&silapb.Attestation{},
			},
			wantTargetIdx: 0,
			wantErr:       attestations.ErrInvalidAttestationCount.Error(),
			keys:          []int{0, 1, 2},
		},
		{
			name:          "no keys",
			wantTargetIdx: 0,
			wantErr:       attestations.ErrInvalidAttestationCount.Error(),
		},
		{
			name: "two attestations, none selected",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
			},
			wantTargetIdx: 0,
			wantErr:       attestations.ErrInvalidAttestationCount.Error(),
			keys:          []int{},
		},
		{
			name: "two attestations, one selected",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x00}},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x01}},
			},
			wantTargetIdx: 0,
			wantErr:       attestations.ErrInvalidAttestationCount.Error(),
			keys:          []int{0},
		},
		{
			name: "two attestations, both selected, empty coverage",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0b00000001, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0b00000110, 0b1}, Signature: sign},
			},
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0b00000111, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0b00000110, 0b1}, Signature: sign},
			},
			wantTargetIdx: 0,
			wantErr:       "invalid or empty coverage",
			keys:          []int{0, 1},
		},
		{
			name: "two attestations, both selected",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000001, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000010, 0b1}, Signature: sign},
			},
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000011, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000010, 0b1}, Signature: sign},
			},
			wantTargetIdx: 0,
			keys:          []int{0, 1},
			coverage: func() *bitfield.Bitlist64 {
				b, err := bitfield.NewBitlist64FromBytes(64, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000011})
				if err != nil {
					t.Fatal(err)
				}
				return b
			}(),
		},
		{
			name: "many attestations, several selected",
			atts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000001, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000010, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000100, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00001000, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00010000, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00100000, 0b1}, Signature: sign},
			},
			wantAtts: []silapb.Att{
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000001, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00010110, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00000100, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00001000, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00010000, 0b1}, Signature: sign},
				&silapb.Attestation{AggregationBits: bitfield.Bitlist{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00100000, 0b1}, Signature: sign},
			},
			wantTargetIdx: 1,
			keys:          []int{1, 2, 4},
			coverage: func() *bitfield.Bitlist64 {
				b, err := bitfield.NewBitlist64FromBytes(64, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0b00010110})
				if err != nil {
					t.Fatal(err)
				}
				return b
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTargetIdx, err := attestations.AggregateAttestations(tt.atts, tt.keys, tt.coverage)
			if tt.wantErr != "" {
				assert.ErrorContains(t, tt.wantErr, err)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantTargetIdx, gotTargetIdx)
			extractBitlists := func(atts []silapb.Att) []bitfield.Bitlist {
				bl := make([]bitfield.Bitlist, len(atts))
				for i, att := range atts {
					bl[i] = att.GetAggregationBits()
				}
				return bl
			}
			assert.DeepEqual(t, extractBitlists(tt.atts), extractBitlists(tt.wantAtts))
		})
	}
}
